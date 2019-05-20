// Package at24cx provides a driver for the AT24C32/64 2-wire serial EEPROM
//
// Datasheet:
// https://www.openimpulse.com/blog/wp-content/uploads/wpsc/downloadables/24C32-Datasheet.pdf
package at24cx

import (
	"machine"
	"time"
)

// Device wraps an I2C connection to a DS3231 device.
type Device struct {
	bus      machine.I2C
	Address  uint16
	pageSize uint16
}

type Config struct {
	PageSize uint16
}

// New creates a new AT24C32/64 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus machine.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
	}
}

// Configure sets up the device for communication
func (d *Device) Configure(cfg Config) {
	if cfg.PageSize == 0 {
		d.pageSize = 32
	} else {
		d.pageSize = cfg.PageSize
	}
}

// WriteByte writes a byte at the specified address
func (d *Device) WriteByte(eepromAddress uint16, value uint8) error {
	address := []uint8{
		uint8((eepromAddress >> 8) & 0xFF),
		uint8(eepromAddress & 0xFF),
		value,
	}
	return d.bus.Tx(d.Address, address, nil)
}

// ReadByte reads the byte at the specified address
func (d *Device) ReadByte(eepromAddress uint16) (uint8, error) {
	address := []uint8{
		uint8(eepromAddress >> 8),
		uint8(eepromAddress & 0xFF),
	}
	data := make([]uint8, 1)
	err := d.bus.Tx(d.Address, address, data)
	return data[0], err
}

// WriteBytes writes a byte array at the specified address
func (d *Device) WriteBytes(eepromAddress uint16, data []uint8) error {
	values := make([]uint8, 32)
	dataLeft := uint16(len(data))
	offsetData := uint16(0)
	var offsetPage uint16
	var chunkLength uint16
	for dataLeft > 0 {
		offsetPage = eepromAddress % d.pageSize
		if dataLeft < 30 { // The 32K/64K EEPROM is capable of 32-byte page writes and we're using 2 for the address
			chunkLength = dataLeft
		} else {
			chunkLength = 30
		}
		if (d.pageSize - offsetPage) < chunkLength {
			chunkLength = d.pageSize - offsetPage
		}
		for i := uint16(0); i < chunkLength; i++ {
			values[2+i] = data[offsetData+i]
		}
		values[0] = uint8(eepromAddress >> 8)
		values[1] = uint8(eepromAddress & 0xFF)
		err := d.bus.Tx(d.Address, values[:chunkLength+2], nil)
		if err != nil {
			return err
		}
		dataLeft -= chunkLength
		offsetData += chunkLength
		eepromAddress += chunkLength
		time.Sleep(2 * time.Millisecond) // writing again too soon will block the device
	}
	return nil
}

// ReadByte reads the bytes at the specified address
func (d *Device) ReadBytes(eepromAddress uint16, data []uint8) error {
	address := []uint8{
		uint8((eepromAddress >> 8) & 0xFF),
		uint8(eepromAddress & 0xFF),
	}
	err := d.bus.Tx(d.Address, address, data)
	return err
}
