// Package mcp9808 implements a driver for the MCP9808 High Accuracy I2C Temperature Sensor
//
// Datasheet: https://cdn-shop.adafruit.com/datasheets/MCP9808.pdf
// Module: https://www.adafruit.com/product/1782
package mcp9808

import (
	"encoding/binary"
	"math"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

type Device struct {
	bus     drivers.I2C
	buf     []byte
	Address uint16
}

func New(bus drivers.I2C) Device {
	return Device{bus, make([]byte, 3), MCP9808_I2CADDR_DEFAULT}
}

func (d *Device) Connected() bool {
	d.Read(MCP9808_REG_DEVICE_ID, d.buf)
	return binary.BigEndian.Uint16(d.buf[:1]) == MCP9808_DEVICE_ID
}

func (d *Device) Temperature() (float64, error) {
	d.buf[0] = MCP9808_REG_AMBIENT_TEMP
	if err := d.Write(d.buf[0], binary.BigEndian.Uint16(d.buf[:1])); err != nil {
		return 0, err
	}
	if err := d.Read(d.buf[0], d.buf[1:]); err != nil {
		return 0, err
	}

	return d.tempConv(), nil
}

func (d *Device) tempConv() float64 {
	d.buf[1] = d.buf[1] & 0x1F
	if d.buf[1]&0x10 == 0x10 {
		d.buf[1] = d.buf[1] & 0x0F
		return (float64(d.buf[1])*16 + float64(d.buf[2])/16.0) - 256
	}
	return float64(d.buf[1])*16 + float64(d.buf[2])/16.0
}

func (d *Device) limitTemperatures(temp int, tAddress byte) error {
	var negative bool
	if temp < 0 {
		negative = true
		temp = int(math.Abs(float64(temp)))
	}

	d.buf[0] = tAddress

	d.buf[1] = byte(temp >> 4)
	if negative {
		d.buf[1] |= 0x10
	}

	d.buf[2] = byte((temp & 0x0F) << 4)

	_, err := d.i2cDevice.Write(d.buf)
	return err
}

func (d *Device) getTemperature(address byte) (float64, error) {
	d.buf[0] = address
	if _, err := d.i2cDevice.Write(d.buf[:1]); err != nil {
		return 0, err
	}
	if _, err := d.i2cDevice.Read(d.buf[1:]); err != nil {
		return 0, err
	}

	return d.tempConv(), nil
}

func (d *Device) setTemperature(temp int, address byte) error {
	return d.limitTemperatures(temp, address)
}

func (d *Device) UpperTemperature() (float64, error) {
	return d.getTemperature(MCP9808_REG_UPPER_TEMP)
}

func (d *Device) SetUpperTemperature(temp int) error {
	return d.limitTemperatures(temp, MCP9808_REG_UPPER_TEMP)
}

func (d *Device) LowerTemperature() (float64, error) {
	return d.getTemperature(MCP9808_REG_LOWER_TEMP)
}

func (d *Device) SetLowerTemperature(temp int) error {
	return d.limitTemperatures(temp, MCP9808_REG_LOWER_TEMP)
}

func (d *Device) CriticalTemperature() (float64, error) {
	return d.getTemperature(MCP9808_REG_CRITICAL_TEMP)
}

func (d *Device) SetCriticalTemperature(temp int) error {
	return d.limitTemperatures(temp, MCP9808_REG_CRITICAL_TEMP)
}

func (d *Device) Resolution() resolution {
	return d.getRWBits(2, MCP9808_REG_RESOLUTION, 0)
}

func (d *Device) SetResolution(r resolution) error {
	return d.setRWBits(2, MCP9808_REG_RESOLUTION, 0, r)
}

func (d *Device) getRWBits(bitCount int, register byte, startBit int) int {
	// Implement the getRWBits functionality
	return 0
}

func (d *Device) setRWBits(bitCount int, register byte, startBit int, value int) error {
	// Implement the setRWBits functionality
	return nil
}

// Convenience method to read the register and avoid repetition.
func (d *Device) Read(reg uint8, buf []byte) error {
	return legacy.ReadRegister(d.bus, uint8(d.Address), reg, buf)
}

// Convenience method to write the register and avoid repetition.
func (d *Device) Write(reg uint8, v uint16) error {
	data := []byte{byte(v)}
	err := legacy.WriteRegister(d.bus, uint8(d.Address), reg, data)
	return err
}
