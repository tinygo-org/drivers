package ssd1306

import (
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
	"tinygo.org/x/drivers/internal/pin"
)

type SPIBus struct {
	wire     drivers.SPI
	dcPin    pin.OutputFunc
	resetPin pin.OutputFunc
	csPin    pin.OutputFunc
	buffer   []byte // buffer to avoid heap allocations
}

// NewSPI creates a new SSD1306 connection. The SPI wire must already be configured.
func NewSPI(bus drivers.SPI, dcPin, resetPin, csPin pin.Output) *Device {
	// configure GPIO pins (on baremetal targets only, for backwards compatibility)
	legacy.ConfigurePinOut(dcPin)
	legacy.ConfigurePinOut(resetPin)
	legacy.ConfigurePinOut(csPin)
	return &Device{
		bus: &SPIBus{
			wire:     bus,
			dcPin:    dcPin.Set,
			resetPin: resetPin.Set,
			csPin:    csPin.Set,
		},
	}
}

// configure pins with the SPI bus and allocate the buffer
func (b *SPIBus) configure(address uint16, size int16) []byte {
	b.csPin.Low()
	b.dcPin.Low()
	b.resetPin.Low()

	b.resetPin.High()
	time.Sleep(1 * time.Millisecond)
	b.resetPin.Low()
	time.Sleep(10 * time.Millisecond)
	b.resetPin.High()

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
	b.csPin.High()
	b.dcPin(!isCommand)
	b.csPin.Low()
	err := b.wire.Tx(data, nil)
	b.csPin.High()
	return err
}
