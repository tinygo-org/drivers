// Package mcp9808 implements a driver for the MCP9808 High Accuracy I2C Temperature Sensor
//
// Datasheet: https://cdn-shop.adafruit.com/datasheets/MCP9808.pdf
// Module: https://www.adafruit.com/product/1782
// Only implemented: temperature reading, resolution read & set
package mcp9808

import (
	"encoding/binary"
	"errors"

	"tinygo.org/x/drivers"
)

type Device struct {
	bus     drivers.I2C
	Address uint16
}

func New(bus drivers.I2C) Device {
	return Device{bus, MCP9808_I2CADDR_DEFAULT}
}

func (d *Device) Connected() bool {
	data := make([]byte, 2)
	d.Read(MCP9808_REG_DEVICE_ID, &data)
	return binary.BigEndian.Uint16(data) == MCP9808_DEVICE_ID
}

func (d *Device) ReadTemperature() (float64, error) {
	data := make([]byte, 2)
	var temp float64
	if err := d.Read(MCP9808_REG_AMBIENT_TEMP, &data); err != nil {
		return 0, err
	}

	data[0] = data[0] & 0x1F
	if data[0]&0x10 == 0x10 {
		data[0] = data[0] & 0x0F
		temp = float64(data[0])*16 + float64(data[1])/16.0 - 256
	}
	temp = float64(data[0])*16 + float64(data[1])/16.0
	return temp, nil
}

func (d *Device) ReadResolution() (resolution, error) {
	data := make([]byte, 2)
	err := d.Read(MCP9808_REG_RESOLUTION, &data)
	if err != nil {
		return 0, err
	}
	switch data[0] {
	case 0:
		return Low, nil
	case 1:
		return Medium, nil
	case 2:
		return High, nil
	case 3:
		return Maximum, nil

	default:
		return 0, errors.New("unknown resolution")
	}
}

func (d *Device) SetResolution(r resolution) error {
	switch r {
	case Low:
		if err := d.Write(MCP9808_REG_RESOLUTION, []byte{0x00}); err != nil {
			return err
		}
	case Medium:
		if err := d.Write(MCP9808_REG_RESOLUTION, []byte{0x01}); err != nil {
			return err
		}
	case High:
		if err := d.Write(MCP9808_REG_RESOLUTION, []byte{0x02}); err != nil {
			return err
		}
	case Maximum:
		if err := d.Write(MCP9808_REG_RESOLUTION, []byte{0x03}); err != nil {
			return err
		}
	default:
		return nil
	}
	return nil
}

func (d *Device) Write(register byte, data []byte) error {
	buf := append([]byte{register}, data...)
	return d.bus.Tx(d.Address, buf, nil)
}

func (d *Device) Read(register byte, data *[]byte) error {
	return d.bus.Tx(d.Address, []byte{register}, *data)
}
