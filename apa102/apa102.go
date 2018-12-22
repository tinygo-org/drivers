// Package apa102 implements a driver for the APA102 SPI LED.
//
// Datasheet: https://cdn-shop.adafruit.com/product-files/2343/APA102C.pdf
package apa102

import (
	"image/color"
	"machine"
)

const (
	// BGR aka "Blue Green Red" is the current APA102 LED color order.
	BGR = iota

	// BRG aka "Blue Red Green" is the typical APA102 color order from 2015-2017.
	BRG

	// GRB aka "Green Red Blue" is the typical APA102 color order from pre-2015.
	GRB
)

// Device wraps APA102 SPI LEDs.
type Device struct {
	bus   machine.SPI
	count int
	Order int
}

// New returns a new APA102 driver. Pass in a fully configured SPI bus, and the count of
// APA102 LEDs that are connected together.
func New(b machine.SPI, count int) Device {
	return Device{bus: b, count: count, Order: BGR}
}

// WriteColors writes the given RGBA color slice out using the APA102 protocol.
// The A value (Alpha channel) is used for brightness, set to 0xff (255) for maximum.
func (d Device) WriteColors(cs []color.RGBA) (n int, err error) {
	d.startFrame()

	// write data
	for _, c := range cs {
		// brightness is scaled to 5 bit value
		d.bus.Tx([]byte{0xe0 | (c.A >> 3)}, nil)

		// set the colors
		switch d.Order {
		case BRG:
			d.bus.Tx([]byte{c.B}, nil)
			d.bus.Tx([]byte{c.R}, nil)
			d.bus.Tx([]byte{c.G}, nil)
		case GRB:
			d.bus.Tx([]byte{c.G}, nil)
			d.bus.Tx([]byte{c.R}, nil)
			d.bus.Tx([]byte{c.B}, nil)
		case BGR:
			d.bus.Tx([]byte{c.B}, nil)
			d.bus.Tx([]byte{c.G}, nil)
			d.bus.Tx([]byte{c.R}, nil)
		}
	}

	d.endFrame()

	return len(cs), nil
}

// Write the raw bytes using the APA102 protocol.
func (d Device) Write(buf []byte) (n int, err error) {
	d.startFrame()
	d.bus.Tx(buf, nil)
	d.endFrame()

	return len(buf), nil
}

func (d Device) startFrame() {
	d.bus.Tx([]byte{0x00, 0x00, 0x00, 0x00}, nil)
}

func (d Device) endFrame() {
	for i := 0; i < (d.count+15)/16; i++ {
		d.bus.Tx([]byte{0xff}, nil)
	}
}
