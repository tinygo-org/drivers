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
)

const (
	RadioEventRxDone    = iota
	RadioEventTxDone    = iota
	RadioEventTimeout   = iota
	RadioEventWatchdog  = iota
	RadioEventCrcError  = iota
	RadioEventUnhandled = iota
)

// So we can keep track of the origin of interruption
const (
	SPI_BUFFER_SIZE = 256
)

// RadioEvent are used for communicating in the radio Event Channel
type RadioEvent struct {
	EventType int
	IRQStatus uint8
	EventData []byte
}

// Device wraps an SPI connection to a SX127x device.
type Device struct {
	spi            drivers.SPI     // SPI bus for module communication
	rstPin, csPin  machine.Pin     // GPIOs for reset and chip select
	radioEventChan chan RadioEvent // Channel for Receiving events
	loraConf       LoraConfig      // Current Lora configuration
	deepSleep      bool            // Internal Sleep state
	deviceType     int             // sx1261,sx1262,sx1268 (defaults sx1261)
	spiBuffer      [SPI_BUFFER_SIZE]uint8
	packetIndex    uint8 // FIXME ... useless ?
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

// --------------------------------------------------
//  Channel and events
// --------------------------------------------------
//NewRadioEvent() returns a new RadioEvent that can be used in the RadioChannel
func NewRadioEvent(eType int, irqStatus uint8, eData []byte) RadioEvent {
	r := RadioEvent{EventType: eType, IRQStatus: irqStatus, EventData: eData}
	return r
}

// Get the RadioEvent channel of the device
func (d *Device) GetRadioEventChan() chan RadioEvent {
	return d.radioEventChan
}

// New creates a new SX127x connection. The SPI bus must already be configured.
func New(spi machine.SPI, csPin machine.Pin, rstPin machine.Pin) *Device {
	k := Device{
		spi:            spi,
		csPin:          csPin,
		rstPin:         rstPin,
		radioEventChan: make(chan RadioEvent, 10),
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
	cur := d.ReadRegister(REG_OP_MODE)
	new := (cur & (^SX127X_OPMODE_MASK)) | mode
	d.WriteRegister(REG_OP_MODE, new)
}

// SetOpMode changes the sx1276 mode
func (d *Device) SetOpModeLora() {
	d.WriteRegister(REG_OP_MODE, SX127X_OPMODE_LORA)
}

// SetupLora configures sx127x Lora mode
func (d *Device) SetupLora(config LoraConfig) error {

	d.loraConf = config

	// Reset the device first
	d.Reset()

	// Switch to Lora mode
	d.SetOpModeLora()
	d.SetOpMode(SX127X_OPMODE_SLEEP)

	// Access High Frequency Mode
	d.SetLowFrequencyModeOn(false)

	// Set PA Ramp time 50 uS
	d.WriteRegister(REG_PA_RAMP, (d.ReadRegister(REG_PA_RAMP)&0xF0)|0x08) // set PA ramp-up time 50 uSec

	// Enable power (manage Over Current, PA_Boost ... etc)
	d.SetTxPower(11, true)

	// Set Low Noise Amplifier to MAX
	d.WriteRegister(REG_LNA, LNA_MAX_GAIN)

	// Set Frequency
	d.SetFrequency(d.loraConf.Freq)

	// Set Bandwidth
	d.SetBandwidth(d.loraConf.Bw)

	//Set Coding Rate (TODO : Check)
	d.SetCodingRate(d.loraConf.Cr)

	// Set explicit header
	d.SetHeaderMode(SX127X_LORA_HEADER_EXPLICIT)

	// Enable CRC
	d.SetRxPayloadCrc(SX127X_LORA_CRC_ON)

	// Disable IQ Polarization
	d.SetIQPolarity(SX127X_LORA_IQ_STANDARD)

	// Disable HOP PERIOD
	d.SetHopPeriod(0x00)

	// Continuous Mode
	d.SetTxContinuousMode(false)

	// Set Lora Sync
	d.SetSyncWord(0x34)

	//Set Max payload length (default value)
	d.WriteRegister(REG_MAX_PAYLOAD_LENGTH, 0xFF)
	// Mandatory in Implicit header Mode (default value)
	d.WriteRegister(REG_PAYLOAD_LENGTH, 0x01)

	// AGC On
	d.SetAgcAuto(SX127X_AGC_AUTO_ON)

	// set FIFO base addresses
	d.WriteRegister(REG_FIFO_TX_BASE_ADDR, 0)
	d.WriteRegister(REG_FIFO_RX_BASE_ADDR, 0)
	return nil
}

// TxLora sends a packet in Lora mode
// Intmode will enable interrupt mode.
// If disabled, function will probe registers for TXDone before
// returning
func (d *Device) TxLora(payload []byte) error {

	// Are we already in Lora mode ?
	r := d.ReadRegister(REG_OP_MODE)
	if (r & SX127X_OPMODE_LORA) != SX127X_OPMODE_LORA {
		return errors.New("Not in Lora mode")
	}

	// set the IRQ mapping DIO0=TxDone DIO1=NOP DIO2=NOP
	d.WriteRegister(REG_DIO_MAPPING_1, MAP_DIO0_LORA_TXDONE|MAP_DIO1_LORA_NOP|MAP_DIO2_LORA_NOP)
	// Clear all radio IRQ Flags
	d.WriteRegister(REG_IRQ_FLAGS, 0xFF)
	// Mask all but TxDone
	d.WriteRegister(REG_IRQ_FLAGS_MASK, ^IRQ_LORA_TXDONE_MASK)

	// initialize the payload size and address pointers
	d.WriteRegister(REG_FIFO_TX_BASE_ADDR, 0)
	d.WriteRegister(REG_FIFO_ADDR_PTR, 0)
	d.WriteRegister(REG_PAYLOAD_LENGTH, uint8(len(payload)))

	// Copy payload to FIFO // TODO: Bulk
	for i := 0; i < len(payload); i++ {
		d.WriteRegister(REG_FIFO, payload[i])
	}

	// Enable TX
	d.SetOpMode(SX127X_OPMODE_TX)

	msg := <-d.GetRadioEventChan()
	if msg.EventType != RadioEventTxDone {
		return errors.New("Unexpected Radio Event while TX")
	}
	return nil
}

/*
//CheckIrq can be called periodicaly to check for RXDONE,TXDONE,RXTOUT
//but It would be more efficient to call it on  DIO0/1 pins rising edge
func (d *Device) CheckIrq() {

	irqFlags := d.ReadRegister(REG_IRQ_FLAGS)
	//println("sx21276: irq=", irqFlags)

	// We have a packet
	if (irqFlags & IRQ_LORA_RXDONE_MASK) > 0 {
		//println("sx1276: RXDONE")
		// Read current packet
		buf := []byte{}
		packetLength := d.ReadRegister(REG_RX_NB_BYTES)
		d.WriteRegister(REG_FIFO_ADDR_PTR, d.ReadRegister(REG_FIFO_RX_CURRENT_ADDR)) // Reset FIFO Read Addr
		for i := uint8(0); i < packetLength; i++ {
			buf = append(buf, d.ReadRegister(REG_FIFO))
		}
		// Send RXDONE to the defined event channel
		d.radioEventChan <- RadioEvent{EventType: EventRxDone, EventData: buf}
	}
	if (irqFlags & IRQ_LORA_TXDONE_MASK) > 0 {
		//println("sx1276: TXDONE")
		d.radioEventChan <- RadioEvent{EventType: EventTxDone, EventData: nil}
	}
	if (irqFlags & IRQ_LORA_RXTOUT_MASK) > 0 {
		//println("sx1276: RXTOUT")
		d.radioEventChan <- RadioEvent{EventType: EventRxTimeout, EventData: nil}
	}

	// Sigh: on some processors, for some unknown reason, doing this only once does not actually
	// clear the radio's interrupt flag. So we do it twice. Why?
	d.WriteRegister(REG_IRQ_FLAGS, irqFlags) // Clear all IRQ flags
	d.WriteRegister(REG_IRQ_FLAGS, irqFlags) // Clear all IRQ flags
}
*/

/*
// SetRadioEventChan defines a channel so the driver can send its Radio Events
func (d *Device) SetRadioEventChan(channel chan RadioEvent) {
	d.radioEventChan = channel
}
*/
/*
// Init reboots the SX1276 module
func (d *Device) Init(cfg LoraConfig) (err error) {
	d.loraConf = cfg
	d.csPin.High()
	d.Reset()
	return nil
}
*/
/*

// ConfigureLoraModem prepares for LORA communications
func (d *Device) ConfigureLoraModem() {

	// Sleep mode required to go LOra
	d.OpMode(OPMODE_SLEEP)
	// Set Lora mode (from sleep)
	d.OpModeLora()
	// Switch to standby mode
	d.OpMode(OPMODE_STANDBY)
	// Set Bandwidth
	d.SetBandwidth(d.cnf.Bandwidth)
	// Disable IQ Polarization
	d.SetInvertedIQ(false)
	// Set implicit header
	d.SetImplicitHeaderModeOn(false)
	// We want CRC
	d.SetRxPayloadCrc(true)
	d.SetAgcAutoOn(true)
	if d.GetBandwidth() == 125000 && (d.GetSpreadingFactor() == 11 || d.GetSpreadingFactor() == 12) {
		d.SetLowDataRateOptimOn(true)
	}

	// Configure Output Power
	d.WriteRegister(REG_PA_RAMP, (d.ReadRegister(REG_PA_RAMP)&0xF0)|0x08) // set PA ramp-up time 50 uSec
	d.WriteRegister(REG_PA_CONFIG, 0xFF)                                  //PA_BOOST MAX
	d.SetOCP(140)                                                         // Over Current protection

	// RX and premamble
	d.WriteRegister(REG_PREAMBLE_MSB, 0x00)     // Preamble set to 8 symp
	d.WriteRegister(REG_PREAMBLE_LSB, 0x08)     // -> 0x0008 + 4 = 12
	d.WriteRegister(REG_SYMB_TIMEOUT_LSB, 0x25) //Rx Timeout 37 symbol

	// set FIFO base addresses
	d.WriteRegister(REG_FIFO_TX_BASE_ADDR, 0)
	d.WriteRegister(REG_FIFO_RX_BASE_ADDR, 0)
}
*/
//GetVersion returns hardware version of sx1276 chipset
func (d *Device) GetVersion() uint8 {
	return (d.ReadRegister(REG_VERSION))
}

// IsTransmitting tests if a packet transmission is in progress
func (d *Device) IsTransmitting() bool {
	return (d.ReadRegister(REG_OP_MODE) & SX127X_OPMODE_TX) == SX127X_OPMODE_TX
}

// ReadPacket reads a received packet into a byte array
func (d *Device) ReadPacket(packet []byte) int {
	available := int(d.ReadRegister(REG_RX_NB_BYTES) - d.packetIndex)
	if available > len(packet) {
		available = len(packet)
	}

	for i := 0; i < available; i++ {
		d.packetIndex++
		packet[i] = d.ReadRegister(REG_FIFO)
	}

	return available
}

// LastPacketRSSI gives the RSSI of the last packet received
func (d *Device) LastPacketRSSI() uint8 {
	// section 5.5.5
	var adjustValue uint8 = 157
	if d.loraConf.Freq < 868000000 {
		adjustValue = 164
	}
	return d.ReadRegister(REG_PKT_RSSI_VALUE) - adjustValue
}

// LastPacketSNR gives the SNR of the last packet received
func (d *Device) LastPacketSNR() uint8 {
	return uint8(d.ReadRegister(REG_PKT_SNR_VALUE) / 4)
}

/*
// GetFrequency returns the frequency the LoRa module is using
func (d *Device) GetFrequency() uint32 {
	f := uint64(d.ReadRegister(REG_FRF_LSB))
	f += uint64(d.ReadRegister(REG_FRF_MID)) << 8
	f += uint64(d.ReadRegister(REG_FRF_MSB)) << 16
	f = (f * 32000000) >> 19 //FSTEP = FXOSC/2^19
	return uint32(f)
}
*/
// SetFrequency updates the frequency the LoRa module is using
func (d *Device) SetFrequency(frequency uint32) {
	d.loraConf.Freq = frequency
	var frf = (uint64(frequency) << 19) / 32000000
	d.WriteRegister(REG_FRF_MSB, uint8(frf>>16))
	d.WriteRegister(REG_FRF_MID, uint8(frf>>8))
	d.WriteRegister(REG_FRF_LSB, uint8(frf>>0))
}

// GetSpreadingFactor returns the spreading factor the LoRa module is using
func (d *Device) GetSpreadingFactor() uint8 {
	return d.ReadRegister(REG_MODEM_CONFIG_2) >> 4
}

// GetRSSI returns current RSSI
func (d *Device) GetRSSI() uint8 {
	return d.ReadRegister(REG_RSSI_VALUE)
}

/*
// GetBandwidth returns the bandwidth the LoRa module is using
func (d *Device) GetBandwidth() int32 {
	return int32(d.loraConf.Bw)
}
*/

//SetSyncWord defines sync word
func (d *Device) SetSyncWord(syncWord uint8) {
	d.WriteRegister(REG_SYNC_WORD, syncWord)
}

// SetIQPolarity Sets I/Q polarity configuration
func (d *Device) SetIQPolarity(val uint8) {
	if val == SX127X_LORA_IQ_INVERTED {
		//Invert IQ Back
		d.WriteRegister(0x33, 0x67)
		d.WriteRegister(0x3B, 0x19)
	} else {
		//Set IQ to normal values
		d.WriteRegister(0x33, 0x27)
		d.WriteRegister(0x3B, 0x1D)
	}
}

// RxLora sets device in receive mode
func (d *Device) RxLora() {
	// set the IRQ mapping DIO0=TxDone DIO1=NOP DIO2=NOP
	d.WriteRegister(REG_DIO_MAPPING_1, MAP_DIO0_LORA_RXDONE|MAP_DIO1_LORA_NOP|MAP_DIO2_LORA_NOP)
	// Clear all radio IRQ Flags
	d.WriteRegister(REG_IRQ_FLAGS, 0xFF)
	// Mask all but TxDone
	d.WriteRegister(REG_IRQ_FLAGS_MASK, ^IRQ_LORA_RXDONE_MASK)

	d.SetOpMode(SX127X_OPMODE_RX) // RX Mode
}

/*
// setLdoFlag() enables LowDataRateOptimize bit (mandated when symbol length >16ms)
// LGTM
func (d *Device) setLdoFlag() {
	// Section 4.1.1.5
	var symbolDuration = 1000 / (d.GetBandwidth() / (1 << d.GetSpreadingFactor()))

	var config3 = d.ReadRegister(REG_MODEM_CONFIG_3)

	// Section 4.1.1.6
	if symbolDuration > 16 {
		config3 = config3 | 0x08
	} else {
		config3 = config3 & 0xF7
	}

	d.WriteRegister(REG_MODEM_CONFIG_3, config3)
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
		d.WriteRegister(REG_PA_CONFIG, uint8(0x70)|uint8(txPower))

	} else {
		//PA_BOOST
		if txPower > 17 {
			if txPower > 20 {
				txPower = 20
			}

			txPower -= 3

			// High Power +20 dBm Operation (Semtech SX1276/77/78/79 5.4.3.)
			d.WriteRegister(REG_PA_DAC, 0x87)
			d.SetOCP(140)
		} else {
			if txPower < 2 {
				txPower = 2
			}

			d.WriteRegister(REG_PA_DAC, 0x84)
			d.SetOCP(100)

		}

		d.WriteRegister(REG_PA_CONFIG, uint8(PA_BOOST)|uint8(txPower-2))

	}
}

// SetRxTimeout defines RX Timeout expressed as number of symbols
func (d *Device) SetRxTimeout(tmoutSymb uint8) {
	d.WriteRegister(REG_SYMB_TIMEOUT_LSB, tmoutSymb)
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

	d.WriteRegister(REG_OCP, 0x20|(0x1F&ocpTrim))
}

// ---------------
// RegModemConfig1
// ---------------

// SetBandwidth updates the bandwidth the LoRa module is using
func (d *Device) SetBandwidth(bw uint8) {
	d.loraConf.Bw = bw
	d.WriteRegister(REG_MODEM_CONFIG_1, (d.ReadRegister(REG_MODEM_CONFIG_1)&0x0f)|(bw<<4))
}

// SetCodingRate updates the coding rate the LoRa module is using
func (d *Device) SetCodingRate(cr uint8) {
	d.loraConf.Cr = cr
	d.WriteRegister(REG_MODEM_CONFIG_1, (d.ReadRegister(REG_MODEM_CONFIG_1)&0xf1)|(cr<<1))
}

// SetImplicitHeaderModeOn Enables implicit header mode ***
func (d *Device) SetHeaderMode(headerType uint8) {
	d.loraConf.HeaderType = headerType
	if headerType == SX127X_LORA_HEADER_IMPLICIT {
		d.WriteRegister(REG_MODEM_CONFIG_1, d.ReadRegister(REG_MODEM_CONFIG_1)|0x01)
	} else {
		d.WriteRegister(REG_MODEM_CONFIG_1, d.ReadRegister(REG_MODEM_CONFIG_1)&0xfe)
	}
}

// ---------------
// RegModemConfig2
// ---------------

// SetSpreadingFactor updates the spreading factor the LoRa module is using
func (d *Device) SetSpreadingFactor(sf uint8) {
	d.loraConf.Sf = sf
	if sf == SX127X_LORA_SF6 {
		d.WriteRegister(REG_DETECTION_OPTIMIZE, 0xc5)
		d.WriteRegister(REG_DETECTION_THRESHOLD, 0x0c)
	} else {
		d.WriteRegister(REG_DETECTION_OPTIMIZE, 0xc3)
		d.WriteRegister(REG_DETECTION_THRESHOLD, 0x0a)
	}
	var newValue = (d.ReadRegister(REG_MODEM_CONFIG_2) & 0x0f) | ((sf << 4) & 0xf0)
	d.WriteRegister(REG_MODEM_CONFIG_2, newValue)
}

// SetTxContinuousMode enable Continuous Tx mode
func (d *Device) SetTxContinuousMode(val bool) {
	if val {
		d.WriteRegister(REG_MODEM_CONFIG_2, d.ReadRegister(REG_MODEM_CONFIG_2)|0x08)
	} else {
		d.WriteRegister(REG_MODEM_CONFIG_2, d.ReadRegister(REG_MODEM_CONFIG_2)&0xf7)
	}
}

// SetRxPayloadCrc Enable CRC generation and check on payload
func (d *Device) SetRxPayloadCrc(val uint8) {
	if val == SX127X_LORA_CRC_ON {
		d.WriteRegister(REG_MODEM_CONFIG_2, d.ReadRegister(REG_MODEM_CONFIG_2)|0x04)
	} else {
		d.WriteRegister(REG_MODEM_CONFIG_2, d.ReadRegister(REG_MODEM_CONFIG_2)&0xfb)
	}
}

// ---------------
// RegModemConfig3
// ---------------

// SetAgcAutoOn enables Automatic Gain Control
func (d *Device) SetAgcAuto(val uint8) {
	if val == SX127X_AGC_AUTO_ON {
		d.WriteRegister(REG_MODEM_CONFIG_3, d.ReadRegister(REG_MODEM_CONFIG_3)|0x04)
	} else {
		d.WriteRegister(REG_MODEM_CONFIG_3, d.ReadRegister(REG_MODEM_CONFIG_3)&0xfb)
	}
}

// SetLowDataRateOptimize enables Low Data Rate Optimization
func (d *Device) SetLowDataRateOptim(val uint8) {
	if val == SX127X_LOW_DATARATE_OPTIM_ON {
		d.WriteRegister(REG_MODEM_CONFIG_3, d.ReadRegister(REG_MODEM_CONFIG_3)|0x08)
	} else {
		d.WriteRegister(REG_MODEM_CONFIG_3, d.ReadRegister(REG_MODEM_CONFIG_3)&0xf7)
	}
}

// SetLowFrequencyModeOn enables Low Data Rate Optimization
func (d *Device) SetLowFrequencyModeOn(val bool) {
	if val {
		d.WriteRegister(REG_OP_MODE, d.ReadRegister(REG_OP_MODE)|0x04)
	} else {
		d.WriteRegister(REG_OP_MODE, d.ReadRegister(REG_OP_MODE)&0xfb)
	}
}

// SetHopPeriod sets number of symbol periods between frequency hops. (0 = disabled).
func (d *Device) SetHopPeriod(val uint8) {
	d.WriteRegister(REG_HOP_PERIOD, val)
}

// HandleInterrupt must be called by main code on DIO state change.
func (d *Device) HandleInterrupt() {
	// Get IRQ and clear
	st := d.ReadRegister(REG_IRQ_FLAGS)
	d.WriteRegister(REG_IRQ_FLAGS, 0xFF)

	rChan := d.GetRadioEventChan()

	if (st & IRQ_LORA_RXDONE_MASK) > 0 {
		rChan <- NewRadioEvent(RadioEventRxDone, st, nil)
	}

	if (st & IRQ_LORA_TXDONE_MASK) > 0 {
		rChan <- NewRadioEvent(RadioEventTxDone, st, nil)
	}

	if (st & IRQ_LORA_RXTOUT_MASK) > 0 {
		rChan <- NewRadioEvent(RadioEventTimeout, st, nil)
	}

	if (st & IRQ_LORA_CRCERR_MASK) > 0 {
		rChan <- NewRadioEvent(RadioEventCrcError, st, nil)
	}
}

// PrintRegisters outputs the sx127x transceiver registers
func (d *Device) PrintRegisters(compact bool) {
	for i := uint8(0); i < 128; i++ {
		v := d.ReadRegister(i)
		print(v, " ")
	}
	println()
}
