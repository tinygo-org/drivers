package main

import (
	"machine"

	"image/color"

	"time"

	"tinygo.org/x/drivers/waveshare-epd/epd2in13"
)

var display epd2in13.Device

func main() {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 8000000,
		Mode:      0,
	})

	display = epd2in13.New(machine.SPI0, machine.P6, machine.P7, machine.P8, machine.P9)
	display.Configure(epd2in13.Config{})

	black := color.RGBA{1, 1, 1, 255}
	white := color.RGBA{0, 0, 0, 255}

	display.ClearBuffer()
	println("Clear the display")
	display.ClearDisplay()
	display.WaitUntilIdle()
	println("Waiting for 2 seconds")
	time.Sleep(2 * time.Second)

	// Show a checkered board
	for i := int16(0); i < 16; i++ {
		for j := int16(0); j < 25; j++ {
			if (i+j)%2 == 0 {
				showRect(i*8, j*10, 8, 10, black)
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

	// There are two memory areas in the display, once the display is refreshed, memory areas are auto-toggled.
	// DisplayRect needs to be called twice
	display.DisplayRect(40, 83, 48, 83)
	display.WaitUntilIdle()
	display.DisplayRect(40, 83, 48, 83)
	display.WaitUntilIdle()

	println("You could remove power now")
}

func showRect(x int16, y int16, w int16, h int16, c color.RGBA) {
	for i := x; i < x+w; i++ {
		for j := y; j < y+h; j++ {
			display.SetPixel(i, j, c)
		}
	}
}
