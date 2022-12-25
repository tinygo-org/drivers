// Package sx127x provides a driver for SX127x LoRa transceivers.
// References:
// https://electronics.stackexchange.com/questions/394296/can-t-get-simple-lora-receiver-to-work
// https://www.st.com/resource/en/user_manual/dm00300436-stm32-lora-expansion-package-for-stm32cube-stmicroelectronics.pdf
package sx127x

import (
	"errors"
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/lora"
)

// So we can keep track of the origin of interruption
const (
	SPI_BUFFER_SIZE = 256
)

// Device wraps an SPI connection to a SX127x device.
type Device struct {
	spi            drivers.SPI          // SPI bus for module communication
	rstPin, csPin  machine.Pin          // GPIOs for reset and chip select
	radioEventChan chan lora.RadioEvent // Channel for Receiving events
	loraConf       lora.Config          // Current Lora configuration
	deepSleep      bool                 // Internal Sleep state
	deviceType     int                  // sx1261,sx1262,sx1268 (defaults sx1261)
	spiBuffer      [SPI_BUFFER_SIZE]uint8
	packetIndex    uint8 // FIXME ... useless ?
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

// New creates a new SX127x connection. The SPI bus must already be configured.
func New(spi machine.SPI, csPin machine.Pin, rstPin machine.Pin) *Device {
	k := Device{
		spi:            spi,
		csPin:          csPin,
		rstPin:         rstPin,
		radioEventChan: make(chan lora.RadioEvent, 10),
	}
	return &k
}

// Reset re-initialize the sx127x device
func (d *Device) Reset() {
	d.rstPin.Low()
	time.Sleep(100 * time.Millisecond)
	d.rstPin.High()
	time.Sleep(100 * time.Millisecond)
}

// DetectDevice checks if device responds on the SPI bus
func (d *Device) DetectDevice() bool {
	id := d.GetVersion()
	return (id == 0x12)
}

// ReadRegister reads register value
func (d *Device) ReadRegister(reg uint8) uint8 {
	d.csPin.Low()
	d.spi.Tx([]byte{reg & 0x7f}, nil)
	var value [1]byte
	d.spi.Tx(nil, value[:])
	d.csPin.High()
	return value[0]
}

// WriteRegister writes value to register
func (d *Device) WriteRegister(reg uint8, value uint8) uint8 {
	var response [1]byte
	d.csPin.Low()
	d.spi.Tx([]byte{reg | 0x80}, nil)
	d.spi.Tx([]byte{value}, response[:])
	d.csPin.High()
	return response[0]
}

// SetOpMode changes the sx1276 mode
func (d *Device) SetOpMode(mode uint8) {
	cur := d.ReadRegister(SX127X_REG_OP_MODE)
	new := (cur & (^SX127X_OPMODE_MASK)) | mode
	d.WriteRegister(SX127X_REG_OP_MODE, new)
}

// SetOpMode changes the sx1276 mode
func (d *Device) SetOpModeLora() {
	d.WriteRegister(SX127X_REG_OP_MODE, SX127X_OPMODE_LORA)
}

// GetVersion returns hardware version of sx1276 chipset
func (d *Device) GetVersion() uint8 {
	return (d.ReadRegister(SX127X_REG_VERSION))
}

// IsTransmitting tests if a packet transmission is in progress
func (d *Device) IsTransmitting() bool {
	return (d.ReadRegister(SX127X_REG_OP_MODE) & SX127X_OPMODE_TX) == SX127X_OPMODE_TX
}

// LastPacketRSSI gives the RSSI of the last packet received
func (d *Device) LastPacketRSSI() uint8 {
	// section 5.5.5
	var adjustValue uint8 = 157
	if d.loraConf.Freq < 868000000 {
		adjustValue = 164
	}
	return d.ReadRegister(SX127X_REG_PKT_RSSI_VALUE) - adjustValue
}

// LastPacketSNR gives the SNR of the last packet received
func (d *Device) LastPacketSNR() uint8 {
	return uint8(d.ReadRegister(SX127X_REG_PKT_SNR_VALUE) / 4)
}

// GetRSSI returns current RSSI
func (d *Device) GetRSSI() uint8 {
	return d.ReadRegister(SX127X_REG_RSSI_VALUE)
}

/*
// GetBandwidth returns the bandwidth the LoRa module is using
func (d *Device) GetBandwidth() int32 {
	return int32(d.loraConf.Bw)
}
*/

// SetTxPower sets the transmitter output power
func (d *Device) SetTxPower(txPower int8, paBoost bool) {
	if !paBoost {
		// RFO
		if txPower < 0 {
			txPower = 0
		} else if txPower > 14 {
			txPower = 14
		}
		d.WriteRegister(SX127X_REG_PA_CONFIG, uint8(0x70)|uint8(txPower))

	} else {
		//PA_BOOST
		if txPower > 17 {
			if txPower > 20 {
				txPower = 20
			}

			txPower -= 3

			// High Power +20 dBm Operation (Semtech SX1276/77/78/79 5.4.3.)
			d.WriteRegister(SX127X_REG_PA_DAC, 0x87)
			d.SetOCP(140)
		} else {
			if txPower < 2 {
				txPower = 2
			}

			d.WriteRegister(SX127X_REG_PA_DAC, 0x84)
			d.SetOCP(100)

		}

		d.WriteRegister(SX127X_REG_PA_CONFIG, uint8(SX127X_PA_BOOST)|uint8(txPower-2))

	}
}

// ---------------
// Internal functions
// ---------------

// SetRxTimeout defines RX Timeout expressed as number of symbols
// Default timeout is 64 * Ts
func (d *Device) SetRxTimeout(tmoutSymb uint8) {
	d.WriteRegister(SX127X_REG_SYMB_TIMEOUT_LSB, tmoutSymb)
}

// SetOCP defines Overload Current Protection configuration
func (d *Device) SetOCP(mA uint8) {
	ocpTrim := uint8(27)
	if mA < 45 {
		mA = 45
	}
	if mA <= 120 {
		ocpTrim = (mA - 45) / 5
	} else if mA <= 240 {
		ocpTrim = (mA + 30) / 10
	}
	d.WriteRegister(SX127X_REG_OCP, 0x20|(0x1F&ocpTrim))
}

// SetAgcAutoOn enables Automatic Gain Control
func (d *Device) SetAgcAuto(val uint8) {
	if val == SX127X_AGC_AUTO_ON {
		d.WriteRegister(SX127X_REG_MODEM_CONFIG_3, d.ReadRegister(SX127X_REG_MODEM_CONFIG_3)|0x04)
	} else {
		d.WriteRegister(SX127X_REG_MODEM_CONFIG_3, d.ReadRegister(SX127X_REG_MODEM_CONFIG_3)&0xfb)
	}
}

// SetLowDataRateOptimize enables Low Data Rate Optimization
func (d *Device) SetLowDataRateOptim(val uint8) {
	if val == lora.LowDataRateOptimizeOn {
		d.WriteRegister(SX127X_REG_MODEM_CONFIG_3, d.ReadRegister(SX127X_REG_MODEM_CONFIG_3)|0x08)
	} else {
		d.WriteRegister(SX127X_REG_MODEM_CONFIG_3, d.ReadRegister(SX127X_REG_MODEM_CONFIG_3)&0xf7)
	}
}

// SetLowFrequencyModeOn enables Low Data Rate Optimization
func (d *Device) SetLowFrequencyModeOn(val bool) {
	if val {
		d.WriteRegister(SX127X_REG_OP_MODE, d.ReadRegister(SX127X_REG_OP_MODE)|0x04)
	} else {
		d.WriteRegister(SX127X_REG_OP_MODE, d.ReadRegister(SX127X_REG_OP_MODE)&0xfb)
	}
}

// SetHopPeriod sets number of symbol periods between frequency hops. (0 = disabled).
func (d *Device) SetHopPeriod(val uint8) {
	d.WriteRegister(SX127X_REG_HOP_PERIOD, val)
}

//
// LORA FUNCTIONS
//

// LoraConfig() defines Lora configuration for next Lora operations
func (d *Device) LoraConfig(cnf lora.Config) {
	// Save given configuration
	d.loraConf = cnf
	d.loraConf.SyncWord = syncword(int(cnf.SyncWord))
}

// SetLoraFrequency updates the frequency the LoRa module is using
func (d *Device) SetLoraFrequency(frequency uint32) {
	d.loraConf.Freq = frequency
	var frf = (uint64(frequency) << 19) / 32000000
	d.WriteRegister(SX127X_REG_FRF_MSB, uint8(frf>>16))
	d.WriteRegister(SX127X_REG_FRF_MID, uint8(frf>>8))
	d.WriteRegister(SX127X_REG_FRF_LSB, uint8(frf>>0))
}

// SetBandwidth updates the bandwidth the LoRa module is using
func (d *Device) SetLoraBandwidth(bw uint8) {
	d.loraConf.Bw = bandwidth(bw)
	d.WriteRegister(SX127X_REG_MODEM_CONFIG_1, (d.ReadRegister(SX127X_REG_MODEM_CONFIG_1)&0x0f)|(bw<<4))
}

// SetCodingRate updates the coding rate the LoRa module is using
func (d *Device) SetLoraCodingRate(cr uint8) {
	d.loraConf.Cr = cr
	d.WriteRegister(SX127X_REG_MODEM_CONFIG_1, (d.ReadRegister(SX127X_REG_MODEM_CONFIG_1)&0xf1)|(cr<<1))
}

// SetImplicitHeaderModeOn Enables implicit header mode ***
func (d *Device) SetLoraHeaderMode(headerType uint8) {
	d.loraConf.HeaderType = headerType
	if headerType == lora.HeaderExplicit {
		d.WriteRegister(SX127X_REG_MODEM_CONFIG_1, d.ReadRegister(SX127X_REG_MODEM_CONFIG_1)|0x01)
	} else {
		d.WriteRegister(SX127X_REG_MODEM_CONFIG_1, d.ReadRegister(SX127X_REG_MODEM_CONFIG_1)&0xfe)
	}
}

// SetLoraSpreadingFactor changes spreading factor
func (d *Device) SetLoraSpreadingFactor(sf uint8) {
	d.loraConf.Sf = sf
	if sf == lora.SpreadingFactor6 {
		d.WriteRegister(SX127X_REG_DETECTION_OPTIMIZE, 0xc5)
		d.WriteRegister(SX127X_REG_DETECTION_THRESHOLD, 0x0c)
	} else {
		d.WriteRegister(SX127X_REG_DETECTION_OPTIMIZE, 0xc3)
		d.WriteRegister(SX127X_REG_DETECTION_THRESHOLD, 0x0a)
	}
	var newValue = (d.ReadRegister(SX127X_REG_MODEM_CONFIG_2) & 0x0f) | ((sf << 4) & 0xf0)
	d.WriteRegister(SX127X_REG_MODEM_CONFIG_2, newValue)
}

// SetTxContinuousMode enable Continuous Tx mode
func (d *Device) SetTxContinuousMode(val bool) {
	if val {
		d.WriteRegister(SX127X_REG_MODEM_CONFIG_2, d.ReadRegister(SX127X_REG_MODEM_CONFIG_2)|0x08)
	} else {
		d.WriteRegister(SX127X_REG_MODEM_CONFIG_2, d.ReadRegister(SX127X_REG_MODEM_CONFIG_2)&0xf7)
	}
}

// SetLoraCrc Enable CRC generation and check on payload
func (d *Device) SetLoraCrc(enable bool) {
	if enable {
		d.loraConf.Crc = lora.CRCOn
		d.WriteRegister(SX127X_REG_MODEM_CONFIG_2, d.ReadRegister(SX127X_REG_MODEM_CONFIG_2)|0x04)
	} else {
		d.loraConf.Crc = lora.CRCOff
		d.WriteRegister(SX127X_REG_MODEM_CONFIG_2, d.ReadRegister(SX127X_REG_MODEM_CONFIG_2)&0xfb)
	}
}

func (d *Device) SetLoraPreamble(pLen uint16) {
	// Sets preamble length
	d.WriteRegister(SX127X_REG_PREAMBLE_MSB, uint8((pLen>>8)&0xFF))
	d.WriteRegister(SX127X_REG_PREAMBLE_LSB, uint8(pLen&0xFF))
}

// SetLoraSyncWord defines sync word
func (d *Device) SetLoraSyncWord(syncWord uint16) {
	d.loraConf.SyncWord = syncWord
	sw := uint8(syncWord & 0xFF)
	d.WriteRegister(SX127X_REG_SYNC_WORD, sw)
}

// SetLoraIQMode Sets I/Q polarity configuration
func (d *Device) SetLoraIqMode(val uint8) {
	d.loraConf.Iq = val
	if val == lora.IQStandard {
		//Set IQ to normal values
		d.WriteRegister(SX127X_REG_INVERTIQ, 0x27)
		d.WriteRegister(SX127X_REG_INVERTIQ2, 0x1D)
	} else {
		//Invert IQ Back
		d.WriteRegister(SX127X_REG_INVERTIQ, 0x66)
		d.WriteRegister(SX127X_REG_INVERTIQ2, 0x19)
	}
}

// LoraTx sends a lora packet, (with timeout)
func (d *Device) LoraTx(pkt []uint8, timeoutMs uint32) error {

	//println("sx127x: LoraTx:", len(pkt), " bytes", hex.EncodeToString(pkt))
	d.SetOpModeLora()
	d.SetOpMode(SX127X_OPMODE_SLEEP)

	d.SetHopPeriod(0x00)
	d.SetLowFrequencyModeOn(false)                                                      // High freq mode
	d.WriteRegister(SX127X_REG_PA_RAMP, (d.ReadRegister(SX127X_REG_PA_RAMP)&0xF0)|0x08) // set PA ramp-up time 50 uSec
	d.WriteRegister(SX127X_REG_LNA, SX127X_LNA_MAX_GAIN)                                // Set Low Noise Amplifier to MAX

	d.SetLoraFrequency(d.loraConf.Freq)
	d.SetLoraPreamble(d.loraConf.Preamble)
	d.SetLoraSyncWord(d.loraConf.SyncWord)
	d.SetLoraBandwidth(d.loraConf.Bw)
	d.SetLoraSpreadingFactor(d.loraConf.Sf)
	d.SetLoraIqMode(d.loraConf.Iq)
	d.SetLoraCodingRate(d.loraConf.Cr)
	d.SetLoraCrc(d.loraConf.Crc == lora.CRCOn)
	d.SetTxPower(d.loraConf.LoraTxPowerDBm, true)
	d.SetLoraHeaderMode(d.loraConf.HeaderType)
	d.SetAgcAuto(SX127X_AGC_AUTO_ON)

	// set the IRQ mapping DIO0=TxDone DIO1=NOP DIO2=NOP
	d.WriteRegister(SX127X_REG_DIO_MAPPING_1, SX127X_MAP_DIO0_LORA_TXDONE|SX127X_MAP_DIO1_LORA_NOP|SX127X_MAP_DIO2_LORA_NOP)
	// Clear all radio IRQ Flags
	d.WriteRegister(SX127X_REG_IRQ_FLAGS, 0xFF)
	// Mask all but TxDone
	d.WriteRegister(SX127X_REG_IRQ_FLAGS_MASK, ^SX127X_IRQ_LORA_TXDONE_MASK)

	// initialize the payload size and address pointers
	d.WriteRegister(SX127X_REG_PAYLOAD_LENGTH, uint8(len(pkt)))
	d.WriteRegister(SX127X_REG_FIFO_TX_BASE_ADDR, 0)
	d.WriteRegister(SX127X_REG_FIFO_ADDR_PTR, 0)

	// FIFO OPs cannot take place in Sleep mode !!!
	d.SetOpMode(SX127X_OPMODE_STANDBY)
	time.Sleep(time.Millisecond)
	// Copy payload to FIFO // TODO: Bulk
	for i := 0; i < len(pkt); i++ {
		d.WriteRegister(SX127X_REG_FIFO, pkt[i])
	}

	// Enable TX
	d.SetOpMode(SX127X_OPMODE_TX)

	msg := <-d.GetRadioEventChan()
	if msg.EventType != lora.RadioEventTxDone {
		return errors.New("Unexpected Radio Event while TX " + string(0x30+msg.EventType))
	}
	return nil
}

// LoraRx tries to receive a Lora packet (with timeout in milliseconds)
func (d *Device) LoraRx(timeoutMs uint32) ([]uint8, error) {

	if d.loraConf.Freq == 0 {
		return nil, lora.ErrUndefinedLoraConf
	}

	d.SetOpModeLora()
	d.SetOpMode(SX127X_OPMODE_SLEEP)

	d.SetHopPeriod(0x00)
	d.SetLowFrequencyModeOn(false)                                                      // High freq mode
	d.WriteRegister(SX127X_REG_PA_RAMP, (d.ReadRegister(SX127X_REG_PA_RAMP)&0xF0)|0x08) // set PA ramp-up time 50 uSec
	d.WriteRegister(SX127X_REG_LNA, SX127X_LNA_MAX_GAIN)                                // Set Low Noise Amplifier to MAX

	d.SetLoraFrequency(d.loraConf.Freq)
	d.SetLoraPreamble(d.loraConf.Preamble)  //OK
	d.SetLoraSyncWord(d.loraConf.SyncWord)  // Should be ok
	d.SetLoraBandwidth(d.loraConf.Bw)       // OK
	d.SetLoraSpreadingFactor(d.loraConf.Sf) // OK
	d.SetLoraIqMode(d.loraConf.Iq)          //OK
	d.SetLoraCodingRate(d.loraConf.Cr)
	d.SetLoraCrc(d.loraConf.Crc == lora.CRCOn)
	d.SetTxPower(d.loraConf.LoraTxPowerDBm, true)
	d.SetLoraHeaderMode(d.loraConf.HeaderType)
	d.SetAgcAuto(SX127X_AGC_AUTO_ON)

	// set the IRQ mapping DIO0=RxDone DIO1=RxTimeout DIO2=NOP
	d.WriteRegister(SX127X_REG_DIO_MAPPING_1, SX127X_MAP_DIO0_LORA_RXDONE|SX127X_MAP_DIO1_LORA_RXTOUT|SX127X_MAP_DIO2_LORA_NOP)
	// Clear all radio IRQ Flags
	d.WriteRegister(SX127X_REG_IRQ_FLAGS, 0xFF)
	// Mask all but RxDone
	d.WriteRegister(SX127X_REG_IRQ_FLAGS_MASK, ^(SX127X_IRQ_LORA_RXDONE_MASK | SX127X_IRQ_LORA_RXTOUT_MASK))
	// Switch to RX Mode
	d.SetOpMode(SX127X_OPMODE_RX_SINGLE) //
	// Wait for Radio Event

	radioCh := d.GetRadioEventChan()

	msg := <-radioCh
	if msg.EventType == lora.RadioEventTimeout {
		return nil, nil
	} else if msg.EventType != lora.RadioEventRxDone {
		return nil, errors.New("Unexpected Radio Event while RX " + string(0x30+msg.EventType))
	}

	d.WriteRegister(SX127X_REG_FIFO_RX_BASE_ADDR, 0)
	d.WriteRegister(SX127X_REG_FIFO_ADDR_PTR, 0)

	pLen := d.ReadRegister(SX127X_REG_RX_NB_BYTES)
	d.WriteRegister(SX127X_REG_FIFO_ADDR_PTR, d.ReadRegister(SX127X_REG_FIFO_RX_CURRENT_ADDR))

	for i := uint8(0); i < pLen; i++ {
		d.spiBuffer[i] = d.ReadRegister(SX127X_REG_FIFO)
	}
	return d.spiBuffer[:pLen], nil
}

//
// HELPER FUNCTIONS
//

// PrintRegisters outputs the sx127x transceiver registers
func (d *Device) PrintRegisters(compact bool) {
	for i := uint8(0); i < 128; i++ {
		v := d.ReadRegister(i)
		print(v, " ")
	}
	println()
}

// PrintRegisters outputs the sx127x transceiver registers
func (d *Device) RandomU32() uint32 {
	// Disable ALL irqs
	d.WriteRegister(SX127X_REG_IRQ_FLAGS, 0xFF)
	d.SetOpModeLora()
	d.SetOpMode(SX127X_OPMODE_SLEEP)
	d.SetLoraFrequency(d.loraConf.Freq)
	d.SetOpMode(SX127X_OPMODE_RX)
	rnd := uint32(0)
	for i := 0; i < 32; i++ {
		time.Sleep(time.Millisecond * 10)
		// Unfiltered RSSI value reading. Only takes the LSB value
		rnd |= (uint32(d.ReadRegister(SX127X_REG_RSSI_WIDEBAND)) & 0x01) << i
	}
	return rnd
}

// HandleInterrupt must be called by main code on DIO state change.
func (d *Device) HandleInterrupt() {
	// Get IRQ and clear
	st := d.ReadRegister(SX127X_REG_IRQ_FLAGS)
	d.WriteRegister(SX127X_REG_IRQ_FLAGS, 0xFF)

	rChan := d.GetRadioEventChan()

	if (st & SX127X_IRQ_LORA_RXDONE_MASK) > 0 {
		rChan <- lora.NewRadioEvent(lora.RadioEventRxDone, uint16(st), nil)
	}

	if (st & SX127X_IRQ_LORA_TXDONE_MASK) > 0 {
		rChan <- lora.NewRadioEvent(lora.RadioEventTxDone, uint16(st), nil)
	}

	if (st & SX127X_IRQ_LORA_RXTOUT_MASK) > 0 {
		rChan <- lora.NewRadioEvent(lora.RadioEventTimeout, uint16(st), nil)
	}

	if (st & SX127X_IRQ_LORA_CRCERR_MASK) > 0 {
		rChan <- lora.NewRadioEvent(lora.RadioEventCrcError, uint16(st), nil)
	}
}

func bandwidth(bw uint8) uint8 {
	switch bw {
	case lora.Bandwidth_7_8:
		return SX127X_LORA_BW_7_8
	case lora.Bandwidth_10_4:
		return SX127X_LORA_BW_10_4
	case lora.Bandwidth_15_6:
		return SX127X_LORA_BW_15_6
	case lora.Bandwidth_20_8:
		return SX127X_LORA_BW_20_8
	case lora.Bandwidth_31_25:
		return SX127X_LORA_BW_31_25
	case lora.Bandwidth_41_7:
		return SX127X_LORA_BW_41_7
	case lora.Bandwidth_62_5:
		return SX127X_LORA_BW_62_5
	case lora.Bandwidth_125_0:
		return SX127X_LORA_BW_125_0
	case lora.Bandwidth_250_0:
		return SX127X_LORA_BW_250_0
	case lora.Bandwidth_500_0:
		return SX127X_LORA_BW_500_0
	default:
		return 0
	}
}

func syncword(sw int) uint16 {
	if sw == lora.SyncPublic {
		return SX127X_LORA_MAC_PUBLIC_SYNCWORD
	}
	return SX127X_LORA_MAC_PRIVATE_SYNCWORD
}
