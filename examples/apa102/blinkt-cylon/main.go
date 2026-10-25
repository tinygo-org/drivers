// Draws a red Cylon sweep on a Pimoroni Blinkt! connected to the 40 pin
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
	pos, dir := 0, 1
	for {
		for i := range c {
			c[i].R /= 4
			c[i].A = 0x20
		}
		c[pos].R = 0xff
		leds.WriteColors(c[:])

		if pos+dir < 0 || pos+dir >= numLEDs {
			dir = -dir
		}
		pos += dir
		time.Sleep(80 * time.Millisecond)
	}
}
