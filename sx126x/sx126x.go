// Package sx126x provides a driver for SX126x LoRa transceivers.
// Inspired from https://github.com/Lora-net/sx126x_driver/

package sx126x

import (
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

// SX126X radio transceiver RF_IN and RF_OUT may be connected
// to RF Switch. This interface allows the creation of struct
// that can drive the RF Switch (Used in Lora RX and Lora Tx)
type RFSwitch interface {
	InitRFSwitch()
	SetRfSwitchMode(mode int) error
}

const (
	DEVICE_TYPE_SX1261 = iota
	DEVICE_TYPE_SX1262 = iota
	DEVICE_TYPE_SX1268 = iota
)

const (
	RFSWITCH_RX    = iota
	RFSWITCH_TX_LP = iota
	RFSWITCH_TX_HP = iota
)

const (
	RadioEventRxDone    = iota
	RadioEventTxDone    = iota
	RadioEventTimeout   = iota
	RadioEventWatchdog  = iota
	RadioEventCrcError  = iota
	RadioEventUnhandled = iota
)

// RadioEvent are used for communicating in the radio Event Channel
type RadioEvent struct {
	EventType int
	IRQStatus uint16
	EventData []byte
}

const (
	PERIOD_PER_SEC  = (uint32)(1000000 / 15.625) // SX1261 DS 13.1.4
	SPI_BUFFER_SIZE = 256
)

// Device wraps an SPI connection to a SX127x device.
type Device struct {
	spi            drivers.SPI     // SPI bus for module communication
	radioEventChan chan RadioEvent // Channel for Receiving events
	loraConf       LoraConfig      // Current Lora configuration
	rfswitch       RFSwitch        // RF Switch, if any
	deepSleep      bool            // Internal Sleep state
	deviceType     int             // sx1261,sx1262,sx1268 (defaults sx1261)
	spiBuffer      [SPI_BUFFER_SIZE]uint8
}

// Config holds the LoRa configuration parameters
type LoraConfig struct {
	Freq           uint32 // Frequency
	Cr             uint8  // Coding Rate
	Sf             uint8  // Spread Factor
	Bw             uint8  // Bandwidth
	Ldr            uint8  // Low Data Rate
	Preamble       uint16 // PreambleLength
	SyncWord       uint16 // Sync Word
	HeaderType     uint8  // Header : Implicit/explicit
	Crc            uint8  // CRC : Yes/No
	Iq             uint8  // iq : Standard/inverted
	LoraTxPowerDBm int8   // Tx power in Dbm
}

const (
	SX126X_RTC_FREQ_IN_HZ uint32 = 64000
)

var (
	errUndefinedLoraConf = errors.New("Undefined Lora configuration")
)

// --------------------------------------------------
//  Helper functions
// --------------------------------------------------

// timeoutMsToRtcSteps converts Timeout (in ms) to RTC Steps
func timeoutMsToRtcSteps(timeoutMs uint32) uint32 {
	r := uint32(timeoutMs * (SX126X_RTC_FREQ_IN_HZ / 1000))
	return r
}

// --------------------------------------------------
//
//	Channel and events
//
// --------------------------------------------------
// NewRadioEvent() returns a new RadioEvent that can be used in the RadioChannel
func NewRadioEvent(eType int, irqStatus uint16, eData []byte) RadioEvent {
	r := RadioEvent{EventType: eType, IRQStatus: irqStatus, EventData: eData}
	return r
}

// Get the RadioEvent channel of the device
func (d *Device) GetRadioEventChan() chan RadioEvent {
	return d.radioEventChan
}

// Specify device type (SX1261/2/8)
func (d *Device) SetDeviceType(devType int) {
	d.deviceType = devType
}

// SetRfSwitch let you define a custom RF Switch driver if needed
func (d *Device) SetRfSwitch(rfswitch RFSwitch) {
	d.rfswitch = rfswitch
	d.rfswitch.InitRFSwitch()
}

// --------------------------------------------------
// Operational modes functions
// --------------------------------------------------

// DetectDevice() tries to detect the radio module by changing SyncWord value
func (d *Device) DetectDevice() bool {
	bak := d.GetSyncWord()
	d.SetSyncWord(0xBEEF)
	tmp := d.GetSyncWord()
	if tmp != 0xBEEF {
		return false
	} else {
		d.SetSyncWord(bak)
		return true
	}
}

// SetSleep sets the device in SLEEP mode with the lowest current consumption possible.
func (d *Device) SetSleep() {
	d.ExecSetCommand(SX126X_CMD_SET_SLEEP, []uint8{SX126X_SLEEP_START_WARM | SX126X_SLEEP_RTC_OFF})
}

// SetStandby sets the device in a configuration mode which is at an intermediate level of consumption
func (d *Device) SetStandby() {
	d.ExecSetCommand(SX126X_CMD_SET_STANDBY, []uint8{SX126X_STANDBY_RC})
}

// SetFs sets the device in frequency synthesis mode where the PLL is locked to the carrier frequency.
func (d *Device) SetFs() {
	d.ExecSetCommand(SX126X_CMD_SET_FS, []uint8{})
}

// SetTxContinuousWave set device in test mode to generate a continuous wave (RF tone)
func (d *Device) SetTxContinuousWave() {
	if d.rfswitch != nil {
		d.rfswitch.SetRfSwitchMode(RFSWITCH_TX_HP)
	}
	d.ExecSetCommand(SX126X_CMD_SET_TX_CONTINUOUS_WAVE, []uint8{})
}

// SetTxContinuousPreamble set device in test mode to constantly modulate LoRa preamble symbols.
// Take care to initialize all Lora settings like it's done in LoraTx before calling this function
// If you don't init properly all the settings, it'll fail
func (d *Device) SetTxContinuousPreamble() {
	if d.rfswitch != nil {
		d.rfswitch.SetRfSwitchMode(RFSWITCH_TX_HP)
	}
	d.ExecSetCommand(SX126X_CMD_SET_TX_INFINITE_PREAMBLE, []uint8{})
}

// SetTx() sets the device in TX mode
// timeout is expressed in RTC Step unit (15uS)
// The device will stay in Tx until countdown or packet transmitted
// Value of 0x000000 will disable timer and device will stay TX
func (d *Device) SetTx(timeoutRtcStep uint32) {
	var p [3]uint8
	p[0] = uint8((timeoutRtcStep >> 16) & 0xFF)
	p[1] = uint8((timeoutRtcStep >> 8) & 0xFF)
	p[2] = uint8((timeoutRtcStep >> 0) & 0xFF)
	d.ExecSetCommand(SX126X_CMD_SET_TX, p[:])
}

// SetRx() sets the device in RX mode
// timeout is expressed in RTC Step unit (15uS)
// Value of 0x000000 => No timeout. Rx Single mode.
// Value of 0xffffff => Rx Continuous mode
// Other values => Timeout active. The device remains in RX until countdown or packet received
func (d *Device) SetRx(timeoutRtcStep uint32) {
	var p [3]uint8
	p[0] = uint8(((timeoutRtcStep >> 16) & 0xFF))
	p[1] = uint8(((timeoutRtcStep >> 8) & 0xFF))
	p[2] = uint8(((timeoutRtcStep >> 0) & 0xFF))
	d.ExecSetCommand(SX126X_CMD_SET_RX, p[:])
}

// StopTimerOnPreamble allows the user to select if the timer is stopped upon preamble detection of SyncWord / header detection.
func (d *Device) StopTimerOnPreamble(enable bool) {
	var p [1]uint8
	if enable {
		p[0] = 1
	} else {
		p[0] = 0
	}
	d.ExecSetCommand(SX126X_CMD_STOP_TIMER_ON_PREAMBLE, p[:])
}

// SetRegulatorMode sets the regulator more (depends on hardware implementation)
func (d *Device) SetRegulatorMode(mode uint8) {
	p := []uint8{mode}
	d.ExecSetCommand(SX126X_CMD_SET_REGULATOR_MODE, p[:])
}

// Calibrate starts the calibration of a block defined by calibParam
func (d *Device) Calibrate(calibParam uint8) {
	p := []uint8{calibParam}
	d.ExecSetCommand(SX126X_CMD_CALIBRATE, p[:])
}

// CalibrateImage calibrates the image rejection of the device for the device operating
func (d *Device) CalibrateImage(freq uint32) {
	var calFreq [2]uint8
	if freq > 900000000 {
		calFreq[0] = 0xE1
		calFreq[1] = 0xE9
	} else if freq > 850000000 {
		calFreq[0] = 0xD7
		calFreq[1] = 0xD8
	} else if freq > 770000000 {
		calFreq[0] = 0xC1
		calFreq[1] = 0xC5
	} else if freq > 460000000 {
		calFreq[0] = 0x75
		calFreq[1] = 0x81
	} else if freq > 425000000 {
		calFreq[0] = 0x6B
		calFreq[1] = 0x6F
	}
	d.ExecSetCommand(SX126X_CMD_CALIBRATE_IMAGE, calFreq[:])
}

// SetPaConfig sets the Power Amplifier configuration
// deviceSel: 0 for SX1262, 1 for SX1261
func (d *Device) SetPaConfig(paDutyCycle, hpMax, deviceSel, paLut uint8) {
	var p [4]uint8
	p[0] = paDutyCycle
	p[1] = hpMax
	p[2] = deviceSel
	p[3] = paLut
	d.ExecSetCommand(SX126X_CMD_SET_PA_CONFIG, p[:])
}

// SetRxTxFallbackMode defines into which mode the chip goes after a successful transmission or after a packet reception.
func (d *Device) SetRxTxFallbackMode(fallbackMode uint8) {
	d.ExecSetCommand(SX126X_CMD_SET_RX_TX_FALLBACK_MODE, []uint8{fallbackMode})
}

// --------------------------------------------------
// Registers and Buffers
// --------------------------------------------------

// ReadRegister reads register value
func (d *Device) ReadRegister(addr, size uint16) ([]uint8, error) {
	d.CheckDeviceReady()
	d.SpiSetNss(false)
	// Send command
	cmd := []uint8{SX126X_CMD_READ_REGISTER, uint8((addr & 0xFF00) >> 8), uint8(addr & 0x00FF), 0x00}
	d.spi.Tx(cmd, nil)
	ret := d.spiBuffer[0:size]
	d.spi.Tx(nil, ret)
	d.SpiSetNss(true)
	d.WaitBusy()
	return ret, nil
}

// WriteRegister writes value to register
func (d *Device) WriteRegister(addr uint16, data []uint8) {
	d.CheckDeviceReady()
	d.SpiSetNss(false)
	cmd := []uint8{SX126X_CMD_WRITE_REGISTER, uint8((addr & 0xFF00) >> 8), uint8(addr & 0x00FF)}
	d.spi.Tx(append(cmd, data...), nil)
	d.SpiSetNss(true)
	d.WaitBusy()
}

// WriteBuffer write data from current buffer position
func (d *Device) WriteBuffer(data []uint8) {
	p := []uint8{0}
	p = append(p, data...)
	d.ExecSetCommand(SX126X_CMD_WRITE_BUFFER, p)
}

// ReadBuffer Reads size bytes from current buffer position
func (d *Device) ReadBuffer(size uint8) []uint8 {
	ret := d.ExecGetCommand(SX126X_CMD_READ_BUFFER, size)
	return ret
}

// --------------------------------------------------
// DIO and IRQ
// --------------------------------------------------

// SetDioIrqParams configures DIO Irq
func (d *Device) SetDioIrqParams(irqMask, dio1Mask, dio2Mask, dio3Mask uint16) {
	var p [8]uint8
	p[0] = uint8((irqMask >> 8) & 0xFF)
	p[1] = uint8(irqMask & 0xFF)
	p[2] = uint8((dio1Mask >> 8) & 0xFF)
	p[3] = uint8(dio1Mask & 0xFF)
	p[4] = uint8((dio2Mask >> 8) & 0xFF)
	p[5] = uint8(dio2Mask & 0xFF)
	p[6] = uint8((dio3Mask >> 8) & 0xFF)
	p[7] = uint8(dio3Mask & 0xFF)
	d.ExecSetCommand(SX126X_CMD_SET_DIO_IRQ_PARAMS, p[:])
}

// GetIrqStatus returns IRQ status
func (d *Device) GetIrqStatus() (irqStatus uint16) {
	r := d.ExecGetCommand(SX126X_CMD_GET_IRQ_STATUS, 2)
	ret := (uint16(r[0]) << 8) | uint16(r[1])
	return ret
}

// ClearIrqStatus clears IRQ flags
func (d *Device) ClearIrqStatus(clearIrqParams uint16) {
	var p [2]uint8
	p[0] = uint8((clearIrqParams >> 8) & 0xFF)
	p[1] = uint8(clearIrqParams & 0xFF)
	d.ExecSetCommand(SX126X_CMD_CLEAR_IRQ_STATUS, p[:])
}

// --------------------------------------------------
// Communication Status Information
// --------------------------------------------------

// GetStatus returns radio status(13.5.1)
func (d *Device) GetStatus() (radioStatus uint8) {
	r := d.ExecGetCommand(SX126X_CMD_GET_STATUS, 1)
	return r[0]
}

// GetRxBufferStatus returns the length of the last received packet (PayloadLengthRx)
// and the address of the first byte received (RxStartBufferPointer). (13.5.2)
func (d *Device) GetRxBufferStatus() (payloadLengthRx uint8, rxStartBufferPointer uint8) {
	r := d.ExecGetCommand(SX126X_CMD_GET_RX_BUFFER_STATUS, 2)
	return r[0], r[1]
}

// GetPackeType returns current Packet Type (13.4.3)
func (d *Device) GetPacketType() (packetType uint8) {
	r := d.ExecGetCommand(SX126X_CMD_GET_PACKET_TYPE, 1)
	return r[0]
}

// GetDeviceErrors returns current Device Errors
func (d *Device) GetDeviceErrors() uint16 {
	r := d.ExecGetCommand(SX126X_CMD_GET_DEVICE_ERRORS, 2)
	ret := uint16(r[0]<<8 + r[1])
	return ret
}

// ClearDeviceErrors clears device Errors
func (d *Device) ClearDeviceErrors() {
	p := [2]uint8{0x00, 0x00}
	d.ExecSetCommand(SX126X_CMD_CLEAR_DEVICE_ERRORS, p[:])
}

// GetStats returns the number of informations received on a few last packets
// Lora: NbPktReceived, NbPktCrcError, NbPktHeaderErr
func (d *Device) GetLoraStats() (nbPktReceived, nbPktCrcError, nbPktHeaderErr uint16) {
	r := d.ExecGetCommand(SX126X_CMD_GET_STATS, 6)
	return uint16(r[0]<<8 | r[1]), uint16(r[2]<<8 | r[3]), uint16(r[4]<<8 | r[5])
}

// ---------------------------------------
// PACKET / RADIO / PROTOCOL CONFIGURATION
// ---------------------------------------

// SetPacketType sets the packet type
func (d *Device) SetPacketType(packetType uint8) {
	var p [1]uint8
	p[0] = packetType
	d.ExecSetCommand(SX126X_CMD_SET_PACKET_TYPE, p[:])
}

// SetSyncWord defines the Sync Word to yse
func (d *Device) SetSyncWord(syncword uint16) {
	var p [2]uint8
	d.loraConf.SyncWord = syncword
	p[0] = uint8((syncword >> 8) & 0xFF)
	p[1] = uint8((syncword >> 0) & 0xFF)
	d.WriteRegister(SX126X_REG_LORA_SYNC_WORD_MSB, p[:])
}

// GetSyncWord gets the Sync Word to use
func (d *Device) GetSyncWord() uint16 {
	p, _ := d.ReadRegister(SX126X_REG_LORA_SYNC_WORD_MSB, 2)
	r := uint16(p[0])<<8 + uint16(p[1])
	return r
}

// SetLoraPublicNetwork sets Sync Word to 0x3444 (Public) or 0x1424 (Private)
func (d *Device) SetLoraPublicNetwork(enable bool) {
	if enable {
		d.SetSyncWord(SX126X_LORA_MAC_PUBLIC_SYNCWORD)
	} else {
		d.SetSyncWord(SX126X_LORA_MAC_PRIVATE_SYNCWORD)
	}
}

// SetPacketParam sets various packet-related params
func (d *Device) SetPacketParam(preambleLength uint16, headerType, crcType, payloadLength, invertIQ uint8) {
	var p [6]uint8
	p[0] = uint8((preambleLength >> 8) & 0xFF)
	p[1] = uint8(preambleLength & 0xFF)
	p[2] = headerType
	p[3] = payloadLength
	p[4] = crcType
	p[5] = invertIQ
	d.ExecSetCommand(SX126X_CMD_SET_PACKET_PARAMS, p[:])
}

// SetBufferBaseAddress sets base address for buffer
func (d *Device) SetBufferBaseAddress(txBaseAddress, rxBaseAddress uint8) {
	var p [2]uint8
	p[0] = txBaseAddress
	p[1] = rxBaseAddress
	d.ExecSetCommand(SX126X_CMD_SET_BUFFER_BASE_ADDRESS, p[:])
}

// SetRfFrequency sets the radio frequency
func (d *Device) SetRfFrequency(frequency uint32) {
	var p [4]uint8
	freq := uint32((uint64(frequency) << 25) / 32000000)
	p[0] = uint8((freq >> 24) & 0xFF)
	p[1] = uint8((freq >> 16) & 0xFF)
	p[2] = uint8((freq >> 8) & 0xFF)
	p[3] = uint8((freq >> 0) & 0xFF)
	d.ExecSetCommand(SX126X_CMD_SET_RF_FREQUENCY, p[:])
}

// SetCurrentLimit sets max current in the module
func (d *Device) SetCurrentLimit(limit uint8) {
	if limit > 140 {
		limit = 140
	}
	rawLimit := uint8(float32(limit) / 2.5)
	p := []uint8{rawLimit}
	d.WriteRegister(SX126X_REG_OCP_CONFIGURATION, p[:])
}

// SetTxConfig sets power and rampup time
func (d *Device) SetTxParams(power int8, rampTime uint8) {
	var p [2]uint8

	if d.deviceType == DEVICE_TYPE_SX1261 {
		if power == 15 {
			d.SetPaConfig(0x06, 0x00, 0x01, 0x01)
		} else {
			d.SetPaConfig(0x04, 0x00, 0x01, 0x01)
		}
		if power > 14 {
			power = 14
		} else if power < -3 {
			power = -3
		}
		d.SetCurrentLimit(80) // Set max current limit to 80mA
	} else { // sx1262 and sx1268
		d.SetPaConfig(0x04, 0x07, 0x00, 0x01)
		if power > 22 {
			power = 22
		} else if power < -3 {
			power = -3
		}
		d.SetCurrentLimit(140) // Set max current limit to 140 mA
	}

	p[0] = uint8(power)
	p[1] = rampTime
	d.ExecSetCommand(SX126X_CMD_SET_TX_PARAMS, p[:])
}

// SetModulationParams sets the Lora modulation frequency
func (d *Device) SetModulationParams(spreadingFactor, bandwidth, codingRate, lowDataRateOptimize uint8) {
	var p [4]uint8
	p[0] = spreadingFactor
	p[1] = bandwidth
	p[2] = codingRate
	p[3] = lowDataRateOptimize
	d.ExecSetCommand(SX126X_CMD_SET_MODULATION_PARAMS, p[:])
}

// CheckDeviceReady sleep until all busy flags clears
func (d *Device) CheckDeviceReady() error {
	if d.deepSleep == true {
		d.SpiSetNss(false)
		time.Sleep(time.Millisecond)
		d.SpiSetNss(true)
		d.deepSleep = false
	}
	return d.WaitBusy()
}

// ExecSetCommand send a command to configure the peripheral
func (d *Device) ExecSetCommand(cmd uint8, buf []uint8) {
	d.CheckDeviceReady()
	if cmd == SX126X_CMD_SET_SLEEP {
		d.deepSleep = true
	} else {
		d.deepSleep = false
	}
	d.SpiSetNss(false)
	// Send command and params
	d.spi.Tx(append([]uint8{cmd}, buf...), nil)
	d.SpiSetNss(true)
	if cmd != SX126X_CMD_SET_SLEEP {
		d.WaitBusy()
	}
}

// ExecGetCommand queries the peripheral the peripheral
func (d *Device) ExecGetCommand(cmd uint8, size uint8) []uint8 {
	d.CheckDeviceReady()
	d.SpiSetNss(false)
	// Send the command and flush first status byte (as not used)
	d.spi.Tx([]uint8{cmd, 0x00}, nil)
	d.spi.Tx(nil, d.spiBuffer[:size])
	d.SpiSetNss(true)
	d.WaitBusy()
	return d.spiBuffer[:size]
}

//
// Configuration
//

// SetLoraFrequency() Sets current Lora Frequency
// NB: Change will be applied at next RX / TX
func (d *Device) SetLoraFrequency(freq uint32) {
	d.loraConf.Freq = d.loraConf.Freq
}

// SetLoraIqMode() defines the current IQ Mode (Standard/Inverted)
// NB: Change will be applied at next RX / TX
func (d *Device) SetLoraIqMode(mode uint8) {
	if mode == 0 {
		d.loraConf.Iq = SX126X_LORA_IQ_STANDARD
	} else {
		d.loraConf.Iq = SX126X_LORA_IQ_INVERTED
	}
}

// SetLoraCodingRate() sets current Lora Coding Rate
// NB: Change will be applied at next RX / TX
func (d *Device) SetLoraCodingRate(cr uint8) {
	d.loraConf.Cr = cr
}

// SetLoraBandwidth() sets current Lora Bandwidth
// NB: Change will be applied at next RX / TX
func (d *Device) SetLoraBandwidth(bw uint8) {
	d.loraConf.Cr = bw
}

// SetLoraCrc() sets current CRC mode (ON/OFF)
// NB: Change will be applied at next RX / TX
func (d *Device) SetLoraCrc(enable bool) {
	if enable {
		d.loraConf.Crc = SX126X_LORA_CRC_ON
	} else {
		d.loraConf.Crc = SX126X_LORA_CRC_OFF
	}
}

// SetLoraSpreadingFactor setc surrent Lora Spreading Factor
// NB: Change will be applied at next RX / TX
func (d *Device) SetLoraSpreadingFactor(sf uint8) {
	d.loraConf.Sf = sf
}

//
// Lora functions
//
//

// LoraConfig() defines Lora configuration for next Lora operations
func (d *Device) LoraConfig(cnf LoraConfig) {
	// Save given configuration
	d.loraConf = cnf
	// Switch to standby prior to configuration changes
	d.SetStandby()
	// Clear errors, disable radio interrupts for the moment
	d.ClearDeviceErrors()
	d.ClearIrqStatus(SX126X_IRQ_ALL)
	d.SetDioIrqParams(0x00, 0x00, 0x00, 0x00)
	// Define radio operation mode
	d.SetPacketType(SX126X_PACKET_TYPE_LORA)
	d.SetRfFrequency(d.loraConf.Freq)
	d.SetModulationParams(d.loraConf.Sf, d.loraConf.Bw, d.loraConf.Cr, d.loraConf.Ldr)
	d.SetTxParams(d.loraConf.LoraTxPowerDBm, SX126X_PA_RAMP_200U)
	d.SetSyncWord(d.loraConf.SyncWord)
	d.SetBufferBaseAddress(0, 0)
}

// LoraTx sends a lora packet, (with timeout)
func (d *Device) LoraTx(pkt []uint8, timeoutMs uint32) error {

	if d.loraConf.Freq == 0 {
		return errUndefinedLoraConf
	}
	if d.rfswitch != nil {
		err := d.rfswitch.SetRfSwitchMode(RFSWITCH_TX_HP)
		if err != nil {
			return err
		}
	}
	d.ClearIrqStatus(SX126X_IRQ_ALL)
	irqVal := uint16(SX126X_IRQ_TX_DONE | SX126X_IRQ_TIMEOUT | SX126X_IRQ_CRC_ERR)
	d.SetStandby()
	d.SetPacketType(SX126X_PACKET_TYPE_LORA)
	d.SetRfFrequency(d.loraConf.Freq)
	d.SetTxParams(d.loraConf.LoraTxPowerDBm, SX126X_PA_RAMP_200U)
	d.SetBufferBaseAddress(0, 0)
	d.WriteBuffer(pkt)
	d.SetModulationParams(d.loraConf.Sf, d.loraConf.Bw, d.loraConf.Cr, d.loraConf.Ldr)
	d.SetPacketParam(d.loraConf.Preamble, d.loraConf.HeaderType, d.loraConf.Crc, uint8(len(pkt)), d.loraConf.Iq)
	d.SetDioIrqParams(irqVal, irqVal, SX126X_IRQ_NONE, SX126X_IRQ_NONE)
	d.SetSyncWord(d.loraConf.SyncWord)
	d.SetTx(timeoutMsToRtcSteps(timeoutMs))

	msg := <-d.GetRadioEventChan()
	if msg.EventType != RadioEventTxDone {
		return errors.New("Unexpected Radio Event while TX")
	}
	return nil
}

// LoraRx tries to receive a Lora packet (with timeout in milliseconds)
func (d *Device) LoraRx(timeoutMs uint32) ([]uint8, error) {

	if d.loraConf.Freq == 0 {
		return nil, errUndefinedLoraConf
	}
	if d.rfswitch != nil {
		err := d.rfswitch.SetRfSwitchMode(RFSWITCH_RX)
		if err != nil {
			return nil, err
		}
	}
	d.ClearIrqStatus(SX126X_IRQ_ALL)
	irqVal := uint16(SX126X_IRQ_RX_DONE | SX126X_IRQ_TIMEOUT | SX126X_IRQ_CRC_ERR)
	d.SetStandby()
	d.SetBufferBaseAddress(0, 0)
	d.SetModulationParams(d.loraConf.Sf, d.loraConf.Bw, d.loraConf.Cr, d.loraConf.Ldr)
	d.SetPacketParam(d.loraConf.Preamble, d.loraConf.HeaderType, d.loraConf.Crc, 0xFF, d.loraConf.Iq)
	d.SetDioIrqParams(irqVal, irqVal, SX126X_IRQ_NONE, SX126X_IRQ_NONE)
	d.SetRx(timeoutMsToRtcSteps(timeoutMs))

	msg := <-d.GetRadioEventChan()

	if msg.EventType == RadioEventTimeout {
		return nil, nil
	} else if msg.EventType != RadioEventRxDone {
		return nil, errors.New("Unexpected Radio Event while RX")
	}

	pLen, pStart := d.GetRxBufferStatus()
	d.SetBufferBaseAddress(0, pStart+1)
	pkt := d.ReadBuffer(pLen + 1)
	pkt = pkt[1:]

	return pkt, nil
}

// HandleInterrupt must be called by main code on DIO state change.
func (d *Device) HandleInterrupt() {
	st := d.GetIrqStatus()
	d.ClearIrqStatus(SX126X_IRQ_ALL)

	rChan := d.GetRadioEventChan()

	if (st & SX126X_IRQ_RX_DONE) > 0 {
		rChan <- NewRadioEvent(RadioEventRxDone, st, nil)
	}

	if (st & SX126X_IRQ_TX_DONE) > 0 {
		rChan <- NewRadioEvent(RadioEventTxDone, st, nil)
	}

	if (st & SX126X_IRQ_TIMEOUT) > 0 {
		rChan <- NewRadioEvent(RadioEventTimeout, st, nil)
	}

	if (st & SX126X_IRQ_CRC_ERR) > 0 {
		rChan <- NewRadioEvent(RadioEventCrcError, st, nil)
	}

}
