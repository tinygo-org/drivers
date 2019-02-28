// Connects to an WS2812 RGB LED strip with 10 LEDS, such as
// on an Adafruit Circuit Playground Express board.
//
// Replace machine.NEOPIXELS in the code below to match the pin
// that you are using, if you have a different board.
package main

import (
	"image/color"
	"machine"
	"time"

	"github.com/tinygo-org/drivers/ws2812"
)

func main() {
	neo := machine.GPIO{machine.NEOPIXELS}
	neo.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})

	ws := ws2812.New(neo)
	leds := make([]color.RGBA, 10)
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

		ws.WriteColors(leds)
		time.Sleep(100 * time.Millisecond)
	}
}
