// Package sx126x provides a driver for SX126x LoRa transceivers.
// Inspired from https://github.com/Lora-net/sx126x_driver/

package sx126x

import (
	"errors"
	"time"

	"machine"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/lora"
)

var (
	errWaitWhileBusyTimeout   = errors.New("WaitWhileBusy Timeout")
	errLowPowerTxNotSupported = errors.New("RFSWITCH_TX_LP not supported")
	errRadioNotFound          = errors.New("LoRa radio not found")
	errUnexpectedRxRadioEvent = errors.New("Unexpected Radio Event during RX")
	errUnexpectedTxRadioEvent = errors.New("Unexpected Radio Event during TX")
)

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
	PERIOD_PER_SEC      = (uint32)(1000000 / 15.625) // SX1261 DS 13.1.4
	SPI_BUFFER_SIZE     = 256
	RADIOEVENTCHAN_SIZE = 1
)

// Device wraps an SPI connection to a SX126x device.
type Device struct {
	spi            drivers.SPI          // SPI bus for module communication
	rstPin         machine.Pin          // GPIO for reset pin
	radioEventChan chan lora.RadioEvent // Channel for Receiving events
	loraConf       lora.Config          // Current Lora configuration
	controller     RadioController      // to manage interactions with the radio
	deepSleep      bool                 // Internal Sleep state
	deviceType     int                  // sx1261,sx1262,sx1268 (defaults sx1261)
	spiTxBuf       []byte               // global Tx buffer to avoid heap allocations in interrupt
	spiRxBuf       []byte               // global Rx buffer to avoid heap allocations in interrupt

}

// New creates a new SX126x connection.
func New(spi drivers.SPI) *Device {
	return &Device{
		spi:            spi,
		radioEventChan: make(chan lora.RadioEvent, RADIOEVENTCHAN_SIZE),
		spiTxBuf:       make([]byte, SPI_BUFFER_SIZE),
		spiRxBuf:       make([]byte, SPI_BUFFER_SIZE),
	}
}

const (
	SX126X_RTC_FREQ_IN_HZ uint32 = 64000
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
// Get the RadioEvent channel of the device
func (d *Device) GetRadioEventChan() chan lora.RadioEvent {
	return d.radioEventChan
}

// Specify device type (SX1261/2/8)
func (d *Device) SetDeviceType(devType int) {
	d.deviceType = devType
}

// SetRadioControl let you define the RadioController
func (d *Device) SetRadioController(rc RadioController) error {
	d.controller = rc
	if err := d.controller.Init(); err != nil {
		return err
	}
	d.controller.SetupInterrupts(d.HandleInterrupt)

	return nil
}

// --------------------------------------------------
// Operational modes functions
// --------------------------------------------------

func (d *Device) Reset() {
	d.rstPin.Low()
	time.Sleep(100 * time.Millisecond)
	d.rstPin.High()
	time.Sleep(100 * time.Millisecond)
}

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
	if d.controller != nil {
		d.controller.SetRfSwitchMode(RFSWITCH_TX_HP)
	}
	d.ExecSetCommand(SX126X_CMD_SET_TX_CONTINUOUS_WAVE, []uint8{})
}

// SetTxContinuousPreamble set device in test mode to constantly modulate LoRa preamble symbols.
// Take care to initialize all Lora settings like it's done in Tx before calling this function
// If you don't init properly all the settings, it'll fail
func (d *Device) SetTxContinuousPreamble() {
	if d.controller != nil {
		d.controller.SetRfSwitchMode(RFSWITCH_TX_HP)
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
	d.controller.SetNss(false)
	// Send command
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, SX126X_CMD_READ_REGISTER, uint8((addr&0xFF00)>>8), uint8(addr&0x00FF), 0x00)
	d.spi.Tx(d.spiTxBuf, nil)
	// Read registers
	d.spiRxBuf = d.spiRxBuf[0:size]
	d.spi.Tx(nil, d.spiRxBuf)
	d.controller.SetNss(true)
	d.controller.WaitWhileBusy()
	return d.spiRxBuf, nil
}

// WriteRegister writes value to register
func (d *Device) WriteRegister(addr uint16, data []uint8) {
	d.CheckDeviceReady()
	d.controller.SetNss(false)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, SX126X_CMD_WRITE_REGISTER, uint8((addr&0xFF00)>>8), uint8(addr&0x00FF))
	d.spiTxBuf = append(d.spiTxBuf, data...)
	d.spi.Tx(d.spiTxBuf, nil)
	d.controller.SetNss(true)
	d.controller.WaitWhileBusy()
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
func (d *Device) SetSyncWord(sw uint16) {
	var p [2]uint8
	d.loraConf.SyncWord = sw
	p[0] = uint8((d.loraConf.SyncWord >> 8) & 0xFF)
	p[1] = uint8((d.loraConf.SyncWord >> 0) & 0xFF)
	d.WriteRegister(SX126X_REG_LORA_SYNC_WORD_MSB, p[:])
}

// GetSyncWord gets the Sync Word to use
func (d *Device) GetSyncWord() uint16 {
	p, _ := d.ReadRegister(SX126X_REG_LORA_SYNC_WORD_MSB, 2)
	r := uint16(p[0])<<8 + uint16(p[1])
	return r
}

// SetPublicNetwork sets Sync Word to 0x3444 (Public) or 0x1424 (Private)
func (d *Device) SetPublicNetwork(enable bool) {
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
		d.controller.SetNss(false)
		time.Sleep(time.Millisecond)
		d.controller.SetNss(true)
		d.deepSleep = false
	}
	return d.controller.WaitWhileBusy()
}

// ExecSetCommand send a command to configure the peripheral
func (d *Device) ExecSetCommand(cmd uint8, buf []uint8) {
	d.CheckDeviceReady()
	if cmd == SX126X_CMD_SET_SLEEP {
		d.deepSleep = true
	} else {
		d.deepSleep = false
	}
	d.controller.SetNss(false)
	// Send command and params
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmd)
	d.spiTxBuf = append(d.spiTxBuf, buf...)
	d.spi.Tx(d.spiTxBuf, nil)
	d.controller.SetNss(true)
	if cmd != SX126X_CMD_SET_SLEEP {
		d.controller.WaitWhileBusy()
	}
}

// ExecGetCommand queries the peripheral the peripheral
func (d *Device) ExecGetCommand(cmd uint8, size uint8) []uint8 {
	d.CheckDeviceReady()
	d.controller.SetNss(false)
	// Send the command and flush first status byte (as not used)
	d.spiTxBuf = d.spiTxBuf[:0]
	d.spiTxBuf = append(d.spiTxBuf, cmd, 0x00)
	d.spi.Tx(d.spiTxBuf, nil)
	// Read resp
	d.spiRxBuf = d.spiRxBuf[:size]
	d.spi.Tx(nil, d.spiRxBuf)
	d.controller.SetNss(true)
	d.controller.WaitWhileBusy()
	return d.spiRxBuf
}

//
// Configuration
//

// SetFrequency() Sets current Lora Frequency
// NB: Change will be applied at next RX / TX
func (d *Device) SetFrequency(freq uint32) {
	d.loraConf.Freq = freq
}

// SetIqMode() defines the current IQ Mode (Standard/Inverted)
// NB: Change will be applied at next RX / TX
func (d *Device) SetIqMode(mode uint8) {
	if mode == 0 {
		d.loraConf.Iq = lora.IQStandard
	} else {
		d.loraConf.Iq = lora.IQInverted
	}
}

// SetCodingRate() sets current Lora Coding Rate
// NB: Change will be applied at next RX / TX
func (d *Device) SetCodingRate(cr uint8) {
	d.loraConf.Cr = cr
}

// SetBandwidth() sets current Lora Bandwidth
// NB: Change will be applied at next RX / TX
func (d *Device) SetBandwidth(bw uint8) {
	d.loraConf.Bw = bw
}

// SetCrc() sets current CRC mode (ON/OFF)
// NB: Change will be applied at next RX / TX
func (d *Device) SetCrc(enable bool) {
	if enable {
		d.loraConf.Crc = lora.CRCOn
	} else {
		d.loraConf.Crc = lora.CRCOn
	}
}

// SetSpreadingFactor sets current Lora Spreading Factor
// NB: Change will be applied at next RX / TX
func (d *Device) SetSpreadingFactor(sf uint8) {
	d.loraConf.Sf = sf
}

// SetPreambleLength sets current Lora Preamble Length
// NB: Change will be applied at next RX / TX
func (d *Device) SetPreambleLength(pl uint16) {
	d.loraConf.Preamble = pl
}

// SetTxPowerDbm sets current Lora TX Power in DBm
// NB: Change will be applied at next RX / TX
func (d *Device) SetTxPower(txpow int8) {
	d.loraConf.LoraTxPowerDBm = txpow
}

// SetHeaderType sets implicit or explicit header mode
// NB: Change will be applied at next RX / TX
func (d *Device) SetHeaderType(headerType uint8) {
	d.loraConf.HeaderType = headerType
}

//
// Lora functions
//
//

// LoraConfig() defines Lora configuration for next Lora operations
func (d *Device) LoraConfig(cnf lora.Config) {
	// Save given configuration
	d.loraConf = cnf
	d.loraConf.SyncWord = syncword(int(cnf.SyncWord))
	// Switch to standby prior to configuration changes
	d.SetStandby()
	// Clear errors, disable radio interrupts for the moment
	d.ClearDeviceErrors()
	d.ClearIrqStatus(SX126X_IRQ_ALL)
	d.SetDioIrqParams(0x00, 0x00, 0x00, 0x00)
	// Define radio operation mode
	d.SetPacketType(SX126X_PACKET_TYPE_LORA)
	d.SetRfFrequency(d.loraConf.Freq)
	d.SetModulationParams(d.loraConf.Sf, bandwidth(d.loraConf.Bw), d.loraConf.Cr, d.loraConf.Ldr)
	d.SetTxParams(d.loraConf.LoraTxPowerDBm, SX126X_PA_RAMP_200U)
	d.SetSyncWord(d.loraConf.SyncWord)
	d.SetBufferBaseAddress(0, 0)
}

// Tx sends a lora packet, (with timeout)
func (d *Device) Tx(pkt []uint8, timeoutMs uint32) error {
	if d.loraConf.Freq == 0 {
		return lora.ErrUndefinedLoraConf
	}

	if d.controller != nil {
		err := d.controller.SetRfSwitchMode(RFSWITCH_TX_HP)
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
	d.SetModulationParams(d.loraConf.Sf, bandwidth(d.loraConf.Bw), d.loraConf.Cr, d.loraConf.Ldr)
	d.SetPacketParam(d.loraConf.Preamble, d.loraConf.HeaderType, d.loraConf.Crc, uint8(len(pkt)), d.loraConf.Iq)
	d.SetDioIrqParams(irqVal, irqVal, SX126X_IRQ_NONE, SX126X_IRQ_NONE)
	d.SetSyncWord(d.loraConf.SyncWord)
	d.SetTx(timeoutMsToRtcSteps(timeoutMs))

	msg := <-d.GetRadioEventChan()
	if msg.EventType != lora.RadioEventTxDone {
		return errUnexpectedTxRadioEvent
	}
	return nil
}

// LoraRx tries to receive a Lora packet (with timeout in milliseconds)
func (d *Device) Rx(timeoutMs uint32) ([]uint8, error) {
	if d.loraConf.Freq == 0 {
		return nil, lora.ErrUndefinedLoraConf
	}

	if d.controller != nil {
		err := d.controller.SetRfSwitchMode(RFSWITCH_RX)
		if err != nil {
			return nil, err
		}
	}

	d.ClearIrqStatus(SX126X_IRQ_ALL)
	irqVal := uint16(SX126X_IRQ_RX_DONE | SX126X_IRQ_TIMEOUT | SX126X_IRQ_CRC_ERR)
	d.SetStandby()
	d.SetBufferBaseAddress(0, 0)
	d.SetRfFrequency(d.loraConf.Freq)
	d.SetModulationParams(d.loraConf.Sf, bandwidth(d.loraConf.Bw), d.loraConf.Cr, d.loraConf.Ldr)
	d.SetPacketParam(d.loraConf.Preamble, d.loraConf.HeaderType, d.loraConf.Crc, 0xFF, d.loraConf.Iq)
	d.SetDioIrqParams(irqVal, irqVal, SX126X_IRQ_NONE, SX126X_IRQ_NONE)
	d.SetRx(timeoutMsToRtcSteps(timeoutMs))

	msg := <-d.GetRadioEventChan()

	if msg.EventType == lora.RadioEventTimeout {
		return nil, nil
	} else if msg.EventType != lora.RadioEventRxDone {
		return nil, errUnexpectedRxRadioEvent
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

	if (st & SX126X_IRQ_RX_DONE) > 0 {
		select {
		case d.radioEventChan <- lora.RadioEvent{lora.RadioEventRxDone, uint16(st), nil}:
		default:
		}
	}

	if (st & SX126X_IRQ_TX_DONE) > 0 {
		select {
		case d.radioEventChan <- lora.RadioEvent{lora.RadioEventTxDone, uint16(st), nil}:
		default:
		}
	}

	if (st & SX126X_IRQ_TIMEOUT) > 0 {
		select {
		case d.radioEventChan <- lora.RadioEvent{lora.RadioEventTimeout, uint16(st), nil}:
		default:
		}

	}

	if (st & SX126X_IRQ_CRC_ERR) > 0 {
		select {
		case d.radioEventChan <- lora.RadioEvent{lora.RadioEventCrcError, uint16(st), nil}:

		default:
		}
	}

}

func bandwidth(bw uint8) uint8 {
	switch bw {
	case lora.Bandwidth_7_8:
		return SX126X_LORA_BW_7_8
	case lora.Bandwidth_10_4:
		return SX126X_LORA_BW_10_4
	case lora.Bandwidth_15_6:
		return SX126X_LORA_BW_15_6
	case lora.Bandwidth_20_8:
		return SX126X_LORA_BW_20_8
	case lora.Bandwidth_31_25:
		return SX126X_LORA_BW_31_25
	case lora.Bandwidth_41_7:
		return SX126X_LORA_BW_41_7
	case lora.Bandwidth_62_5:
		return SX126X_LORA_BW_62_5
	case lora.Bandwidth_125_0:
		return SX126X_LORA_BW_125_0
	case lora.Bandwidth_250_0:
		return SX126X_LORA_BW_250_0
	case lora.Bandwidth_500_0:
		return SX126X_LORA_BW_500_0
	default:
		return 0
	}
}

func syncword(sw int) uint16 {
	if sw == lora.SyncPublic {
		return SX126X_LORA_MAC_PUBLIC_SYNCWORD
	}
	return SX126X_LORA_MAC_PRIVATE_SYNCWORD
}
