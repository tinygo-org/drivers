// Package rfm69 RFM69 Implementation in Go
package rfm69

import (
	"errors"
	"fmt"
	"machine"
	"time"
	"tinygo.org/x/drivers"
)

// OnReceiveHandler is the receive callback
type OnReceiveHandler func(*Data)

type RFMOptions struct {
	NodeID        byte
	NetworkID     byte
	IsRfm69HCW    bool
	EncryptionKey string
	ResetPin      machine.Pin
	IrqPin        machine.Pin
	CsPin         machine.Pin
}

// Device RFM69 Device
type Device struct {
	bus        drivers.SPI
	mode       byte
	Config     *RFMOptions
	powerLevel byte
	tx         chan *Data
	quit       chan bool
	irq        chan bool
	OnReceive  OnReceiveHandler
}

// Global settings
const (
	CsmaLimit  = -80
	MaxDataLen = 66
)

// NewDevice creates a new device
func NewDevice(spi drivers.SPI, options *RFMOptions) (*Device, error) {
	options.IrqPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	options.CsPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	if options == nil {
		options = new(RFMOptions)
		options.IsRfm69HCW = true
		options.NodeID = 1
		options.NetworkID = 100
		options.ResetPin = machine.NoPin
		options.IrqPin = machine.GP20
		options.CsPin = machine.GP17
	}

	ret := &Device{
		bus:        spi,
		Config:     options,
		powerLevel: 31,
		tx:         make(chan *Data, 5),
		quit:       make(chan bool),
		irq:        make(chan bool),
	}

	err := ret.setup()
	if err != nil {
		return nil, err
	}

	go ret.loop()

	return ret, nil
}

// Close cleans up
func (r *Device) Close() error {

	r.quit <- true
	<-r.quit

	return nil
}

func (r *Device) SetupInterrupts() error {
	if err := r.Config.IrqPin.SetInterrupt(machine.PinRising, r.IRQHandle); err != nil {
		return err
	}
	return nil
}

func (r *Device) IRQHandle(pin machine.Pin) {
	defer close(r.irq)
	r.irq <- true
}

func (r *Device) writeReg(addr, data byte) error {
	tx := make([]byte, 2)
	tx[0] = addr | 0x80
	tx[1] = data
	fmt.Printf("write %x: %x\n", addr, data)
	length := len(tx)
	rx := make([]byte, length)
	r.Config.CsPin.High()
	time.Sleep(1 * time.Millisecond)
	r.Config.CsPin.Low()
	err := r.bus.Tx(tx, rx)
	if err != nil {
		fmt.Print(err)
	}
	r.Config.CsPin.High()
	return err
}

func (r *Device) readReg(addr byte) (byte, error) {
	tx := make([]uint8, 2)
	tx[0] = addr & 0x7f
	tx[1] = 0
	length := len(tx)
	rx := make([]byte, length)
	r.Config.CsPin.High()
	time.Sleep(1 * time.Millisecond)
	r.Config.CsPin.Low()
	err := r.bus.Tx(tx, rx)
	if err != nil {
		fmt.Print(err)
	}
	r.Config.CsPin.High()
	fmt.Printf("read %x: %x\n", rx[1], tx)
	return rx[1], err
}

func (r *Device) setup() error {
	r.Config.CsPin.Low()
	fmt.Println("running initialization")
	Config := [][]byte{
		/* 0x01 */ {REG_OPMODE, RF_OPMODE_SEQUENCER_ON | RF_OPMODE_LISTEN_OFF | RF_OPMODE_STANDBY},
		/* 0x02 */ {REG_DATAMODUL, RF_DATAMODUL_DATAMODE_PACKET | RF_DATAMODUL_MODULATIONTYPE_FSK | RF_DATAMODUL_MODULATIONSHAPING_00}, // no shaping
		/* 0x03 */ {REG_BITRATEMSB, RF_BITRATELSB_250000}, // default: 4.8 KBPS
		/* 0x04 */ {REG_BITRATELSB, RF_BITRATELSB_250000},
		/* 0x05 */ {REG_FDEVMSB, RF_FDEVMSB_25000}, // default: 5KHz, (FDEV + BitRate / 2 <= 500KHz)
		/* 0x06 */ {REG_FDEVLSB, RF_FDEVLSB_25000},
		/* 0x07 */ {REG_FRFMSB, RF_FRFMSB_433},
		/* 0x08 */ {REG_FRFMID, RF_FRFMID_433},
		/* 0x09 */ {REG_FRFLSB, RF_FRFLSB_433},
		// looks like PA1 and PA2 are not implemented on RFM69W, hence the max output power is 13dBm
		// +17dBm and +20dBm are possible on RFM69HW
		// +13dBm formula: Pout = -18 + OutputPower (with PA0 or PA1**)
		// +17dBm formula: Pout = -14 + OutputPower (with PA1 and PA2)**
		// +20dBm formula: Pout = -11 + OutputPower (with PA1 and PA2)** and high power PA settings (section 3.3.7 in datasheet)
		///* 0x11 */ { REG_PALEVEL, RF_PALEVEL_PA0_ON | RF_PALEVEL_PA1_OFF | RF_PALEVEL_PA2_OFF | RF_PALEVEL_OUTPUTPOWER_11111},
		///* 0x13 */ { REG_OCP, RF_OCP_ON | RF_OCP_TRIM_95 }, // over current protection (default is 95mA)
		// RXBW defaults are { REG_RXBW, RF_RXBW_DCCFREQ_010 | RF_RXBW_MANT_24 | RF_RXBW_EXP_5} (RxBw: 10.4KHz)
		/* 0x19 */ {REG_RXBW, RF_RXBW_DCCFREQ_010 | RF_RXBW_MANT_24 | RF_RXBW_EXP_3},
		/* 0x25 */ {REG_DIOMAPPING1, RF_DIOMAPPING1_DIO0_01}, // DIO0 is the only IRQ we're using
		/* 0x26 */ {REG_DIOMAPPING2, RF_DIOMAPPING2_CLKOUT_OFF}, // DIO5 ClkOut disable for power saving
		/* 0x28 */ {REG_IRQFLAGS2, RF_IRQFLAGS2_FIFOOVERRUN}, // writing to this bit ensures that the FIFO & status flags are reset
		/* 0x29 */ //{REG_RSSITHRESH, 220}, // must be set to dBm = (-Sensitivity / 2), default is 0xE4 = 228 so -114dBm
		///* 0x2D */ { REG_PREAMBLELSB, RF_PREAMBLESIZE_LSB_VALUE } // default 3 preamble bytes 0xAAAAAA
		/* 0x2E */ {REG_SYNCCONFIG, RF_SYNC_ON | RF_SYNC_FIFOFILL_AUTO | RF_SYNC_SIZE_2 | RF_SYNC_TOL_0},
		/* 0x2F */ {REG_SYNCVALUE1, 0x2D}, // attempt to make this compatible with sync1 byte of RFM12B lib
		/* 0x30 */ {REG_SYNCVALUE2, r.Config.NetworkID}, // NETWORK ID
		/* 0x37 */ {REG_PACKETCONFIG1, RF_PACKET1_FORMAT_VARIABLE | RF_PACKET1_DCFREE_OFF | RF_PACKET1_CRC_ON | RF_PACKET1_CRCAUTOCLEAR_ON | RF_PACKET1_ADRSFILTERING_OFF},
		/* 0x38 */ {REG_PAYLOADLENGTH, 66}, // in variable length mode: the max frame size, not used in TX
		///* 0x39 */ { REG_NODEADRS, nodeID }, // turned off because we're not using address filtering
		/* 0x3C */ {REG_FIFOTHRESH, RF_FIFOTHRESH_TXSTART_FIFONOTEMPTY | RF_FIFOTHRESH_VALUE}, // TX on FIFO not empty
		/* 0x3D */ {REG_PACKETCONFIG2, RF_PACKET2_RXRESTARTDELAY_NONE | RF_PACKET2_AUTORXRESTART_ON | RF_PACKET2_AES_OFF}, // RXRESTARTDELAY must match transmitter PA ramp-down time (bitrate dependent)
		/* 0x6F */ {REG_TESTDAGC, RF_DAGC_IMPROVED_LOWBETA0}, // run DAGC continuously in RX mode for Fading Margin Improvement, recommended default for AfcLowBetaOn=0
	}
	fmt.Println("writing first sync value")
	for data, err := r.readReg(REG_SYNCVALUE1); err == nil && data != 0xAA; data, err = r.readReg(REG_SYNCVALUE1) {
		err := r.writeReg(REG_SYNCVALUE1, 0xAA)
		if err != nil {
			return err
		}
	}
	fmt.Println("writing second sync value")
	for data, err := r.readReg(REG_SYNCVALUE1); err == nil && data != 0x55; data, err = r.readReg(REG_SYNCVALUE1) {
		r.writeReg(REG_SYNCVALUE1, 0x55)
		if err != nil {
			return err
		}
	}
	for _, c := range Config {
		err := r.writeReg(c[0], c[1])
		if err != nil {
			return err
		}
	}
	err := r.Encrypt([]byte{})
	if err != nil {
		return err
	}
	err = r.setHighPower(r.Config.IsRfm69HCW)
	if err != nil {
		return err
	}
	err = r.SetMode(RF_OPMODE_STANDBY)
	if err != nil {
		return err
	}
	err = r.waitForMode()
	return err
}

func (r *Device) waitForMode() error {
	errChan := make(chan error)
	go func() {
		for {
			reg, err := r.readReg(REG_IRQFLAGS1)
			if err != nil {
				errChan <- err
				break
			}
			if reg&RF_IRQFLAGS1_MODEREADY != 0 {
				errChan <- nil
				break
			}
		}
	}()
	time.AfterFunc(5*time.Second, func() {
		errChan <- errors.New("timeout")
	})
	return <-errChan
}

// Encrypt sets the encryption key and enables AES encryption
func (r *Device) Encrypt(key []byte) error {
	var turnOn byte
	if len(key) == 16 {
		turnOn = 1
		tx := make([]byte, 17)
		tx[0] = REG_AESKEY1 | 0x80
		copy(tx[1:], key)
		rx := make([]byte, len(tx))
		err := r.bus.Tx(tx, rx)
		if err != nil {
			return err
		}
	}
	return r.readWriteReg(REG_PACKETCONFIG2, 0xFE, turnOn)
}

// SetMode sets operation mode
func (r *Device) SetMode(newMode byte) error {
	if newMode == r.mode {
		return nil
	}
	err := r.readWriteReg(REG_OPMODE, 0xE3, newMode)
	if err != nil {
		return err
	}
	if r.Config.IsRfm69HCW && (newMode == RF_OPMODE_RECEIVER || newMode == RF_OPMODE_TRANSMITTER) {
		err := r.setHighPowerRegs(newMode == RF_OPMODE_TRANSMITTER)
		if err != nil {
			return err
		}
	}
	if r.mode == RF_OPMODE_SLEEP {
		err = r.waitForMode()
		if err != nil {
			return err
		}
	}
	r.mode = newMode
	return nil
}

// SetModeAndWait sets the mode and waits for it
func (r *Device) SetModeAndWait(newMode byte) error {
	err := r.SetMode(RF_OPMODE_STANDBY)
	if err != nil {
		return err
	}
	err = r.waitForMode()
	if err != nil {
		return err
	}
	return nil
}

func (r *Device) setHighPower(turnOn bool) error {
	r.Config.IsRfm69HCW = turnOn
	ocp := byte(RF_OCP_ON)
	if r.Config.IsRfm69HCW {
		ocp = RF_OCP_OFF
	}
	err := r.writeReg(REG_OCP, ocp)
	if err != nil {
		return err
	}
	if r.Config.IsRfm69HCW {
		err = r.readWriteReg(REG_PALEVEL, 0x1F, RF_PALEVEL_PA1_ON|RF_PALEVEL_PA2_ON)
	} else {
		err = r.readWriteReg(REG_PALEVEL, 0, RF_PALEVEL_PA0_ON|RF_PALEVEL_PA1_OFF|RF_PALEVEL_PA2_OFF|r.powerLevel)
	}
	return err
}

func (r *Device) setHighPowerRegs(turnOn bool) (err error) {
	var (
		testPa1 byte = 0x55
		testPa2 byte = 0x70
	)

	if turnOn {
		testPa1 = 0x5D
		testPa2 = 0x7C
	}
	err = r.writeReg(REG_TESTPA1, testPa1)
	if err != nil {
		return
	}
	err = r.writeReg(REG_TESTPA2, testPa2)
	return
}

// SetNetwork sets the network ID
func (r *Device) SetNetwork(networkID byte) error {
	r.Config.NetworkID = networkID
	return r.writeReg(REG_SYNCVALUE2, networkID)
}

// SetAddress sets the node address
func (r *Device) SetAddress(address byte) error {
	r.Config.NodeID = address
	return r.writeReg(REG_NODEADRS, address)
}

// SetPowerLevel sets the TX power
func (r *Device) SetPowerLevel(powerLevel byte) error {
	r.powerLevel = powerLevel
	if r.powerLevel > 31 {
		r.powerLevel = 31
	}
	return r.readWriteReg(REG_PALEVEL, 0xE0, r.powerLevel)
}

func (r *Device) canSend() (bool, error) {
	// if signal stronger than -100dBm is detected assume channel activity
	if r.mode == RF_OPMODE_RECEIVER {
		rssi, err := r.readRSSI(false)
		if err != nil {
			return false, err
		}
		if rssi < CsmaLimit {
			err = r.SetMode(RF_OPMODE_STANDBY)
			return true, err
		}
	}
	return false, nil
}

func (r *Device) readRSSI(forceTrigger bool) (rssi int, err error) {
	if forceTrigger {
		// RSSI trigger not needed if DAGC is in continuous mode
		err = r.writeReg(REG_RSSICONFIG, RF_RSSI_START)
		if err != nil {
			return
		}
		for {
			data, err := r.readReg(REG_RSSICONFIG)
			if err != nil {
				return 0, err
			}
			if data&RF_RSSI_DONE != 0 {
				break
			}
		}
	}
	data, err := r.readReg(REG_RSSIVALUE)
	if err != nil {
		return
	}
	rssi = -int(data) / 2
	return
}

func (r *Device) readWriteReg(reg, andMask, orMask byte) error {
	regValue, err := r.readReg(reg)
	if err != nil {
		return err
	}
	regValue = (regValue & andMask) | orMask
	return r.writeReg(reg, regValue)
}

func (r *Device) writeFifo(data *Data) error {
	buffersize := len(data.Data)
	if buffersize > MaxDataLen {
		buffersize = MaxDataLen
	}
	tx := make([]byte, buffersize+5)
	// write to FIFO
	tx[0] = REG_FIFO | 0x80
	tx[1] = byte(buffersize + 3)
	tx[2] = data.ToAddress
	tx[3] = r.Config.NodeID
	if data.RequestAck {
		tx[4] = 0x40
	}
	if data.SendAck {
		tx[4] = 0x80
	}
	copy(tx[5:], data.Data[:buffersize])
	rx := make([]byte, len(tx))
	err := r.bus.Tx(tx, rx)
	return err
}

func (r *Device) readFifo() (Data, error) {
	var err error
	data := Data{}
	data.Rssi, err = r.readRSSI(false)
	if err != nil {
		return data, err
	}
	tx := new([70]byte)
	tx[0] = REG_FIFO & 0x7f
	rx := make([]byte, len(tx[:3]))
	err = r.bus.Tx(tx[:3], rx)
	if err != nil {
		return data, err
	}
	data.ToAddress = rx[2]
	length := rx[1] - 3
	if length > 66 {
		length = 66
	}
	rx = make([]byte, len(tx[:length+3]))
	err = r.bus.Tx(tx[:length+3], rx)
	if err != nil {
		return data, err
	}
	data.FromAddress = rx[1]
	data.SendAck = bool(rx[2]&0x80 > 0)
	data.RequestAck = bool(rx[2]&0x40 > 0)
	data.Data = rx[3:]
	return data, nil
}
