package main

import (
	"image/color"
	"time"

	"tinygo.org/x/drivers/ssd1306"
)

func main() {
	const height = 32
	const width = 128
	var display *ssd1306.Device
	var err error
	display, err = makeSSD1306(width, height)
	if err != nil {
		panic(err)
	}
	display.ClearDisplay()

	x := int16(36)
	y := int16(20)
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

		if x == 0 || x == width-1 {
			deltaX = -deltaX
		}

		if y == 0 || y == height-1 {
			deltaY = -deltaY
		}
		time.Sleep(1 * time.Millisecond)
	}
}
