// Package stusb4500 provides a driver for the STUSB4500 USB PD sink controller
// by STMicroelectronics.
//
// Datasheet: https://www.st.com/resource/en/datasheet/stusb4500.pdf
package stusb4500 // import "tinygo.org/x/drivers/stusb4500"

import (
	"errors"
	"machine"
	"runtime/volatile"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/stusb4500/conf"
)

var (
	ErrDeviceNotConfigured = errors.New("stusb4500: device not configured")
	ErrDeviceNotFound      = errors.New("stusb4500: device not found")
	ErrInvalidNumPDO       = errors.New("stusb4500: invalid number of PDO")
	ErrUndefinedRDO        = errors.New("stusb4500: undefined RDO")
	ErrCableDisconnected   = errors.New("stusb4500: cable is disconnected")
	ErrByteCountPDHeader   = errors.New("stusb4500: invalid byte count in USB PD message header")
	ErrSourcePDOTimeout    = errors.New("stusb4500: timeout requesting source capabilities")
	ErrMonitorStarted      = errors.New("stusb4500: monitor already started")
	ErrMonitorNotStarted   = errors.New("stusb4500: monitor not yet started")
	ErrMonitorResetUndef   = errors.New("stusb4500: reset pin not defined for monitor")
	ErrMonitorAlertUndef   = errors.New("stusb4500: alert pin not defined for monitor")
	ErrMonitorAttachUndef  = errors.New("stusb4500: attach pin not defined for monitor")
)

// Address is the default I2C peripheral address of the STUSB4500 used when
// creating a new connection. Use the GetAddress method to get the peripheral
// address of a connected Device.
var Address uint8 = 0x28

// CableStatus represents the connection status of a USB Type-C cable.
type CableStatus uint8

const (
	Disconnected CableStatus = 0 // no cable attached to the USB Type-C port
	ConnectedCC1 CableStatus = 1 // cable attached (normal/unflipped orientation)
	ConnectedCC2 CableStatus = 2 // cable attached (twisted/flipped orientation)
	CableInvalid CableStatus = 0xFF
)

const (
	// negotiationDuration defines the maximum time required for the STUSB4500 to
	// negotiate a USB PD explicit contract after connecting to a source.
	//
	// After this amount of time has elapsed, the device will have initialized its
	// registers, finished loading from NVM, established a power setting with the
	// source, and is able to respond over the I2C bus.
	//
	// However, the negotiated power setting may not yet be stabilized. The most
	// reliable way to detect this condition is by measuring VBUS externally.
	//
	//   | At the connection, the STUSB4500 connects first in USB-C mode before
	//   | negotiating any USB PD contract. In order to know the final connection
	//   | status (USB-C or USB PD explicit contract), it is recommended to wait
	//   | 500 ms after ATTACH event.
	//   + -- The STUSB4500 software programming guide (UM2650), Ch 1.7, Pg 4
	negotiationDuration = 500 * time.Millisecond

	hardResetDuration     = 250 * time.Millisecond
	softResetDuration     = 250 * time.Millisecond
	attachBounceDuration  = 27 * time.Millisecond
	reconnectWaitDuration = 5 * time.Millisecond
)

// maximum number of sink PDOs
const pdoSnkMax = 3

const (
	pdoVoltageUnits = 50 // step size of PDO voltage (50 mV)
	pdoCurrentUnits = 10 // step size of PDO current (10 mA)
)

// Device wraps the I2C connection to an STUSB4500 device.
type Device struct {
	bus     drivers.I2C
	address uint8
	config  conf.Configuration

	ready   bool        // whether or not the device has been configured
	monitor monitorLock // whether or not the monitor is running

	// most recent PDOs advertised from source.
	SnkRDO PDO
	SnkPDO []PDO
	SrcPDO []PDO

	status deviceStatus
	fsm    deviceState
}

// device register info kept in application memory
type deviceStatus struct {
	reset    bool
	hwFault  regStatusHWFault  // 0x13
	typeCMon regStatusTypeCMon // 0x10
	ccDetect regStatusCCDetect // 0x0E
	cc       regStatusCC       // 0x11
	prt      regStatusPRT      // 0x16
	phy      regStatusPHY      // 0x17
	rdo      regStatusRDO      // 0x91-0x94
	snk      []regStatusSnkPDO // 0x85-0x90
	src      []regStatusSrcPDO
}

// USB PD state machine
type deviceState struct {
	alert     volatile.Register8
	attach    volatile.Register8
	request   volatile.Register8
	receive   volatile.Register8
	cable     CableStatus
	irqRecv   uint16
	irqReset  uint16
	trans     uint16
	psrdyRecv uint16
	msgRecv   uint16
	msgAccept uint16
	msgReject uint16
	msgCRC    uint16
	peState   uint8
}

// New creates a new STUSB4500 connection. The given I2C interface must already
// be configured.
func New(bus drivers.I2C) *Device {
	return &Device{
		bus:     bus,
		address: Address,
		fsm: deviceState{
			cable: CableInvalid,
		},
	}
}

// GetAddress returns the I2C peripheral address of a connected Device.
func (d *Device) GetAddress() uint8 {
	return d.address
}

// Configure modifies the device configuration settings and must be called after
// a Device is created via New.
//
// Note that the given pins must already be configured as follows:
//   config.ResetPin  - GPIO output
//   config.AlertPin  - GPIO input pullup (must be external interrupt capable)
//   config.AttachPin - GPIO input pullup (must be external interrupt capable)
func (d *Device) Configure(config conf.Configuration) *Device {
	d.config = config
	if d.config.ResetPin != machine.NoPin {
		d.config.ResetPin.Low() // RST is active high, so drive low before init
	}
	if d.config.AlertPin != machine.NoPin {
		d.config.AlertPin.SetInterrupt(machine.PinFalling, d.onAlert)
	}
	if d.config.AttachPin != machine.NoPin {
		d.config.AttachPin.SetInterrupt(machine.PinFalling, d.onAttach)
	}
	if 0 == d.config.USBPDTimeout {
		d.config.USBPDTimeout = conf.DefaultUSBPDTimeout
	}
	d.ready = true
	return d
}

// Initialize clears and unmasks all needed interrupts, and initializes device
// registers and USB PD state machine.
// A non-nil error is returned if communication with device was unsuccessful, or
// the device was unable to negotiate with a USB PD-capable supply.
func (d *Device) Initialize() error {

	// clear the lists of PDOs
	d.status.snk = []regStatusSnkPDO{}
	d.SnkPDO = []PDO{}
	d.status.rdo = regStatusRDO{}
	d.SnkRDO = PDO{}
	d.status.src = []regStatusSrcPDO{}
	d.SrcPDO = []PDO{}

	// ensure Configure has been called to register interrupt handlers
	if !d.ready {
		return ErrDeviceNotConfigured
	}
	// ensure enough time has elapsed for NVM loading and PD negotiation
	time.Sleep(negotiationDuration)

	// check that we are communicating with the device
	found := false
	for i := 0; !found && i < d.config.USBPDTimeout; i++ {
		if found = d.Connected(); !found {
			time.Sleep(reconnectWaitDuration)
		}
	}
	if !found {
		return ErrDeviceNotFound
	}

	if err := d.enableAlerts(); nil != err {
		return err
	}
	if err := d.updateStatus(); nil != err {
		return err
	}
	if err := d.setNumSnkPDO(1); nil != err {
		return err
	}
	if err := d.updateSnkPDO(); nil != err {
		return err
	}
	if err := d.updateSnkRDO(); nil != err {
		return err
	}
	if connected, _ := d.GetCableStatus(); connected {
		if err := d.updateSrcPDO(); nil != err {
			return err
		}
	}
	return nil
}

// Connected returns true if and only if we are connected to an STUSB4500.
// A valid connection is determined by reading the device ID register over I2C.
func (d *Device) Connected() bool {
	id, err := d.readRegister(REG_DEVICE_ID)
	if nil != err {
		return false
	}
	return isDeviceIDValid(id)
}

// GetCableStatus returns whether a USB cable is connected and the orientation
// of its plug to the STUSB4500 Type-C port.
func (d *Device) GetCableStatus() (bool, CableStatus) {
	// first verify a cable is attached
	cc, err := d.readRegister(REG_PORT_STATUS_1)
	if nil != err {
		return false, CableInvalid
	}
	var ccDetect regStatusCCDetect
	ccDetect.parse(cc)
	if !ccDetect.attached {
		return false, Disconnected
	}
	// next check which CC1/CC2 is connected through to a source
	tc, err := d.readRegister(REG_TYPEC_STATUS)
	if nil != err {
		return false, CableInvalid
	}
	var typeC regStatusTypeC
	typeC.parse(tc)
	if typeC.reverse {
		return true, ConnectedCC2
	}
	return true, ConnectedCC1
}

// Update processes interrupts and internal alerts, and progresses the PD state
// machine by one cycle. The user application must call Update continuously for
// successful operation of the STUSB4500.
func (d *Device) Update() error {

	if pe, err := d.readRegister(REG_PE_FSM); nil != err {
		return err
	} else {
		d.fsm.peState = pe
	}

	// if alert line is low, ensure we run through the ISR at least once
	if !d.config.AlertPin.Get() {
		d.fsm.alert.Set(d.fsm.alert.Get() + 1)
	}

	for {
		if d.fsm.alert.Get() > 0 {
			d.fsm.alert.Set(d.fsm.alert.Get() - 1)
			if err := d.processAlerts(); nil != err {
				return err
			}
		} else {
			break // no alerts remaining
		}
	}

	for {
		if d.fsm.attach.Get() > 0 {
			d.fsm.attach.Set(d.fsm.attach.Get() - 1)
			connected, cable := d.GetCableStatus()
			// ignore duplicate interrupts, only handle connection state changes
			if d.fsm.cable != cable {
				if connected {
					time.Sleep(attachBounceDuration)
					if nil != d.config.OnCableAttach {
						d.config.OnCableAttach()
					}
				} else {
					if Disconnected == cable {
						if nil != d.config.OnCableDetach {
							d.config.OnCableDetach()
						}
					}
					// clear source PDO list
					d.status.src = []regStatusSrcPDO{}
					d.SrcPDO = []PDO{}
				}
				d.fsm.cable = cable
			}
		} else {
			break // no attach events remaining
		}
	}

	if d.fsm.receive.Get() > 0 {
		d.fsm.receive.Set(0) // only accept a single source capabilities event
		if nil != d.config.OnCapabilities {
			d.config.OnCapabilities()
		}
	}

	return nil
}

// SoftReset performs a soft reset of the STUSB4500 by setting and then clearing
// the REG_RESET_CTRL (0x23) register.
//
// This resets the internal registers and USB PD state machine, and it causes
// electrical disconnect on both source and sink sides. While reset, all pending
// interrupts and internal alerts are cleared.
//
// If the given bool argument wait is true, SoftReset will not return until the
// STUSB4500 is identified and communicating over I2C.
func (d *Device) SoftReset(wait bool) error {
	return d.resetUntilReady(
		func() error {
			// set reset-enable register field
			if err := d.writeRegister(REG_RESET_CTRL,
				(&regControlReset{reset: true}).format()[0]); nil != err {
				return err
			}
			// start measuring time once we initiated reset
			start := time.Now()
			// clear alerts, updating port status
			if err := d.updateStatus(); nil != err {
				return err
			}
			// sleep for the remaining reset duration (if any), after subtracting the
			// duration it took to clear the alert registers
			if remaining := softResetDuration - time.Now().Sub(start); remaining > 0 {
				time.Sleep(remaining)
			}
			// clear reset-enable register field
			if err := d.writeRegister(REG_RESET_CTRL,
				(&regControlReset{reset: false}).format()[0]); nil != err {
				return err
			}
			return nil
		},
		wait)
}

// Reset performs a hard reset of the STUSB4500 by asserting and de-asserting
// the active-high RST pin for a short duration.
//
// All registers and power contracts are reset to their default power-on state.
//
// If the given bool argument wait is true, Reset will not return until the
// STUSB4500 is identified and communicating over I2C.
func (d *Device) Reset(wait bool) error {
	return d.resetUntilReady(
		func() error {
			// perform hardware reset using RST pin (active high)
			d.config.ResetPin.High()
			time.Sleep(hardResetDuration)
			d.config.ResetPin.Low()
			return nil
		},
		wait)
}

// SetPower attempts to negotiate an explicit power contract with given voltage
// and current from the PD source.
//
// The given voltage, in millivolts (mV), must be a multiple of 50 mV.
// The given current, in milliamps (mA), must be a multiple of 10 mA.
func (d *Device) SetPower(voltage uint32, current uint32) error {
	if connected, _ := d.GetCableStatus(); connected {
		// PDO 1 is fixed (see comments inside SetPowerUSBDefault). So we construct
		// a custom PDO with given voltage/current and set it as PDO 2. Leave PDO 3
		// undefined, and send a PD negotiation request with the contents of PDO 2,
		// falling back to PDO 1 (USB default +5V) if there was a failure.
		if err := d.setNumSnkPDO(2); nil != err {
			return err
		}
		if err := d.setSnkPDO(
			PDO{
				Number:  2,
				Voltage: (voltage / pdoVoltageUnits) * pdoVoltageUnits,
				Current: (current / pdoCurrentUnits) * pdoCurrentUnits,
			}); nil != err {
			return err
		}
		if err := d.updateSnkPDO(); nil != err {
			return err
		}
		if err := d.updateSnkRDO(); nil != err {
			return err
		}
		if err := d.resetCable(); nil != err {
			return err
		}
	}
	return nil
}

// SetPowerUSBDefault selects the default +5V (0.5-3.0A) USB power profile. This
// is the fallback power profile used when USB PD negotiation fails and is the
// same power profile used by all USB 2.0 devices.
func (d *Device) SetPowerUSBDefault() error {
	if connected, _ := d.GetCableStatus(); connected {
		// PDO 1 is fixed and is always USB default (+5V). So simply remove all
		// other PDOs, forcing negotiation to always select that profile.
		if err := d.setNumSnkPDO(1); nil != err {
			return err
		}
		if err := d.resetCable(); nil != err {
			return err
		}
		if err := d.updateSnkPDO(); nil != err {
			return err
		}
		if err := d.updateSnkRDO(); nil != err {
			return err
		}
	}
	return nil
}

func (d *Device) writeRegister(addr uint8, data uint8) error {
	return d.bus.WriteRegister(d.address, addr, []uint8{data})
}

func (d *Device) writeRegisters(addr uint8, data ...uint8) error {
	return d.bus.WriteRegister(d.address, addr, data)
}

// readRegister returns the byte in a given register subaddress from the
// connected peripheral Device.
func (d *Device) readRegister(addr uint8) (uint8, error) {
	data := make([]uint8, 1)
	if err := d.bus.ReadRegister(d.address, addr, data); nil != err {
		return 0, err
	}
	return data[0], nil
}

// readRegisters returns a slice of count bytes starting at the given register
// subaddress from the connected peripheral Device.
func (d *Device) readRegisters(addr uint8, count int) ([]uint8, error) {
	data := make([]uint8, count)
	if err := d.bus.ReadRegister(d.address, addr, data); nil != err {
		return nil, err
	}
	return data, nil
}

func (d *Device) setNumSnkPDO(num uint8) error {
	if num < 1 || num > pdoSnkMax {
		return ErrInvalidNumPDO
	}
	return d.writeRegister(REG_DPM_PDO_NUMB, num)
}

func (d *Device) setSnkPDO(pdo PDO) error {
	const pdoSnkMax = 3
	if pdo.Number <= 1 || pdo.Number > pdoSnkMax {
		return ErrInvalidNumPDO
	}

	// First we need to convert the convenience wrapper type `PDO` to the 32-bit
	// register (4x 8-bit registers) definition for the STUSB4500. To simplify
	// this, we simply copy PDO 1 (USB default +5V) 32-bit register definition as
	// template, replacing its voltage and current from our given PDO. We choose
	// PDO 1 because it is fixed and must always exist.
	//
	// It's assumed we already received the PDO 1 32-bit register definition
	// during device initialization. If for some reason it does not exist, you
	// will need to call `updateSnkPDO` to retrieve it.
	def := d.status.snk[0]
	addr := uint8(REG_DPM_SNK_PDO1_0 + 4*(pdo.Number-1))
	def.cons.voltage = pdo.Voltage / pdoVoltageUnits
	def.cons.operationalCurrent = pdo.Current / pdoCurrentUnits

	return d.writeRegisters(addr, def.format()[0:4]...)
}

func (d *Device) updateSnkPDO() error {
	numSnkPDO, err := d.readRegister(REG_DPM_PDO_NUMB)
	if err != nil {
		return err
	}
	d.status.snk = make([]regStatusSnkPDO, numSnkPDO)
	d.SnkPDO = []PDO{}
	// read all of the 32-bit (4 registers each) SNK PDO registers
	p, err := d.readRegisters(REG_DPM_SNK_PDO1_0, int(numSnkPDO)*4)
	if nil != err {
		return err
	}
	for i, j := 0, 0; i < int(numSnkPDO); i, j = i+1, j+4 {
		d.status.snk[i].parse(p[j : j+4]...)
		d.SnkPDO = append(d.SnkPDO, PDO{
			Number:  i + 1,
			Voltage: d.status.snk[i].cons.voltage * 50,
			Current: d.status.snk[i].cons.operationalCurrent * 10,
		})
	}
	return nil
}

func (d *Device) updateSnkRDO() error {
	// read the 32-bit (4 registers) SNK RDO register
	r, err := d.readRegisters(REG_RDO_REG_STATUS_0, 4)
	if nil != err {
		return err
	}
	d.status.rdo.parse(r...)
	if d.status.rdo.objectPos == 0 {
		d.SnkRDO = invalidPDO
		return ErrUndefinedRDO
	}
	d.SnkRDO = PDO{
		Number:     int(d.status.rdo.objectPos),
		Current:    d.status.rdo.operatingCurrent * 10,
		MaxCurrent: d.status.rdo.maxCurrent * 10,
	}
	if d.SnkRDO.Number <= len(d.status.snk) {
		d.SnkRDO.Voltage = d.status.snk[d.SnkRDO.Number-1].cons.voltage * 50
	}
	return nil
}

func (d *Device) updateSrcPDO() error {
	d.fsm.request.Set(1)
	for try := 0; try < d.config.USBPDTimeout; try++ {
		if !d.fsm.request.HasBits(1) {
			return nil // source capabilities received
		}
		if err := d.resetCable(); nil != err {
			return err
		}
		if err := d.Update(); nil != err {
			return err
		}
	}
	if d.fsm.request.HasBits(1) {
		return ErrSourcePDOTimeout
	}
	return nil
}

func (d *Device) enableAlerts() error {
	// unmask the alert interrupts needed for general operation. any field set to
	// false below will effectively enable that alert (specifically, it will NOT
	// be masked off when reading the alert status register).
	// note that calling `updateStatus` will read all of these status registers
	// and update the internal status struct of the receiver device regardless of
	// what is masked here. the masking only affects which status registers are
	// handled by the alert interrupt service routine `processAlerts`.
	return d.writeRegister(REG_ALERT_STATUS_1_MASK,
		(&regControlAlert{
			phy:       true, // don't care
			prt:       false,
			typeC:     false,
			hwFault:   true, // don't care
			monitor:   false,
			ccDetect:  false,
			hardReset: false,
		}).format()[0])
}

func (d *Device) updateStatus() error {
	// read all of the port-related status registers in a single I2C multi-read
	// request. many of these are alert registers (R/C) and will be cleared when
	// read, so we copy their content to local application memory.
	data, err := d.readRegisters(REG_ALERT_STATUS_1, 13)
	if nil != err {
		return err
	}
	// I2C read base address = 0x0B (REG_ALERT_STATUS_1)
	d.status.ccDetect.parse(data[3]) // regStatusCCDetect = 0x0E [0x0B+03]
	d.status.typeCMon.parse(data[5]) // regStatusTypeCMon = 0x10 [0x0B+05]
	d.status.cc.parse(data[6])       // regStatusCC       = 0x11 [0x0B+06]
	d.status.prt.parse(data[11])     // regStatusPRT      = 0x16 [0x0B+11]
	d.status.phy.parse(data[12])     // regStatusPHY      = 0x17 [0x0B+12]
	return nil
}

func (d *Device) resetCable() error {
	// send PD message "soft reset" to source by setting TX header (0x51) to 0x0D,
	// and set PD command (0x1A) to 0x26.
	if err := d.writeRegister(REG_TX_HEADER_LOW, ctrlSoftReset); nil != err {
		return err
	}
	return d.writeRegister(REG_PD_COMMAND_CTRL,
		(&regControlPDCmd{cmd: ctrlPDCommand}).format()[0])
}

type resetFunc func() error

func (d *Device) resetUntilReady(reset resetFunc, wait bool) error {
	// always perform the initial reset
	if err := reset(); nil != err {
		return err
	}
	// wait for an STUSB4500 to start responding, occassionally repeating the
	// reset again if a response isn't received after adequate time/attempts.
	for wait {
		found := false
		// ensure enough time has elapsed for NVM loading and PD negotiation
		time.Sleep(negotiationDuration)
		// timeout loop: check that we are communicating with the device
		for i := 0; i < d.config.USBPDTimeout; i++ {
			// delay before try (NOT after); so if the last try fails, we return ASAP
			if i > 0 { // also, never delay before the first try
				time.Sleep(reconnectWaitDuration)
			}
			// check if we found connection, breaking out of timeout loop on success
			if found = d.Connected(); found {
				break
			}
		}
		// found will be true if and only if we broke timeout loop prematurely, in
		// which case we do not issue another reset, but instead gracefully exit
		// outer loop and return to caller (because wait will be false).
		if wait = !found; wait {
			// reset device again!
			if err := reset(); nil != err {
				return err
			}
			// next we will sleep for negotiationDuration and then enter timeout loop.
		}
	}
	return nil
}

func (d *Device) processAlerts() error {

	al, err := d.readRegisters(REG_ALERT_STATUS_1, 2)
	if nil != err {
		return err
	}
	stat := al[0] & ^(al[1])

	// parse the interrupt status
	var alert regStatusAlert
	alert.parse(stat)

	if 0 != stat {

		if alert.prt { // bit 2

			// parse PRT status
			prt, err := d.readRegister(REG_PRT_STATUS)
			if nil != err {
				return err
			}
			d.status.prt.parse(prt)

			// if interrupt status contains a new USB PD message
			if d.status.prt.msgReceived {

				// parse the PD message header
				heads, err := d.readRegisters(REG_RX_HEADER_LOW, 2)
				if nil != err {
					return err
				}
				var header msgUsbpdHeader
				header.parse(heads...)

				// if header contains PDO definitions
				if header.dataObjectCount > 0 {

					// verify the received byte count matches expected number of PDOs
					cnt, err := d.readRegister(REG_RX_BYTE_CNT)
					if nil != err {
						return err
					}
					if cnt != 4*header.dataObjectCount {
						return ErrByteCountPDHeader
					}

					switch header.messageType {
					case dataMsgSourceCap:
						// read all SRC PDO registers
						p, err := d.readRegisters(REG_RX_DATA_OBJ1_0, int(cnt))
						if nil != err {
							return err
						}
						// clear existing SRC PDO list
						d.status.src = make([]regStatusSrcPDO, header.dataObjectCount)
						d.SrcPDO = []PDO{}
						// parse received source PDO list
						for i, j := 0, 0; i < int(header.dataObjectCount); i, j = i+1, j+4 {
							d.status.src[i].parse(p[j : j+4]...)
							if i == 0 {
								d.status.src[i].cons.voltage = 100
							}
							d.SrcPDO = append(d.SrcPDO, PDO{
								Number:  i + 1,
								Voltage: d.status.src[i].cons.voltage * 50,
								Current: d.status.src[i].cons.maxOperatingCurrent * 10,
							})
						}
						d.fsm.request.Set(0)
						d.fsm.receive.Set(d.fsm.receive.Get() + 1)

					case dataMsgRequest:
					case dataMsgSinkCap:
					case dataMsgVendorDefined:
					default:
					}

				} else {

					switch header.messageType {
					case ctrlMsgGoodCRC:
						d.fsm.msgCRC++

					case ctrlMsgAccept:
						d.fsm.msgAccept++

					case ctrlMsgReject:
						d.fsm.msgReject++

					case ctrlMsgPSRDY:
						d.fsm.psrdyRecv++

					case ctrlMsgGetSourceCap:
					case ctrlMsgGetSinkCap:
					case ctrlMsgWait:
					case ctrlMsgSoftReset:
					case ctrlMsgNotSupported:
					case ctrlMsgGetSourceCapExt:
					case ctrlMsgGetStatus:
					case ctrlMsgFRSwap:
					case ctrlMsgGetPPSStatus:
					default:
					}

				}
			}
		}

		d.status.reset = alert.hardReset
		if alert.hardReset { // bit 8
			d.fsm.irqReset++
		}

		if alert.ccDetect { // bit 7
			ccDetect, err := d.readRegisters(REG_PORT_STATUS_0, 2)
			if nil != err {
				return err
			}
			d.status.ccDetect.parse(ccDetect[1])
			var trans regStatusCCDetectTrans
			trans.parse(ccDetect[0])
			if trans.attached {
				d.fsm.trans++
			}
		}

		if alert.monitor { // bit 6
			typeCMon, err := d.readRegisters(REG_TYPEC_MONITORING_STATUS_0, 2)
			if nil != err {
				return err
			}
			d.status.typeCMon.parse(typeCMon[1])
		}

		// always read and update CC attachment status
		cc, err := d.readRegister(REG_CC_STATUS)
		if nil != err {
			return err
		}
		d.status.cc.parse(cc)

		if alert.hwFault { // bit 5
			hwFault, err := d.readRegisters(REG_CC_HW_FAULT_STATUS_0, 2)
			if nil != err {
				return err
			}
			d.status.hwFault.parse(hwFault[1])
		}
	}
	return nil
}

func (d *Device) onAlert(machine.Pin) {
	d.fsm.alert.Set(d.fsm.alert.Get() + 1)
}

func (d *Device) onAttach(machine.Pin) {
	d.fsm.attach.Set(d.fsm.attach.Get() + 1)
}
