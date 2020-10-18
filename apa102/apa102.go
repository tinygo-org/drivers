// Package apa102 implements a driver for the APA102 SPI LED.
//
// Datasheet: https://cdn-shop.adafruit.com/product-files/2343/APA102C.pdf
package apa102 // import "tinygo.org/x/drivers/apa102"

import (
	"image/color"
	"machine"

	"tinygo.org/x/drivers"
)

const (
	// BGR aka "Blue Green Red" is the current APA102 LED color order.
	BGR = iota

	// BRG aka "Blue Red Green" is the typical APA102 color order from 2015-2017.
	BRG

	// GRB aka "Green Red Blue" is the typical APA102 color order from pre-2015.
	GRB
)

var startFrame = []byte{0x00, 0x00, 0x00, 0x00}

// Device wraps APA102 SPI LEDs.
type Device struct {
	bus   drivers.SPI
	Order int
}

// New returns a new APA102 driver. Pass in a fully configured SPI bus.
func New(b drivers.SPI) Device {
	return Device{bus: b, Order: BGR}
}

// NewSoftwareSPI returns a new APA102 driver that will use a software based
// implementation of the SPI protocol.
func NewSoftwareSPI(sckPin, sdoPin machine.Pin, delay uint32) Device {
	return New(&bbSPI{SCK: sckPin, SDO: sdoPin, Delay: delay})
}

// WriteColors writes the given RGBA color slice out using the APA102 protocol.
// The A value (Alpha channel) is used for brightness, set to 0xff (255) for maximum.
func (d Device) WriteColors(cs []color.RGBA) (n int, err error) {
	d.startFrame()

	// write data
	for _, c := range cs {
		// brightness is scaled to 5 bit value
		d.bus.Transfer(0xe0 | (c.A >> 3))

		// set the colors
		switch d.Order {
		case BRG:
			d.bus.Transfer(c.B)
			d.bus.Transfer(c.R)
			d.bus.Transfer(c.G)
		case GRB:
			d.bus.Transfer(c.G)
			d.bus.Transfer(c.R)
			d.bus.Transfer(c.B)
		case BGR:
			d.bus.Transfer(c.B)
			d.bus.Transfer(c.G)
			d.bus.Transfer(c.R)
		}
	}

	d.endFrame(len(cs))

	return len(cs), nil
}

// Write the raw bytes using the APA102 protocol.
func (d Device) Write(buf []byte) (n int, err error) {
	d.startFrame()
	d.bus.Tx(buf, nil)
	d.endFrame(len(buf) / 4)

	return len(buf), nil
}

// startFrame sends the start bytes for a strand of LEDs.
func (d Device) startFrame() {
	d.bus.Tx(startFrame, nil)
}

// endFrame sends the end frame marker with one extra bit per LED so
// long strands of LEDs receive the necessary termination for updates.
// See https://cpldcpu.wordpress.com/2014/11/30/understanding-the-apa102-superled/
func (d Device) endFrame(count int) {
	for i := 0; i < count/16; i++ {
		d.bus.Transfer(0xff)
	}
}
