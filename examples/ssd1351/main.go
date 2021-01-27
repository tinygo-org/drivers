package main

import (
	"machine"

	"image/color"

	"tinygo.org/x/drivers/ssd1351"
)

func main() {
	machine.SPI1.Configure(machine.SPIConfig{
		Frequency: 2000000,
	})
	display := ssd1351.New(machine.SPI1, machine.D18, machine.D17, machine.D16, machine.D4, machine.D19)

	display.Configure(ssd1351.Config{
		Width:        96,
		Height:       96,
		ColumnOffset: 16,
	})

	width, height := display.Size()

	white := color.RGBA{255, 255, 255, 255}
	red := color.RGBA{255, 0, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}
	green := color.RGBA{0, 255, 0, 255}

	display.FillRectangle(0, 0, width, height/4, white)
	display.FillRectangle(0, height/4, width, height/4, red)
	display.FillRectangle(0, height/2, width, height/4, green)
	display.FillRectangle(0, 3*height/4, width, height/4, blue)

	display.Display()
}
