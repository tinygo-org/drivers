// Package ws2812 implements a driver for WS2812 and SK6812 RGB LED strips.
package ws2812

import (
	"image/color"
	"machine"
)

type ColorOrder string

const (
	RGB  ColorOrder = "RGB"
	GRB             = "GRB"
	GRBW            = "GRBW"
)

// Device wraps a pin object for an easy driver interface.
type Device struct {
	Pin   machine.GPIO
	order ColorOrder
}

// Opt is a functional parameter option
type DeviceOpt func(*Device)

func WithOrder(o ColorOrder) DeviceOpt {
	return func(d *Device) {
		d.order = o
	}
}

type ColorOptions struct {
	Whites []uint8
}

type ColorOpt func(*ColorOptions)

func WithWhite(whites []uint8) ColorOpt {
	return func(co *ColorOptions) {
		co.Whites = whites
	}
}

type ColorRGBWA struct {
	R, G, B, W, A uint8
}

// New returns a new WS2812 driver. It does not touch the pin object: you have
// to configure it as an output pin before calling New.
func New(pin machine.GPIO, opts ...DeviceOpt) Device {
	d := Device{
		Pin: pin,
	}

	// set default color order
	WithOrder(GRB)(&d)

	for _, opt := range opts {
		opt(&d)
	}

	return d
}

// Write the raw bitstring out using the WS2812 protocol.
func (d Device) Write(buf []byte) (n int, err error) {
	for _, c := range buf {
		d.WriteByte(c)
	}
	return len(buf), nil
}

// Write the given color slice out using the WS2812 protocol.
// Colors are sent out in the requested order.
func (d Device) WriteColors(buf []color.RGBA, colorOpt ...ColorOpt) (n int, err error) {
	co := &ColorOptions{}
	for _, opt := range colorOpt {
		opt(co)
	}

	_ = co
	for i, c := range buf {
		// This would probably be cleaner as a call to a function set at
		// initialization time. But
		// https://github.com/tinygo-org/tinygo/issues/349
		switch d.order {
		case RGB:
			d.WriteByte(c.R)
			d.WriteByte(c.G)
			d.WriteByte(c.B)
		case GRB:
			d.WriteByte(c.G)
			d.WriteByte(c.R)
			d.WriteByte(c.B)
		case GRBW:
			d.WriteByte(c.G)
			d.WriteByte(c.R)
			d.WriteByte(c.B)
			_ = i
			if co.Whites != nil && i <= len(co.Whites) {
				d.WriteByte(co.Whites[i])
			} else {
				d.WriteByte(0)
			}
		}

	}

	return len(buf), nil
}

// WriteRGBW writes an RGBWA pixel out.
// TODO for 3 color pixels, this drops the white channel. It could, instead, blend it.
func (d Device) WriteRGBW(buf []ColorRGBWA, colorOpt ...ColorOpt) (n int, err error) {
	co := &ColorOptions{}
	for _, opt := range colorOpt {
		opt(co)
	}

	for _, c := range buf {
		// This would probably be cleaner as a call to a function set at
		// initialization time. But
		// https://github.com/tinygo-org/tinygo/issues/349
		switch d.order {
		case RGB:
			d.WriteByte(c.R)
			d.WriteByte(c.G)
			d.WriteByte(c.B)
		case GRB:
			d.WriteByte(c.G)
			d.WriteByte(c.R)
			d.WriteByte(c.B)
		case GRBW:
			d.WriteByte(c.G)
			d.WriteByte(c.R)
			d.WriteByte(c.B)
			d.WriteByte(c.W)
		}
	}

	return len(buf), nil
}
