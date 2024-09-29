// Package ws2812 implements a driver for WS2812 and SK6812 RGB LED strips.
package ws2812 // import "tinygo.org/x/drivers/ws2812"

//go:generate go run gen-ws2812.go -arch=cortexm 16 48 64 120 125 168
//go:generate go run gen-ws2812.go -arch=tinygoriscv 160 320

import (
	"errors"
	"image/color"
	"machine"

	"tinygo.org/x/drivers"
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

// DeviceSPI wraps a SPI object for driving a string of WS2812 LEDs.
type DeviceSPI struct {
	Bus drivers.SPI

	// Use a buffer embedded in the device struct so that at most one allocation
	// happens at NewSPI and no allocation during transmission.
	buf []byte
}

// NewSPI returns a WS2812 driver using a SPI bus. This SPI bus must already be
// configured at exactly 4MHz otherwise WS2812 won't work properly with it.
//
// The advantage of using a SPI bus over bitbanging is that it doesn't require
// custom assembly for each new platform and that it may avoid needing to
// disable interrupts while sending color data if the SPI peripheral uses DMA.
// The disadvantage is of course that it is limited in which pins can be used
// for WS2812 output.
func NewSPI(bus drivers.SPI) *DeviceSPI {
	return &DeviceSPI{
		Bus: bus,
	}
}

// WriteColors wries the given color slice out using the WS2812 protocol.
// Colors are sent out in the usual GRB format.
func (d *DeviceSPI) WriteColors(buf []color.RGBA) error {
	// Each color needs 15 bytes: 5 SPI bits per WS2812 bit with 3*8 WS2812 bits
	// per color means 120 SPI bits. In addition to that, an extra 0 byte seems
	// to be necessary on nRF5x chips to avoid having the SDO line pulled high
	// at the end of the transfer.
	if len(d.buf) < len(buf)*15+1 {
		d.buf = make([]byte, len(buf)*15+1)
	}

	for i, color := range buf {
		bitBuf := makeSPIBits(color.G)
		copy(d.buf[i*15+0:], bitBuf[:])
		bitBuf = makeSPIBits(color.R)
		copy(d.buf[i*15+5:], bitBuf[:])
		bitBuf = makeSPIBits(color.B)
		copy(d.buf[i*15+10:], bitBuf[:])
	}
	return d.Bus.Tx(d.buf, nil)
}

func makeSPIBits(b byte) [5]byte {
	// Create a 40 bit bitstring from this one byte.
	var bitstring uint64
	for i := 0; i < 8; i++ {
		bitstring <<= 5
		if b&0x80 != 0 {
			// 0b11100 means the output is high for 750ns (three high bits at
			// 4MHz) and low for 500ns (two low bits). This outputs a 1 bit in
			// the custom WS2812 protocol.
			bitstring |= 0b11100 // T1H (0b111) + TLD (0b00)
		} else {
			// 0b10000 means the output is high for 250ns (one high bit at 4MHz)
			// and low for 1000ns (four low bits at 4MHz). This outputs a 0 bit
			// in the custom WS2812 protocol.
			bitstring |= 0b10000 // T0H (0b100) + TLD (0b00)
		}
		b <<= 1
	}

	// Create a 5 byte array from this bitstring.
	bitstring <<= 7
	var buf [5]byte
	for i := 0; i < 5; i++ {
		buf[i] = byte(bitstring >> 40)
		bitstring <<= 8
	}
	return buf
}
