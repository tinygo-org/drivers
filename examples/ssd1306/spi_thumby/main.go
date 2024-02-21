// This example using the SSD1306 OLED display over SPI on the Thumby board
// A very tiny 72x40 display.
package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ssd1306"
)

func main() {
	machine.SPI0.Configure(machine.SPIConfig{})
	display := ssd1306.NewSPI(machine.SPI0, machine.THUMBY_DC_PIN, machine.THUMBY_RESET_PIN, machine.THUMBY_CS_PIN)
	display.Configure(ssd1306.Config{
		Width:     72,
		Height:    40,
		ResetCol:  ssd1306.ResetValue{28, 99},
		ResetPage: ssd1306.ResetValue{0, 5},
	})

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

		if x == 0 || x == 71 {
			deltaX = -deltaX
		}

		if y == 0 || y == 39 {
			deltaY = -deltaY
		}
		time.Sleep(1 * time.Millisecond)
	}
}
