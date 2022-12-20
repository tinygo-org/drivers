//go:build !baremetal

package ws2812

// This file implements the WS2812 protocol for simulation.

import "machine"

// Send a single byte using the WS2812 protocol.
func (d Device) WriteByte(c byte) error {
	writeByte(d.Pin, c)
	return nil
}

//go:export __tinygo_ws2812_write_byte
func writeByte(pin machine.Pin, c byte)
