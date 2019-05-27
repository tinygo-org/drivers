// Connects to an APA102 SPI RGB LED strip with 30 LEDS.
package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/apa102"
)

func main() {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 500000,
		Mode:      0})

	a := apa102.New(machine.SPI0)
	leds := make([]color.RGBA, 30)
	rg := false

	for {
		rg = !rg
		for i := range leds {
			rg = !rg
			if rg {
				leds[i] = color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0x77}
			} else {
				leds[i] = color.RGBA{R: 0x00, G: 0xff, B: 0x00, A: 0x77}
			}
		}

		a.WriteColors(leds)
		time.Sleep(100 * time.Millisecond)
	}
}
