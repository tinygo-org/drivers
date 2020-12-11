package main

import (
	"image/color"
	"machine"
	"time"

	// reuse the image data from example for driver package "hub75"
	"tinygo.org/x/drivers/examples/hub75/gopherimg"

	"tinygo.org/x/drivers/rgb75"
)

func main() {

	// actual rgb75 Device object
	display := rgb75.New(
		machine.HUB75_OE, machine.HUB75_LAT, machine.HUB75_CLK,
		[6]machine.Pin{
			machine.HUB75_R1, machine.HUB75_G1, machine.HUB75_B1,
			machine.HUB75_R2, machine.HUB75_G2, machine.HUB75_B2,
		},
		[]machine.Pin{
			machine.HUB75_ADDR_A, machine.HUB75_ADDR_B, machine.HUB75_ADDR_C,
			machine.HUB75_ADDR_D, machine.HUB75_ADDR_E,
		})

	// panel layout and color depth
	config := rgb75.Config{
		Width:      64,
		Height:     32,
		ColorDepth: 4,
	}

	if err := display.Configure(config); nil != err {
		for {
			println("error: " + err.Error())
			time.Sleep(time.Second)
		}
	}

	// begin screen updates
	display.Resume()

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
				showRect(display, size, x*size, y*size, colors[c])
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
				showGopher(display)
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

func showGopher(display *rgb75.Device) {
	width, height := display.Size()
	for i := int16(0); i < width; i++ {
		for j := int16(0); j < height; j++ {
			display.SetPixel(i, j, gopherimg.Int2Color(gopherimg.ImageArray[height*i+j]))
		}
	}
}

func showRect(display *rgb75.Device, size int16, x int16, y int16, c color.RGBA) {
	for i := x; i < x+size; i++ {
		for j := y; j < y+size; j++ {
			display.SetPixel(i, j, c)
		}
	}
}
