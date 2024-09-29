// Package sx127x provides a driver for SX127x LoRa transceivers.
//
// Datasheet:
// https://www.semtech.com/uploads/documents/DS_SX1276-7-8-9_W_APP_V6.pdf
//
// LoRa Configuration Parameters:
//
// Frequency: is the frequency the tranceiver uses. Valid frequencies depend on
//  the type of LoRa module, typically around 433MHz or 866MHz. It has
//  a granularity of about 23Hz, how close it can be to others depends on the
//  Bandwidth being used.
//
// Bandwidth: is the bandwidth used for tranmissions, ranging from 7k8 to 512k
//  A higher bandwidth gives faster transmissions, lower gives greater range
//
// SpreadingFactor: is how a transmission is spread over the spectrum. It ranges
//  from 6 to 12, a higher value gives greater range but slower transmissions.
//
// CodingRate: is the cyclic error coding used to improve the robustness of the
//  transmission. It ranges from 5 to 8, a higher value gives greater
//  reliability but slower transmissions.
//
// TxPower: is the power used for the transmission, ranging from 1 to 20.
//  A higher power gives greater range. Regulations in your country likely
//  limit the maximum power permited.
//
// Presently this driver is only synchronous and so does not use any DIOx pins
//
package sx127x

import (
	"errors"
	"fmt"
	"machine"
	"time"
)

// Device wraps an SPI connection to a SX127x device.
type Device struct {
	spi                machine.SPI
	csPin              machine.Pin
	rstPin             machine.Pin
	packetIndex        uint8
	implicitHeaderMode bool
}

// Config holds the LoRa configuration parameters
type Config struct {
	Frequency       uint32
	SpreadingFactor uint8
	Bandwidth       int32
	CodingRate      uint8
	TxPower         int8
}

// New creates a new SX127x connection. The SPI bus must already be configured.
func New(spi machine.SPI, csPin machine.Pin, rstPin machine.Pin) Device {
	return Device{
		spi:    spi,
		csPin:  csPin,
		rstPin: rstPin,
	}
}

// Configure initializes and configures the LoRa module
func (d *Device) Configure(cfg Config) (err error) {
	d.csPin.High()

	d.Reset()

	if d.readRegister(REG_VERSION) != 0x12 {
		return errors.New("SX127x module not found")
	}

	d.Sleep()

	// set base addresses
	d.writeRegister(REG_FIFO_TX_BASE_ADDR, 0)
	d.writeRegister(REG_FIFO_RX_BASE_ADDR, 0)

	// set LNA boost
	d.writeRegister(REG_LNA, d.readRegister(REG_LNA)|0x03)

	// set auto AGC
	d.writeRegister(REG_MODEM_CONFIG_3, 0x04)

	if cfg.Frequency != 0 {
		d.SetFrequency(cfg.Frequency)
	}
	if cfg.SpreadingFactor != 0 {
		d.SetSpreadingFactor(cfg.SpreadingFactor)
	}
	if cfg.Bandwidth != 0 {
		d.SetBandwidth(cfg.Bandwidth)
	}
	if cfg.CodingRate != 0 {
		d.SetCodingRate(cfg.CodingRate)
	}
	if cfg.TxPower != 0 {
		d.SetTxPower(cfg.TxPower)
	}

	d.Standby()

	return err
}

// SendPacket transmits a packet
// Note that this will return before the packet has finished being sent,
// use the IsTransmitting() function if you need to know when sending is done.
func (d *Device) SendPacket(packet []byte) {

	// wait for any previous SendPacket to be done
	for d.IsTransmitting() {
		time.Sleep(1 * time.Millisecond)
	}

	// reset TX_DONE, FIFO address and payload length
	d.writeRegister(REG_IRQ_FLAGS, IRQ_TX_DONE_MASK)
	d.writeRegister(REG_FIFO_ADDR_PTR, 0)
	d.writeRegister(REG_PAYLOAD_LENGTH, 0)

	if len(packet) > MAX_PKT_LENGTH {
		packet = packet[0:MAX_PKT_LENGTH]
	}

	for i := 0; i < len(packet); i++ {
		d.writeRegister(REG_FIFO, packet[i])
	}

	d.writeRegister(REG_PAYLOAD_LENGTH, uint8(len(packet)))
	d.writeRegister(REG_OP_MODE, MODE_LONG_RANGE_MODE|MODE_TX)
}

// IsTransmitting tests if a packet transmission is in progress
func (d *Device) IsTransmitting() bool {
	return (d.readRegister(REG_OP_MODE) & MODE_TX) == MODE_TX
}

// ParsePacket returns the size of a received packet waiting to be read
func (d *Device) ParsePacket(size uint8) uint8 {
	var packetLength uint8
	irqFlags := d.readRegister(REG_IRQ_FLAGS)

	if size > 0 {
		d.implicitMode()
		d.writeRegister(REG_PAYLOAD_LENGTH, size&0xff)
	} else {
		d.explicitMode()
	}

	// clear IRQ's
	d.writeRegister(REG_IRQ_FLAGS, irqFlags)

	if (irqFlags&IRQ_RX_DONE_MASK) == 0 && (irqFlags&IRQ_PAYLOAD_CRC_ERROR_MASK) == 0 {
		// received a packet
		d.packetIndex = 0

		// read packet length
		if d.implicitHeaderMode {
			packetLength = d.readRegister(REG_PAYLOAD_LENGTH)
		} else {
			packetLength = d.readRegister(REG_RX_NB_BYTES)
		}

		// set FIFO address to current RX address
		d.writeRegister(REG_FIFO_ADDR_PTR, d.readRegister(REG_FIFO_RX_CURRENT_ADDR))

		// put in standby mode
		d.Standby()

	} else if d.readRegister(REG_OP_MODE) != (MODE_LONG_RANGE_MODE | MODE_RX_SINGLE) {
		// not currently in RX mode

		// reset FIFO address
		d.writeRegister(REG_FIFO_ADDR_PTR, 0)

		// put in single RX mode
		d.writeRegister(REG_OP_MODE, MODE_LONG_RANGE_MODE|MODE_RX_SINGLE)
	}

	return packetLength
}

// ReadPacket reads a received packet into a byte array
func (d *Device) ReadPacket(packet []byte) int {
	available := int(d.readRegister(REG_RX_NB_BYTES) - d.packetIndex)
	if available > len(packet) {
		available = len(packet)
	}

	for i := 0; i < available; i++ {
		d.packetIndex++
		packet[i] = d.readRegister(REG_FIFO)
	}

	return available
}

// LastPacketRSSI gives the RSSI of the last packet received
func (d *Device) LastPacketRSSI() uint8 {
	// section 5.5.5
	var adjustValue uint8 = 157
	if d.GetFrequency() < 868000000 {
		adjustValue = 164
	}
	return d.readRegister(REG_PKT_RSSI_VALUE) - adjustValue
}

// LastPacketSNR gives the SNR of the last packet received
func (d *Device) LastPacketSNR() uint8 {
	return uint8(d.readRegister(REG_PKT_SNR_VALUE) / 4)
}

// LastPacketFrequencyError gives the frequency error of the last packet received
// You can use this to adjust this transeiver frequency to more closly match the
// frequency being used by the sender, as this can drift over time
func (d *Device) LastPacketFrequencyError() int32 {
	// TODO
	// int32_t freqError = 0;
	// freqError = static_cast<int32_t>(readRegister(REG_FREQ_ERROR_MSB) & B111);
	// freqError <<= 8L;
	// freqError += static_cast<int32_t>(readRegister(REG_FREQ_ERROR_MID));
	// freqError <<= 8L;
	// freqError += static_cast<int32_t>(readRegister(REG_FREQ_ERROR_LSB));
	//
	// if (readRegister(REG_FREQ_ERROR_MSB) & B1000) { // Sign bit is on
	//    freqError -= 524288; // B1000'0000'0000'0000'0000
	// }
	//
	// const float fXtal = 32E6; // FXOSC: crystal oscillator (XTAL) frequency (2.5. Chip Specification, p. 14)
	// const float fError = ((static_cast<float>(freqError) * (1L << 24)) / fXtal) * (getSignalBandwidth() / 500000.0f); // p. 37
	//
	// return static_cast<long>(fError);
	return 0
}

// PrintRegisters outputs the sx127x transceiver registers
func (d *Device) PrintRegisters() {
	for i := 0; i < 128; i++ {
		fmt.Printf("%02x: %02x\n", i, d.readRegister(uint8(i)))
	}
}

// Reset the sx127x device
func (d *Device) Reset() {
	d.rstPin.Low()
	time.Sleep(10 * time.Millisecond)
	d.rstPin.High()
	time.Sleep(10 * time.Millisecond)
}

// Sleep puts the sx127x device into sleep mode
func (d *Device) Sleep() {
	d.writeRegister(REG_OP_MODE, MODE_LONG_RANGE_MODE|MODE_SLEEP)
}

// Standby puts the sx127x device into standby mode
func (d *Device) Standby() {
	d.writeRegister(REG_OP_MODE, MODE_LONG_RANGE_MODE|MODE_STDBY)
}

// GetFrequency returns the frequency the LoRa module is using
func (d *Device) GetFrequency() uint32 {
	f := uint64(d.readRegister(REG_FRF_LSB))
	f += uint64(d.readRegister(REG_FRF_MID)) << 8
	f += uint64(d.readRegister(REG_FRF_MSB)) << 16
	f = (f * 32000000) >> 19
	return uint32(f)
}

// SetFrequency updates the frequency the LoRa module is using
func (d *Device) SetFrequency(frequency uint32) {
	var frf = (uint64(frequency) << 19) / 32000000
	d.writeRegister(REG_FRF_MSB, uint8(frf>>16))
	d.writeRegister(REG_FRF_MID, uint8(frf>>8))
	d.writeRegister(REG_FRF_LSB, uint8(frf>>0))
}

// GetSpreadingFactor returns the spreading factor the LoRa module is using
func (d *Device) GetSpreadingFactor() uint8 {
	return d.readRegister(REG_MODEM_CONFIG_2) >> 4
}

// SetSpreadingFactor updates the spreading factor the LoRa module is using
func (d *Device) SetSpreadingFactor(spreadingFactor uint8) {
	if spreadingFactor < 6 {
		spreadingFactor = 6
	} else if spreadingFactor > 12 {
		spreadingFactor = 12
	}

	if spreadingFactor == 6 {
		d.writeRegister(REG_DETECTION_OPTIMIZE, 0xc5)
		d.writeRegister(REG_DETECTION_THRESHOLD, 0x0c)
	} else {
		d.writeRegister(REG_DETECTION_OPTIMIZE, 0xc3)
		d.writeRegister(REG_DETECTION_THRESHOLD, 0x0a)
	}

	var newValue = (d.readRegister(REG_MODEM_CONFIG_2) & 0x0f) | ((spreadingFactor << 4) & 0xf0)
	d.writeRegister(REG_MODEM_CONFIG_2, newValue)
	d.setLdoFlag()
}

// GetBandwidth returns the bandwidth the LoRa module is using
func (d *Device) GetBandwidth() int32 {
	var bw = d.readRegister(REG_MODEM_CONFIG_1) >> 4

	switch bw {
	case 0:
		return 7800
	case 1:
		return 10400
	case 2:
		return 15600
	case 3:
		return 20800
	case 4:
		return 31250
	case 5:
		return 41700
	case 6:
		return 62500
	case 7:
		return 125000
	case 8:
		return 250000
	case 9:
		return 500000
	}

	return -1
}

// SetBandwidth updates the bandwidth the LoRa module is using
func (d *Device) SetBandwidth(sbw int32) {
	var bw uint8

	if sbw <= 7800 {
		bw = 0
	} else if sbw <= 10400 {
		bw = 1
	} else if sbw <= 15600 {
		bw = 2
	} else if sbw <= 20800 {
		bw = 3
	} else if sbw <= 31250 {
		bw = 4
	} else if sbw <= 41700 {
		bw = 5
	} else if sbw <= 62500 {
		bw = 6
	} else if sbw <= 125000 {
		bw = 7
	} else if sbw <= 250000 {
		bw = 8
	} else {
		bw = 9
	}

	d.writeRegister(REG_MODEM_CONFIG_1, (d.readRegister(REG_MODEM_CONFIG_1)&0x0f)|(bw<<4))
	d.setLdoFlag()
}

func (d *Device) setLdoFlag() {
	// Section 4.1.1.5
	var symbolDuration = 1000 / (d.GetBandwidth() / (1 << d.GetSpreadingFactor()))

	var config3 = d.readRegister(REG_MODEM_CONFIG_3)

	// Section 4.1.1.6
	if symbolDuration > 16 {
		config3 = config3 | 0x08
	} else {
		config3 = config3 & 0xF7
	}

	d.writeRegister(REG_MODEM_CONFIG_3, config3)
}

// SetCodingRate updates the coding rate the LoRa module is using
func (d *Device) SetCodingRate(denominator uint8) {
	if denominator < 5 {
		denominator = 5
	} else if denominator > 8 {
		denominator = 8
	}
	var cr = denominator - 4
	d.writeRegister(REG_MODEM_CONFIG_1, (d.readRegister(REG_MODEM_CONFIG_1)&0xf1)|(cr<<1))
}

// SetTxPower sets the transmitter output power
func (d *Device) SetTxPower(txPower int8) {
	//TODO
	// if txPower < 2 {
	// 	// power is less than 2 dBm, enable PA on RFO
	// 	writeRegister(REG_PA_CONFIG, PA_SELECT_RFO, 7, 7)
	// 	writeRegister(REG_PA_CONFIG, LOW_POWER|(txPower+3), 6, 0)
	// 	writeRegister(REG_PA_DAC, PA_BOOST_OFF, 2, 0)
	// } else if (txPower >= 2) && (txPower <= 17) {
	// 	// power is 2 - 17 dBm, enable PA1 + PA2 on PA_BOOST
	// 	writeRegister(REG_PA_CONFIG, PA_SELECT_BOOST, 7, 7)
	// 	writeRegister(REG_PA_CONFIG, MAX_POWER|(txPower-2), 6, 0)
	// 	writeRegister(REG_PA_DAC, PA_BOOST_OFF, 2, 0)
	// } else if txPower == 20 {
	// 	// power is 20 dBm, enable PA1 + PA2 on PA_BOOST and enable high power mode
	// 	writeRegister(REG_PA_CONFIG, PA_SELECT_BOOST, 7, 7)
	// 	writeRegister(REG_PA_CONFIG, MAX_POWER|(txPower-5), 6, 0)
	// 	writeRegister(REG_PA_DAC, PA_BOOST_ON, 2, 0)
	// }
}

func (d *Device) explicitMode() {
	d.implicitHeaderMode = false
	d.writeRegister(REG_MODEM_CONFIG_1, d.readRegister(REG_MODEM_CONFIG_1)&0xfe)
}

func (d *Device) implicitMode() {
	d.implicitHeaderMode = true
	d.writeRegister(REG_MODEM_CONFIG_1, d.readRegister(REG_MODEM_CONFIG_1)|0x01)
}

func (d *Device) readRegister(reg uint8) uint8 {
	d.csPin.Low()
	d.spi.Tx([]byte{reg & 0x7f}, nil)
	var value [1]byte
	d.spi.Tx(nil, value[:])
	d.csPin.High()
	return value[0]
}

func (d *Device) writeRegister(reg uint8, value uint8) uint8 {
	var response [1]byte
	d.csPin.Low()
	d.spi.Tx([]byte{reg | 0x80}, nil)
	d.spi.Tx([]byte{value}, response[:])
	d.csPin.High()
	return response[0]
}
