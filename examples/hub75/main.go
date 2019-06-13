package main

import (
	"machine"

	"image/color"
	"time"

	"tinygo.org/x/drivers/examples/hub75/gopherimg"
	"tinygo.org/x/drivers/hub75"
)

var display hub75.Device

func main() {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 8000000,
		Mode:      0},
	)

	display = hub75.New(machine.SPI0, 11, 12, 6, 10, 18, 20)
	display.Configure(hub75.Config{
		Width:      64,
		Height:     32,
		RowPattern: 16,
		ColorDepth: 6,
	})

	colors := []color.RGBA{
		{255, 0, 0, 255},
		{255, 255, 0, 255},
		{0, 255, 0, 255},
		{0, 255, 255, 255},
		{0, 0, 255, 255},
		{255, 0, 255, 255},
		{255, 255, 255, 255},
	}

	display.ClearDisplay()
	display.SetBrightness(100)

	step := 0
	then := time.Now()
	size := int16(8)
	x := int16(0)
	y := int16(0)
	dx := int16(1)
	dy := int16(1)
	c := 0
	for {
		if time.Since(then).Nanoseconds() > 800000000 {
			then = time.Now()
			step++

			if step < 23 {
				showRect(size, x*size, y*size, colors[c])
				c = (c + 1) % 7
				x += dx
				y += dy
				if x >= (64 / size) {
					dx = -1
					x += 2 * dx
				}
				if y >= (32 / size) {
					dy = -1
					y += 2 * dy
				}
				if x < 0 {
					dx = 1
					x += 2 * dx
				}
				if y < 0 {
					dy = 1
					y += 2 * dy
				}
			} else if step == 23 {
				showGopher()
			} else if step == 30 {
				display.ClearDisplay()
				step = 0
				x = 0
				y = 0
			}
		}
		display.Display()
	}
}

func showGopher() {
	for i := int16(0); i < 64; i++ {
		for j := int16(0); j < 32; j++ {
			display.SetPixel(i, j, gopherimg.Int2Color(gopherimg.ImageArray[32*i+j]))
		}
	}
}

func showRect(size int16, x int16, y int16, c color.RGBA) {
	for i := x; i < x+size; i++ {
		for j := y; j < y+size; j++ {
			display.SetPixel(i, j, c)
		}
	}
}
