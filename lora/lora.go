// Package lora provides a driver for SX127x LoRa transceivers.
//
// Datasheet:
// https://www.semtech.com/uploads/documents/DS_SX1276-7-8-9_W_APP_V6.pdf
//
package lora

import (
	"errors"
	"fmt"
	"machine"
	"time"
)

// Device wraps an SPI connection to a SX127x device.
type Device struct {
	bus     machine.SPI
	csPin   machine.GPIO
	rstPin  machine.GPIO
	dio0Pin machine.GPIO
}

type Config struct {
	Frequency       uint32
	SpreadingFactor uint8
	Bandwidth       int32
	CodingRate      uint8
	TxPower         int8
}

// New creates a new SX127x connection. The SPI bus must already be configured.
func New(b machine.SPI, csPin machine.GPIO, rstPin machine.GPIO, dio0Pin machine.GPIO) Device {
	return Device{bus: b,
		csPin:   csPin,
		rstPin:  rstPin,
		dio0Pin: dio0Pin,
	}
}

// Configure initializes the display with default configuration
func (d *Device) Configure(cfg Config) (err error) {
	d.csPin.High()

	d.reset()

	if d.readRegister(REG_VERSION) != 0x12 {
		return errors.New("SX127x module not found")
	}

	d.sleep()
	// println(d.getFrequency())

	// set base addresses
	d.writeRegister(REG_FIFO_TX_BASE_ADDR, 0)
	d.writeRegister(REG_FIFO_RX_BASE_ADDR, 0)

	// set LNA boost
	d.writeRegister(REG_LNA, d.readRegister(REG_LNA)|0x03)

	// set auto AGC
	d.writeRegister(REG_MODEM_CONFIG_3, 0x04)

	err = d.ReConfigure(cfg)

	d.idle()

	return err
}

func (d *Device) ReConfigure(cfg Config) (err error) {
	if cfg.Frequency != 0 {
		d.setFrequency(cfg.Frequency)
	}
	if cfg.SpreadingFactor != 0 {
		d.setSpreadingFactor(cfg.SpreadingFactor)
	}
	if cfg.Bandwidth != 0 {
		d.setBandwidth(cfg.Bandwidth)
	}
	if cfg.CodingRate != 0 {
		d.setCodingRate(cfg.CodingRate)
	}
	if cfg.TxPower != 0 {
		d.setTxPower(cfg.TxPower)
	}
	return err
}

// ReadTemperature returns the temperature in celsius milli degrees (ºC/1000).
func (d *Device) SendPacket(packet []byte) {
}

func (d *Device) LastPacketRSSI() uint8 {
	// section 5.5.5
	var adjustValue uint8 = 157
	if d.getFrequency() < 868000000 {
		adjustValue = 164
	}
	return d.readRegister(REG_PKT_RSSI_VALUE) - adjustValue
}

func (d *Device) LastPacketSNR() uint8 {
	return uint8(d.readRegister(REG_PKT_SNR_VALUE) / 4)
}

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

func (d *Device) PrintRegisters() {
	for i := 0; i < 128; i++ {
		fmt.Printf("%02x: %02x\n", i, d.readRegister(uint8(i)))
	}
}

func (d *Device) reset() {
	d.rstPin.Low()
	time.Sleep(10 * time.Millisecond)
	d.rstPin.High()
	time.Sleep(10 * time.Millisecond)
}

func (d *Device) sleep() {
	d.writeRegister(REG_OP_MODE, MODE_LONG_RANGE_MODE|MODE_SLEEP)
}

func (d *Device) idle() {
	d.writeRegister(REG_OP_MODE, MODE_LONG_RANGE_MODE|MODE_STDBY)
}

func (d *Device) getFrequency() uint32 {
	var f uint64 = uint64(d.readRegister(REG_FRF_LSB))
	f += uint64(d.readRegister(REG_FRF_MID)) << 8
	f += uint64(d.readRegister(REG_FRF_MSB)) << 16
	f = (f * 32000000) >> 19
	return uint32(f)
}

func (d *Device) setFrequency(frequency uint32) {
	var frf uint64 = (uint64(frequency) << 19) / 32000000
	d.writeRegister(REG_FRF_MSB, uint8(frf>>16))
	d.writeRegister(REG_FRF_MID, uint8(frf>>8))
	d.writeRegister(REG_FRF_LSB, uint8(frf>>0))
}

func (d *Device) getSpreadingFactor() uint8 {
	return 0
}
func (d *Device) setSpreadingFactor(spreadingFactor uint8) {
}
func (d *Device) getBandwidth() int32 {
	return 0
}
func (d *Device) setBandwidth(sbw int32) {
	var bw uint8

	if sbw <= 7800 {
		bw = 0
	} else if sbw <= 10400 {
		bw = 1
	} else if sbw <= 15600 {
		bw = 2
	} else if sbw <= 20800 {
		bw = 3
	} else if sbw <= 312500 {
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
	var symbolDuration int32 = 1000 / (d.getBandwidth() / (1 << d.getSpreadingFactor()))

	var config3 uint8 = d.readRegister(REG_MODEM_CONFIG_3)

	// Section 4.1.1.6
	if symbolDuration > 16 {
		config3 = config3 | 0x08
	} else {
		config3 = config3 & 0xF7
	}

	d.writeRegister(REG_MODEM_CONFIG_3, config3)
}

func (d *Device) setCodingRate(denominator uint8) {
	if denominator < 5 {
		denominator = 5
	} else if denominator > 8 {
		denominator = 8
	}

	var cr = denominator - 4

	d.writeRegister(REG_MODEM_CONFIG_1, (d.readRegister(REG_MODEM_CONFIG_1)&0xf1)|(cr<<1))
}

func (d *Device) setTxPower(txPower int8) {
}

func (d *Device) readRegister(reg uint8) uint8 {
	d.csPin.Low()
	d.bus.Tx([]byte{reg & 0x7f}, nil)
	var value [1]byte
	d.bus.Tx(nil, value[:])
	d.csPin.High()
	return value[0]
}

func (d *Device) writeRegister(reg uint8, value uint8) uint8 {
	var response [1]byte
	d.csPin.Low()
	d.bus.Tx([]byte{reg | 0x80}, nil)
	d.bus.Tx([]byte{value}, response[:])
	d.csPin.High()
	return response[0]
}
