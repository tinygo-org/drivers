// Package ws2812 implements a driver for WS2812 and SK6812 RGB LED strips.
package ws2812

import (
	"machine"
)

// WS2812 wraps a pin object for an easy driver interface.
type WS2812 struct {
	Pin machine.GPIO
}

// New returns a new WS2812 driver. It does not touch the pin object: you have
// to configure it as an output pin before calling New.
func New(pin machine.GPIO) WS2812 {
	return WS2812{pin}
}

// Write the raw bitstring out using the WS2812 protocol.
func (p WS2812) Write(buf []byte) {
	for _, c := range buf {
		p.WriteByte(c)
	}
}

// Write the given color slice out using the WS2812 protocol.
// Colors are specified in RGB format, and are send out in the common GRB
// format.
func (p WS2812) WriteColors(buf []uint32) {
	for _, color := range buf {
		p.WriteByte(byte(color >> 8))  // green
		p.WriteByte(byte(color >> 16)) // red
		p.WriteByte(byte(color >> 0))  // blue
	}
}
