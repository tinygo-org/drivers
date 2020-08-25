package main

import (
	"machine"

	"image/color"

	"time"

	"tinygo.org/x/drivers/waveshare-epd/epd4in2"
)

var display epd4in2.Device

func main() {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 8000000,
		Mode:      0,
	})

	display = epd4in2.New(machine.SPI0, machine.P6, machine.P7, machine.P8, machine.P9)
	display.Configure(epd4in2.Config{})

	black := color.RGBA{1, 1, 1, 255}

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

	println("You could remove power now")
}

func showRect(x int16, y int16, w int16, h int16, c color.RGBA) {
	for i := x; i < x+w; i++ {
		for j := y; j < y+h; j++ {
			display.SetPixel(i, j, c)
		}
	}
}
