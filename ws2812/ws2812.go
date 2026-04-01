// Package ws2812 implements a driver for WS2812 and SK6812 RGB LED strips.
//
// On most platforms NewWS2812 uses bit-banging.
// On RP2040/RP2350 it uses PIO for hardware-timed control.
package ws2812 // import "tinygo.org/x/drivers/ws2812"

//go:generate go run gen-ws2812.go -arch=cortexm 16 48 64 120 125 150 168 200
//go:generate go run gen-ws2812.go -arch=tinygoriscv 160 320

import (
	"errors"
	"image/color"
	"machine"
)

var errUnknownClockSpeed = errors.New("ws2812: unknown CPU clock speed")

// Device wraps a pin object for an easy driver interface.
type Device struct {
	Pin            machine.Pin
	brightness     uint8
	writeColorFunc func(Device, []color.RGBA, uint8) error
}

// deprecated, use NewWS2812 or NewSK6812 depending on which device you want.
// calls NewWS2812() to avoid breaking everyone's existing code.
func New(pin machine.Pin) Device {
	return NewWS2812(pin)
}

// NewWS2812 returns a new WS2812(RGB) driver.
// On RP2040/RP2350, it uses PIO for hardware-timed control.
// On other platforms, you must configure the pin as output before calling this.
func NewWS2812(pin machine.Pin) Device {
	return newWS2812Device(pin)
}

// New returns a new SK6812(RGBA) driver.
// It does not touch the pin object: you have
// to configure it as an output pin before calling New.
func NewSK6812(pin machine.Pin) Device {
	return Device{
		Pin:            pin,
		brightness:     255,
		writeColorFunc: writeColorsRGBA,
	}
}

// SetBrightness sets the global brightness (0-255).
func (d *Device) SetBrightness(b uint8) {
	d.brightness = b
}

// Write the raw bitstring out using the WS2812 protocol.
func (d Device) Write(buf []byte) (n int, err error) {
	for _, c := range buf {
		d.WriteByte(c)
	}
	return len(buf), nil
}

// Write the given color slice out using the WS2812 protocol.
// Colors are sent out in the usual GRB(A) format.
func (d Device) WriteColors(buf []color.RGBA) (err error) {
	return d.writeColorFunc(d, buf, d.brightness)
}

func writeColorsRGB(d Device, buf []color.RGBA, brightness uint8) (err error) {
	for _, color := range buf {
		r, g, b := applyBrightness(color, brightness)
		d.WriteByte(g)       // green
		d.WriteByte(r)       // red
		err = d.WriteByte(b) // blue
	}
	return
}

func writeColorsRGBA(d Device, buf []color.RGBA, brightness uint8) (err error) {
	for _, color := range buf {
		r, g, b := applyBrightness(color, brightness)

		d.WriteByte(g)             // green
		d.WriteByte(r)             // red
		d.WriteByte(b)             // blue
		err = d.WriteByte(color.A) // alpha
	}
	return
}

// applyBrightness scales a color by the brightness value.
func applyBrightness(c color.RGBA, brightness uint8) (r, g, b uint8) {
	r = uint8((uint16(c.R) * uint16(brightness)) >> 8)
	g = uint8((uint16(c.G) * uint16(brightness)) >> 8)
	b = uint8((uint16(c.B) * uint16(brightness)) >> 8)
	return
}
