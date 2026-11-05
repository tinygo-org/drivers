// Draws a scrolling rainbow on the Adafruit NeoPixel Shield, a grid of
// 8x5 WS2812 LEDs driven from D6.
// https://www.adafruit.com/product/1430
package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

const (
	cols = 8
	rows = 5
)

func main() {
	pin := machine.D6
	pin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Full brightness on all 40 LEDs draws about 2.4A, far more than USB can supply.
	shield := ws2812.NewWS2812(pin)
	shield.SetBrightness(16)

	var leds [cols * rows]color.RGBA
	for step := uint8(0); ; step++ {
		for y := 0; y < rows; y++ {
			for x := 0; x < cols; x++ {
				leds[y*cols+x] = wheel(uint8(x*32+y*16) + step*4)
			}
		}
		shield.WriteColors(leds[:])
		time.Sleep(20 * time.Millisecond)
	}
}

// wheel maps 0-255 to a colour on a red, green, blue colour wheel.
func wheel(pos uint8) color.RGBA {
	switch {
	case pos < 85:
		return color.RGBA{R: 255 - pos*3, G: pos * 3, A: 255}
	case pos < 170:
		pos -= 85
		return color.RGBA{G: 255 - pos*3, B: pos * 3, A: 255}
	default:
		pos -= 170
		return color.RGBA{R: pos * 3, B: 255 - pos*3, A: 255}
	}
}
