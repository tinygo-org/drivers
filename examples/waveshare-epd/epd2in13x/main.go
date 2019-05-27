package main

import (
	"machine"

	"image/color"

	"tinygo.org/x/drivers/waveshare-epd/epd2in13x"
)

var display epd2in13x.Device

func main() {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 8000000,
		Mode:      0,
	})

	display = epd2in13x.New(machine.SPI0, machine.P6, machine.P7, machine.P8, machine.P9)
	display.Configure(epd2in13x.Config{})

	white := color.RGBA{0, 0, 0, 255}
	colored := color.RGBA{255, 0, 0, 255}
	black := color.RGBA{1, 1, 1, 255}

	display.ClearBuffer()
	display.ClearDisplay()

	// Show a checkered board
	for i := int16(0); i < 27; i++ {
		showRect((i%3)*35, i*8, 35, 8, colored)
		showRect(((i+1)%3)*35, i*8, 35, 8, black)
		showRect(((i+2)%3)*35, i*8, 35, 8, white)
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
