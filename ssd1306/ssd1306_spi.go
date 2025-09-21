package ssd1306

import (
	"time"

	"tinygo.org/x/drivers"
)

type SPIBus struct {
	wire     drivers.SPI
	dcPin    drivers.PinOutput
	resetPin drivers.PinOutput
	csPin    drivers.PinOutput
	buffer   []byte // buffer to avoid heap allocations
}

// NewSPI creates a new SSD1306 connection. The SPI wire must already be configured.
func NewSPI(bus drivers.SPI, dcPin, resetPin, csPin drivers.PinOutput) *Device {
	return &Device{
		bus: &SPIBus{
			wire:     bus,
			dcPin:    drivers.SafePinOutput(dcPin),
			resetPin: drivers.SafePinOutput(resetPin),
			csPin:    drivers.SafePinOutput(csPin),
		},
	}
}

// configure pins with the SPI bus and allocate the buffer
func (b *SPIBus) configure(address uint16, size int16) []byte {
	b.csPin.Set(false)
	b.dcPin.Set(false)
	b.resetPin.Set(false)

	b.resetPin.Set(true)
	time.Sleep(1 * time.Millisecond)
	b.resetPin.Set(false)
	time.Sleep(10 * time.Millisecond)
	b.resetPin.Set(true)

	b.buffer = make([]byte, size+1) // +1 for a command
	return b.buffer[1:]             // return the image buffer
}

// command sends a command to the display
func (b *SPIBus) command(cmd uint8) error {
	b.buffer[0] = cmd
	return b.tx(b.buffer[:1], true)
}

// flush sends the image to the display
func (b *SPIBus) flush() error {
	return b.tx(b.buffer[1:], false)
}

// tx sends data to the display
func (b *SPIBus) tx(data []byte, isCommand bool) error {
	b.csPin.Set(true)
	b.dcPin.Set(!isCommand)
	b.csPin.Set(false)
	err := b.wire.Tx(data, nil)
	b.csPin.Set(true)
	return err
}
