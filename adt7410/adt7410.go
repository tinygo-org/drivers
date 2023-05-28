// Package adt7410 provides a driver for the adt7410 I2C Temperature Sensor.
//
// Datasheet: https://www.analog.com/media/en/technical-documentation/data-sheets/ADT7410.pdf
package adt7410 // import "tinygo.org/x/drivers/adt7410"

import (
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

type Error uint8

const (
	ErrInvalidID Error = 0x1
)

func (e Error) Error() string {
	switch e {
	case ErrInvalidID:
		return "Invalid chip ID"
	default:
		return "Unknown error"
	}
}

type Device struct {
	bus     drivers.I2C
	buf     []byte
	Address uint8
}

// New returns ADT7410 device for the provided I2C bus using default address.
// of 0x48 (1001000).  To use multiple ADT7410 devices, the last 2 bits of the address
// can be set using by connecting to the A1 and A0 pins to VDD or GND (for a
// total of up to 4 devices on a I2C bus).  Also note that 10k pullups are
// recommended for the SDA and SCL lines.
func New(i2c drivers.I2C) *Device {
	return &Device{
		bus:     i2c,
		buf:     make([]byte, 2),
		Address: Address,
	}
}

// Configure the ADT7410 device.
func (d *Device) Configure() (err error) {
	// reset the chip
	d.writeByte(RegReset, 0xFF)
	time.Sleep(10 * time.Millisecond)
	return
}

// Connected returns whether sensor has been found.
func (d *Device) Connected() bool {
	data := []byte{0}
	legacy.ReadRegister(d.bus, uint8(d.Address), RegID, data)
	return data[0]&0xF8 == 0xC8
}

// ReadTemperature returns the temperature in celsius milli degrees (°C/1000)
func (d *Device) ReadTemperature() (temperature int32, err error) {
	return (int32(d.readUint16(RegTempValueMSB)) * 1000) / 128, nil
}

// ReadTempC returns the value in the temperature value register, in Celsius.
func (d *Device) ReadTempC() float32 {
	t := d.readUint16(RegTempValueMSB)
	return float32(int(t)) / 128.0
}

// ReadTempF returns the value in the temperature value register, in Fahrenheit.
func (d *Device) ReadTempF() float32 {
	return d.ReadTempC()*1.8 + 32.0
}

func (d *Device) writeByte(reg uint8, data byte) {
	d.buf[0] = reg
	d.buf[1] = data
	d.bus.Tx(uint16(d.Address), d.buf, nil)
}

func (d *Device) readByte(reg uint8) byte {
	legacy.ReadRegister(d.bus, d.Address, reg, d.buf)
	return d.buf[0]
}

func (d *Device) readUint16(reg uint8) uint16 {
	legacy.ReadRegister(d.bus, d.Address, reg, d.buf)
	return uint16(d.buf[0])<<8 | uint16(d.buf[1])
}
