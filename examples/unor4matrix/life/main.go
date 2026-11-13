//go:build arduino_uno_r4_wifi

package main

import (
	"image/color"
	"time"

	"tinygo.org/x/drivers/unor4matrix"
)

const generationInterval = 300 * time.Millisecond

func main() {
	display := unor4matrix.NewOnboard()
	if err := display.Configure(); err != nil {
		println(err.Error())
		return
	}
	cells := seed()
	draw := func() {
		display.ClearDisplay()
		for y, row := range cells {
			for x, alive := range row {
				if alive {
					display.SetPixel(int16(x), int16(y), color.RGBA{G: 255})
				}
			}
		}
	}
	draw()
	nextUpdate := time.Now().Add(generationInterval)
	for {
		if err := display.Display(); err != nil {
			println(err.Error())
			return
		}
		if now := time.Now(); !now.Before(nextUpdate) {
			cells = nextGeneration(cells)
			draw()
			nextUpdate = now.Add(generationInterval)
		}
	}
}
