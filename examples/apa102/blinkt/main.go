// Draws a scrolling rainbow on a Pimoroni Blinkt! connected to the 40 pin
// header of a Waveshare RP2040-PiZero.
// https://shop.pimoroni.com/products/blinkt
package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/apa102"
)

const numLEDs = 8

func main() {
	// Blinkt! uses BCM23 for data and BCM24 for clock. These are not
	// hardware SPI pins on the RP2040, so use software SPI.
	leds := apa102.NewSoftwareSPI(machine.GP24, machine.GP23, 1)

	var c [numLEDs]color.RGBA
	for step := uint8(0); ; step++ {
		for i := range c {
			c[i] = wheel(uint8(i*256/numLEDs) + step*2)
		}
		leds.WriteColors(c[:])
		time.Sleep(20 * time.Millisecond)
	}
}

func wheel(pos uint8) color.RGBA {
	const brightness = 0x20
	switch {
	case pos < 85:
		return color.RGBA{R: 255 - pos*3, G: pos * 3, A: brightness}
	case pos < 170:
		pos -= 85
		return color.RGBA{G: 255 - pos*3, B: pos * 3, A: brightness}
	default:
		pos -= 170
		return color.RGBA{R: pos * 3, B: 255 - pos*3, A: brightness}
	}
}
