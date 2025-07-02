//go:build thumby

package main

import (
	"machine"

	"tinygo.org/x/drivers/ssd1306"
)

func makeSSD1306(_, _ int16) (*ssd1306.Device, error) {
	// width and height are known for thumby.
	machine.SPI0.Configure(machine.SPIConfig{})
	display := ssd1306.NewSPI(machine.SPI0, machine.THUMBY_DC_PIN, machine.THUMBY_RESET_PIN, machine.THUMBY_CS_PIN)
	display.Configure(ssd1306.Config{
		Width:     72,
		Height:    40,
		ResetCol:  ssd1306.ResetValue{28, 99},
		ResetPage: ssd1306.ResetValue{0, 5},
	})
	return &display, nil
}
