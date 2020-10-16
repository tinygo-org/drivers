// Package ws2812 implements a driver for WS2812 and SK6812 RGB LED strips.
package ws2812 // import "tinygo.org/x/drivers/ws2812"

import (
	"errors"
	"image/color"
	"machine"
)

var errUnknownClockSpeed = errors.New("ws2812: unknown CPU clock speed")

// Device wraps a pin object for an easy driver interface.
type Device struct {
	Pin machine.Pin
}

// New returns a new WS2812 driver. It does not touch the pin object: you have
// to configure it as an output pin before calling New.
func New(pin machine.Pin) Device {
	return Device{pin}
}

// Write the raw bitstring out using the WS2812 protocol.
func (d Device) Write(buf []byte) (n int, err error) {
	for _, c := range buf {
		d.WriteByte(c)
	}
	return len(buf), nil
}

// Write the given color slice out using the WS2812 protocol.
// Colors are sent out in the usual GRB format.
func (d Device) WriteColors(buf []color.RGBA) error {
	for _, color := range buf {
		d.WriteByte(color.G) // green
		d.WriteByte(color.R) // red
		d.WriteByte(color.B) // blue
	}
	return nil
}
