package main

import (
	"machine"

	"image/color"

	"tinygo.org/x/drivers/ssd1331"
)

func main() {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 8000000,
	})
	display := ssd1331.New(machine.SPI0, machine.P6, machine.P7, machine.P8)
	display.Configure(ssd1331.Config{})
	display.SetContrast(0x30, 0x20, 0x30)
	width, height := display.Size()

	white := color.RGBA{255, 255, 255, 255}
	red := color.RGBA{255, 0, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}
	green := color.RGBA{0, 255, 0, 255}
	black := color.RGBA{0, 0, 0, 255}

	display.FillScreen(black)

	display.FillRectangle(0, 0, width/2, height/2, white)
	display.FillRectangle(width/2, 0, width/2, height/2, red)
	display.FillRectangle(0, height/2, width/2, height/2, green)
	display.FillRectangle(width/2, height/2, width/2, height/2, blue)
	display.FillRectangle(width/4, height/4, width/2, height/2, black)
}
