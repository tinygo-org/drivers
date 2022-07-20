package main

import (
	"machine"

	"image/color"

	"time"

	"tinygo.org/x/drivers/waveshare-epd/epd2in9"
)

var display epd2in9.Device

const (
	width  = 128
	height = 296
)

func main() {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 8000000,
		Mode:      0,
	})

	display = epd2in9.New(machine.SPI0, machine.GPIO2, machine.GPIO3, machine.GPIO4, machine.GPIO5)
	display.Configure(epd2in9.Config{
		Width:        width,
		LogicalWidth: width,
		Height:       height,
	})

	black := color.RGBA{1, 1, 1, 255}
	white := color.RGBA{0, 0, 0, 255}

	display.ClearBuffer()
	println("Clear the display")
	display.ClearDisplay()
	display.WaitUntilIdle()
	println("Waiting for 2 seconds")
	time.Sleep(2 * time.Second)

	// Show a checkered board
	for i := int16(0); i < width/8; i++ {
		for j := int16(0); j < height/8; j++ {
			if (i+j)%2 == 0 {
				showRect(i*8, j*8, 8, 8, black)
			}
		}
	}
	println("Show checkered board")
	display.Display()
	display.WaitUntilIdle()
	println("Waiting for 2 seconds")
	time.Sleep(2 * time.Second)

	println("Set partial lut")
	display.SetLUT(false) // partial updates (faster, but with some ghosting)
	println("Show smaller striped area")
	for i := int16(40); i < 88; i++ {
		for j := int16(83); j < 166; j++ {
			if (i+j)%4 == 0 || (i+j)%4 == 1 {
				display.SetPixel(i, j, black)
			} else {
				display.SetPixel(i, j, white)
			}
		}
	}

	display.Display()
	display.WaitUntilIdle()

	display.DeepSleep()

	println("You could remove power now")
}

func showRect(x int16, y int16, w int16, h int16, c color.RGBA) {
	for i := x; i < x+w; i++ {
		for j := y; j < y+h; j++ {
			display.SetPixel(i, j, c)
		}
	}
}
