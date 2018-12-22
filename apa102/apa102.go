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
	tx    []byte
	count int
	Order int
}

// New returns a new APA102 driver. Pass in a fully configured SPI bus, and the count of
// APA102 LEDs that are connected together.
func New(b machine.SPI, count int) Device {
	t := make([]byte, count*4)
	return Device{bus: b, tx: t, count: count, Order: BGR}
}

// WriteColors writes the given RGBA color slice out using the APA102 protocol.
// The A value (Alpha channel) is used for brightness, set to 0xff (255) for maximum.
func (d Device) WriteColors(cs []color.RGBA) error {
	for i, c := range cs {
		d.tx[i*4] = 0xe0 | (c.A >> 3) // brightness is scaled to 5 bit value
		switch d.Order {
		case BRG:
			d.tx[i*4+1] = byte(c.B)
			d.tx[i*4+2] = byte(c.R)
			d.tx[i*4+3] = byte(c.G)
		case GRB:
			d.tx[i*4+1] = byte(c.G)
			d.tx[i*4+2] = byte(c.R)
			d.tx[i*4+3] = byte(c.B)
		case BGR:
			d.tx[i*4+1] = byte(c.B)
			d.tx[i*4+2] = byte(c.G)
			d.tx[i*4+3] = byte(c.R)
		}
	}

	return d.Write(d.tx)
}

// Write the raw bytes using the APA102 protocol.
func (d Device) Write(buf []byte) error {
	// start frame
	d.bus.Tx([]byte{0x00, 0x00, 0x00, 0x00}, nil)

	// data
	d.bus.Tx(buf, nil)

	// end frame
	for i := 0; i < (d.count+15)/16; i++ {
		d.bus.Tx([]byte{0xff}, nil)
	}

	return nil
}
