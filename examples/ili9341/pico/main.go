//go:build rp2040 || rp2350

package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ili9341"
)

func main() {
	if err := machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 8_000_000,
		Mode:      0,
		SCK:       machine.GPIO18,
		SDO:       machine.GPIO19,
	}); err != nil {
		panic(err)
	}

	display := ili9341.NewSPI(machine.SPI0, machine.GPIO20, machine.GPIO17, machine.GPIO21)
	display.Configure(ili9341.Config{})

	colors := [...]color.RGBA{
		{R: 255, A: 255},
		{G: 255, A: 255},
		{B: 255, A: 255},
		{R: 255, G: 255, B: 255, A: 255},
	}
	display.FillScreen(color.RGBA{A: 255})
	for i, c := range colors {
		if err := display.FillRectangle(0, int16(i*80), 240, 80, c); err != nil {
			panic(err)
		}
	}

	for {
		time.Sleep(time.Hour)
	}
}
