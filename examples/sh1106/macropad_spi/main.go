//go:build macropad_rp2040

package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/sh1106"
)

var (
	display = sh1106.NewSPI(machine.SPI1, machine.OLED_DC, machine.OLED_RST, machine.OLED_CS)
)

func init() {
	machine.SPI1.Configure(machine.SPIConfig{
		Frequency: 48000000,
	})
	display.Configure(sh1106.Config{
		Width:  128,
		Height: 64,
	})
}

func main() {

	display.ClearDisplay()

	x := int16(0)
	y := int16(0)
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
