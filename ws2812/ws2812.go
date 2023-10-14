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

type DeviceType uint8

const (
	WS2812 DeviceType = iota // RGB
	SK6812                   // RGBA / RGBW
)

// Device wraps a pin object for an easy driver interface.
type Device struct {
	Pin        machine.Pin
	DeviceType DeviceType
}

// New returns a new WS2812 driver. It does not touch the pin object: you have
// to configure it as an output pin before calling New.
// WS2812 is for RGB, SK6812 is for RGBA
func New(pin machine.Pin, deviceType DeviceType) Device {
	return Device{
		Pin:        pin,
		DeviceType: deviceType,
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
	switch d.DeviceType {
	case WS2812:
		err = d.writeColorsRGB(buf)
	case SK6812:
		err = d.writeColorsRGBA(buf)
	}
	return
}

func (d Device) writeColorsRGB(buf []color.RGBA) (err error) {
	for _, color := range buf {
		d.WriteByte(color.G)       // green
		d.WriteByte(color.R)       // red
		err = d.WriteByte(color.B) // blue
	}
	return
}

func (d Device) writeColorsRGBA(buf []color.RGBA) (err error) {
	for _, color := range buf {
		d.WriteByte(color.G)       // green
		d.WriteByte(color.R)       // red
		d.WriteByte(color.B)       // blue
		err = d.WriteByte(color.A) // alpha
	}
	return
}
