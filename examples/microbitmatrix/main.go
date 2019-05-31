package main

import (
	"image/color"
	"math/rand"
	"time"

	"tinygo.org/x/drivers/microbitmatrix"
)

var display microbitmatrix.Device

func main() {
	display = microbitmatrix.New()
	display.Configure(microbitmatrix.Config{})

	display.ClearDisplay()

	x := int16(1)
	y := int16(2)
	deltaX := int16(1)
	deltaY := int16(1)
	then := time.Now()
	c := color.RGBA{255, 255, 255, 255}

	for {
		if time.Since(then).Nanoseconds() > 80000000 {
			then = time.Now()

			pixel := display.GetPixel(x, y)
			if pixel {
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
			display.SetPixel(x, y, c)

			x += deltaX
			y += deltaY

			if x == 0 || x == 4 {
				deltaX = -deltaX
			}

			if y == 0 || y == 4 {
				deltaY = -deltaY
			}
		}
		display.Display()
	}
}
