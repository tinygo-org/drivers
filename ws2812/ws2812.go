// Package ws2812 implements a driver for WS2812 and SK6812 RGB LED strips.
package ws2812

import (
	"machine"
)

// Device wraps a pin object for an easy driver interface.
type Device struct {
	Pin machine.GPIO
}

// New returns a new WS2812 driver. It does not touch the pin object: you have
// to configure it as an output pin before calling New.
func New(pin machine.GPIO) Device {
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
// Colors are specified in RGB format, and are send out in the common GRB
// format.
func (d Device) WriteColors(buf []uint32) {
	for _, color := range buf {
		d.WriteByte(byte(color >> 8))  // green
		d.WriteByte(byte(color >> 16)) // red
		d.WriteByte(byte(color >> 0))  // blue
	}
}
