package main

import (
	"machine"
	"time"

	"image/color"

	"tinygo.org/x/drivers/gc9a01"
)

func main() {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 80000000,
	})
	display := gc9a01.New(machine.SPI0, machine.P6, machine.P7, machine.P8, machine.P9)
	display.Configure(gc9a01.Config{Orientation: gc9a01.HORIZONTAL, Width: 240, Height: 240})

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

	for {
		time.Sleep(time.Hour)
	}

}
