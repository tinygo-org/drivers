package main

import (
	"machine"

	"image/color"
	"math/rand"

	"tinygo.org/x/drivers/unoqmatrix"
)

var on = color.RGBA{255, 255, 255, 255}

func main() {
	display := unoqmatrix.NewFromBasePin(machine.PF0)
	display.ClearDisplay()

	w, h := display.Size()
	x := int16(0)
	y := int16(0)
	deltaX := int16(1)
	deltaY := int16(1)

	for {
		pixel := display.GetPixel(x, y)
		if pixel.R != 0 || pixel.G != 0 || pixel.B != 0 {
			display.ClearDisplay()
			x = 1 + int16(rand.Int31n(3))
			y = 1 + int16(rand.Int31n(3))
			deltaX = 1
			deltaY = 1
			if rand.Int31n(2) == 0 {
				deltaX = -1
			}
			if rand.Int31n(2) == 0 {
				deltaY = -1
			}
		}
		display.SetPixel(x, y, on)

		x += deltaX
		y += deltaY

		if x == 0 || x == w-1 {
			deltaX = -deltaX
		}

		if y == 0 || y == h-1 {
			deltaY = -deltaY
		}

		display.Display()
	}
}
