package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ssd1306"
)

func main() {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 8000000,
	})
	display := ssd1306.NewSPI(machine.SPI0, machine.P8, machine.P7, machine.P9)
	display.Configure(ssd1306.Config{
		Width:  128,
		Height: 64,
	})

	display.ClearDisplay()

	x := int16(64)
	y := int16(32)
	deltaX := int16(1)
	deltaY := int16(1)
	for {
		pixel := display.GetPixel(x, y)
		c := color.RGBA{255, 255, 255, 255}
		if pixel {
			c = color.RGBA{0, 0, 0, 255}
		}
		display.SetPixel(x, y, c)
		display.Display()

		x += deltaX
		y += deltaY

		if x == 0 || x == 127 {
			deltaX = -deltaX
		}

		if y == 0 || y == 63 {
			deltaY = -deltaY
		}
		time.Sleep(1 * time.Millisecond)
	}
}
