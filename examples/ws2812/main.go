// Connects to an WS2812 RGB LED strip with 10 LEDS.
//
// See either the others.go or digispark.go files in this directory
// for the neopixels pin assignments.
package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

var leds [10]color.RGBA

func main() {
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	neo.Configure(machine.PinConfig{Mode: machine.PinOutput})

	ws := ws2812.NewWS2812(neo)
	rg := false

	for {
		rg = !rg
		for i := range leds {
			rg = !rg
			if rg {
				// Alpha channel is not supported by WS2812 so we leave it out
				leds[i] = color.RGBA{R: 0xff, G: 0x00, B: 0x00}
			} else {
				leds[i] = color.RGBA{R: 0x00, G: 0xff, B: 0x00}
			}
		}

		ws.WriteColors(leds[:])
		led.Set(rg)
		time.Sleep(100 * time.Millisecond)
	}
}
