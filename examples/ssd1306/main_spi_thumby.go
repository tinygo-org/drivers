//go:build thumby

// This initializes SSD1306 OLED display driver over SPI.
//
// Thumby board has a tiny built-in 72x40 display.
//
// As the display is built-in, no wiring is needed.

package main

import (
	"machine"

	"tinygo.org/x/drivers/ssd1306"
)

func newSSD1306Display() *ssd1306.Device {
	machine.SPI0.Configure(machine.SPIConfig{})
	display := ssd1306.NewSPI(machine.SPI0, machine.THUMBY_DC_PIN, machine.THUMBY_RESET_PIN, machine.THUMBY_CS_PIN)
	display.Configure(ssd1306.Config{
		Width:     72,
		Height:    40,
		ResetCol:  ssd1306.ResetValue{28, 99},
		ResetPage: ssd1306.ResetValue{0, 5},
	})
	return display
}
