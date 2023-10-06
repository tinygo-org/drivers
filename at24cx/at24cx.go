// Package at24cx provides a driver for the AT24C32/64/128/256/512 2-wire serial EEPROM
//
// Datasheet:
// https://www.openimpulse.com/blog/wp-content/uploads/wpsc/downloadables/24C32-Datasheet.pdf
package at24cx // import "tinygo.org/x/drivers/at24cx"

import (
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

// Device wraps an I2C connection to an AT24CX device.
type Device struct {
	bus               drivers.I2C
	Address           uint16
	pageSize          uint16
	currentRAMAddress uint16
	startRAMAddress   uint16
	endRAMAddress     uint16
}

type Config struct {
	PageSize        uint16
	StartRAMAddress uint16
	EndRAMAddress   uint16
}

// New creates a new AT24C32/64 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
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
	if cfg.EndRAMAddress == 0 {
		d.endRAMAddress = 4096
	} else {
		d.endRAMAddress = cfg.EndRAMAddress
	}
	d.startRAMAddress = cfg.StartRAMAddress
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

// WriteAt writes a byte array at the specified address
func (d *Device) WriteAt(data []byte, offset int64) (n int, err error) {
	return d.writeAt(data, uint16(offset))
}

// writeAt writes a byte array at the specified address
func (d *Device) writeAt(data []byte, offset uint16) (n int, err error) {
	values := make([]uint8, 32)
	dataLeft := uint16(len(data))
	d.currentRAMAddress = offset
	offset = 0
	var offsetPage uint16
	var chunkLength uint16
	for dataLeft > 0 {
		offsetPage = d.currentRAMAddress % d.pageSize
		if dataLeft < 30 { // The 32K/64K EEPROM is capable of 32-byte page writes and we're using 2 for the address
			chunkLength = dataLeft
		} else {
			chunkLength = 30
		}
		if (d.pageSize - offsetPage) < chunkLength {
			chunkLength = d.pageSize - offsetPage
		}
		for i := uint16(0); i < chunkLength; i++ {
			values[2+i] = data[offset+i]
		}
		values[0] = uint8(d.currentRAMAddress >> 8)
		values[1] = uint8(d.currentRAMAddress & 0xFF)
		err := d.bus.Tx(d.Address, values[:chunkLength+2], nil)
		if err != nil {
			return 0, err
		}
		dataLeft -= chunkLength
		offset += chunkLength
		if d.endRAMAddress-chunkLength < d.currentRAMAddress {
			d.currentRAMAddress = d.startRAMAddress + (d.currentRAMAddress+uint16(len(data)))%d.endRAMAddress
		} else {
			d.currentRAMAddress += chunkLength
		}
		time.Sleep(2 * time.Millisecond) // writing again too soon will block the device
	}
	return len(data), nil
}

// ReadAt reads the bytes at the specified address
func (d *Device) ReadAt(data []byte, offset int64) (n int, err error) {
	return d.readAt(data, uint16(offset))
}

// readAt reads the bytes at the specified address
func (d *Device) readAt(data []byte, offset uint16) (n int, err error) {
	address := []uint8{
		uint8((offset >> 8) & 0xFF),
		uint8(offset & 0xFF),
	}
	err = d.bus.Tx(d.Address, address, data)

	if d.endRAMAddress-uint16(len(data)) < offset {
		d.currentRAMAddress = d.startRAMAddress + (offset+uint16(len(data)))%d.endRAMAddress
	} else {
		d.currentRAMAddress = offset + uint16(len(data))
	}
	return len(data), err
}

// Seek sets the offset for the next Read or Write on SRAM to offset, interpreted
// according to whence: 0 means relative to the origin of the SRAM, 1 means
// relative to the current offset, and 2 means relative to the end.
// returns new offset and error, if any
func (d *Device) Seek(offset int64, whence int) (int64, error) {
	w := uint16(0)
	switch whence {
	case 0:
		w = d.startRAMAddress
	case 1:
		w = d.currentRAMAddress
	case 2:
		w = d.endRAMAddress
	default:
		return 0, errors.New("invalid whence")
	}
	d.currentRAMAddress = w + uint16(offset)
	return int64(d.currentRAMAddress), nil
}

// Write writes len(data) bytes to SRAM
// returns number of bytes written and error, if any
func (d *Device) Write(data []byte) (n int, err error) {
	return d.writeAt(data, d.currentRAMAddress)
}

// Read reads len(data) from SRAM
// returns number of bytes written and error, if any
func (d *Device) Read(data []uint8) (n int, err error) {
	return d.readAt(data, d.currentRAMAddress)
}
