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
	SpreadingFactor int8
	Bandwidth       int32
	CodingRate      int8
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

func (d *Device) setSpreadingFactor(spreadingFactor int8) {
}
func (d *Device) setBandwidth(bandwidth int32) {
}
func (d *Device) setCodingRate(codingRate int8) {
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
