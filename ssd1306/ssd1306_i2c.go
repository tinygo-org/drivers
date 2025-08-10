package ssd1306

import (
	"tinygo.org/x/drivers"
)

type I2CBus struct {
	wire    drivers.I2C
	address uint16
	buffer  []byte // buffer to avoid heap allocations
}

// NewI2C creates a new SSD1306 connection. The I2C wire must already be configured.
func NewI2C(bus drivers.I2C) *Device {
	return &Device{
		bus: &I2CBus{
			wire:    bus,
			address: Address,
		},
	}
}

// configure address for the I2C bus and allocate the buffer
func (b *I2CBus) configure(address uint16, size int16) []byte {
	if address != 0 {
		b.address = address
	}
	b.buffer = make([]byte, size+2) // +1 for the mode and +1 for a command
	return b.buffer[2:]             // return the image buffer
}

// command sends a command to the display
func (b *I2CBus) command(cmd uint8) error {
	b.buffer[0] = 0x00 // Command mode
	b.buffer[1] = cmd
	return b.wire.Tx(b.address, b.buffer[:2], nil)
}

// flush sends the image to the display
func (b *I2CBus) flush() error {
	b.buffer[1] = 0x40 // Data mode
	return b.wire.Tx(b.address, b.buffer[1:], nil)
}

// tx sends data to the display
func (b *I2CBus) tx(data []byte, isCommand bool) error {
	if isCommand {
		return b.command(data[0])
	}
	copy(b.buffer[2:], data)
	return b.flush()
}
