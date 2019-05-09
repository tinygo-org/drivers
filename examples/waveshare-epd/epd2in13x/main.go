package main

import (
	"machine"

	"image/color"

	"github.com/tinygo-org/drivers/waveshare-epd/epd2in13x"
)

var display epd2in13x.Device

func main() {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 8000000,
		Mode:      0,
	})

	display = epd2in13x.New(machine.SPI0, machine.P6, machine.P7, machine.P8, machine.P9)
	display.Configure(epd2in13x.Config{})

	colors := []color.RGBA{
		{255, 0, 0, 255},
		{255, 255, 0, 255},
		{0, 255, 0, 255},
		{0, 255, 255, 255},
		{0, 0, 255, 255},
		{255, 0, 255, 255},
		{255, 255, 255, 255},
		{0, 0, 0, 255},
	}

	display.ClearBuffer()
	display.ClearDisplay()

	// Show a checkered board
	for i := int16(0); i < 27; i++ {
		showRect((i%3)*35, i*8, 35, 8, colors[0])     // COLORED
		showRect(((i+1)%3)*35, i*8, 35, 8, colors[1]) // BLACK
		showRect(((i+2)%3)*35, i*8, 35, 8, colors[7]) // WHITE
	}
	display.Display()
	display.WaitUntilIdle()
	println("You could remove power now")
}

func showRect(x int16, y int16, w int16, h int16, c color.RGBA) {
	for i := x; i < x+w; i++ {
		for j := y; j < y+h; j++ {
			display.SetPixel(i, j, c)
		}
	}
}
