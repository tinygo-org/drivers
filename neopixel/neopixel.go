// Package neopixel implements a driver for WS2812B/NeoPixel RGB LED strips.
//
// On RP2040/RP2350, it uses PIO for hardware-timed control.
// On other platforms, it falls back to the ws2812 bit-bang driver.
package neopixel // import "tinygo.org/x/drivers/neopixel"

import (
	"image/color"
	"machine"
)

// Device controls a strip of WS2812B/NeoPixel LEDs.
type Device struct {
	pin        machine.Pin
	pixels     []color.RGBA
	brightness uint8
	backend    backend
}

// backend is the platform-specific implementation for sending pixel data.
type backend interface {
	init(pin machine.Pin) error
	writePixels(pixels []color.RGBA, brightness uint8) error
}

// New creates a new NeoPixel device on the given pin with the specified number of pixels.
func New(pin machine.Pin, numPixels int) (*Device, error) {
	d := &Device{
		pin:        pin,
		pixels:     make([]color.RGBA, numPixels),
		brightness: 255,
		backend:    newBackend(),
	}
	if err := d.backend.init(pin); err != nil {
		return nil, err
	}
	return d, nil
}

// SetPixel sets the color of pixel at index i.
func (d *Device) SetPixel(i int, c color.RGBA) {
	if i >= 0 && i < len(d.pixels) {
		d.pixels[i] = c
	}
}

// Show sends the current pixel buffer to the LED strip.
func (d *Device) Show() error {
	return d.backend.writePixels(d.pixels, d.brightness)
}

// SetBrightness sets the global brightness (0-255).
func (d *Device) SetBrightness(b uint8) {
	d.brightness = b
}

// Fill sets all pixels to the given color.
func (d *Device) Fill(c color.RGBA) {
	for i := range d.pixels {
		d.pixels[i] = c
	}
}

// Clear turns off all pixels.
func (d *Device) Clear() {
	d.Fill(color.RGBA{})
}

// NumPixels returns the number of pixels in the strip.
func (d *Device) NumPixels() int {
	return len(d.pixels)
}

// applyBrightness scales a color by the brightness value.
func applyBrightness(c color.RGBA, brightness uint8) (r, g, b uint8) {
	r = uint8((uint16(c.R) * uint16(brightness)) >> 8)
	g = uint8((uint16(c.G) * uint16(brightness)) >> 8)
	b = uint8((uint16(c.B) * uint16(brightness)) >> 8)
	return
}
