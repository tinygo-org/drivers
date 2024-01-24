// Package ws2812 implements a driver for WS2812 and SK6812 RGB LED strips.
package ws2812 // import "tinygo.org/x/drivers/ws2812"

//go:generate go run gen-ws2812.go -arch=cortexm 16 48 64 120 125 168
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
	writeColorFunc func(Device, []color.RGBA) error
}

// deprecated, use NewWS2812 or NewSK6812 depending on which device you want.
// calls NewWS2812() to avoid breaking everyone's existing code.
func New(pin machine.Pin) Device {
	return NewWS2812(pin)
}

// New returns a new WS2812(RGB) driver.
// It does not touch the pin object: you have
// to configure it as an output pin before calling New.
func NewWS2812(pin machine.Pin) Device {
	return Device{
		Pin:            pin,
		writeColorFunc: writeColorsRGB,
	}
}

// New returns a new SK6812(RGBA) driver.
// It does not touch the pin object: you have
// to configure it as an output pin before calling New.
func NewSK6812(pin machine.Pin) Device {
	return Device{
		Pin:            pin,
		writeColorFunc: writeColorsRGBA,
	}
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
	return d.writeColorFunc(d, buf)
}

func writeColorsRGB(d Device, buf []color.RGBA) (err error) {
	for _, color := range buf {
		d.WriteByte(color.G)       // green
		d.WriteByte(color.R)       // red
		err = d.WriteByte(color.B) // blue
	}
	return
}

func writeColorsRGBA(d Device, buf []color.RGBA) (err error) {
	for _, color := range buf {
		d.WriteByte(color.G)       // green
		d.WriteByte(color.R)       // red
		d.WriteByte(color.B)       // blue
		err = d.WriteByte(color.A) // alpha
	}
	return
}
