// Demonstrates NeoPixel (WS2812B) RGB LED control.
//
// Tested on Waveshare RP2350-LCD-1.47 with onboard RGB LED on GP22.
//
// To flash:
//
//	tinygo flash -target=pico2 ./examples/neopixel/
package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/neopixel"
)

func main() {
	pixel, err := neopixel.New(machine.GP22, 1)
	if err != nil {
		panic(err)
	}
	pixel.SetBrightness(50)

	colors := []color.RGBA{
		{R: 255, G: 0, B: 0},
		{R: 0, G: 255, B: 0},
		{R: 0, G: 0, B: 255},
	}

	for i := 0; ; i = (i + 1) % 3 {
		pixel.SetPixel(0, colors[i])
		pixel.Show()
		time.Sleep(time.Second)
	}
}
