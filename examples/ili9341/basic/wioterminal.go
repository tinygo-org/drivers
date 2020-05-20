// +build wioterminal

package main

import (
	"machine"

	"tinygo.org/x/drivers/ili9341"
)

var (
	display = ili9341.NewSpi(
		machine.SPI3,
		machine.LCD_DC,
		machine.LCD_CS,
		machine.LCD_RESET,
	)

	backlight = machine.LCD_BACKLIGHT_CTR
)

func init() {
	machine.OUTPUT_CTR_5V.Configure(machine.PinConfig{machine.PinOutput})
	machine.OUTPUT_CTR_3V3.Configure(machine.PinConfig{machine.PinOutput})

	machine.OUTPUT_CTR_5V.High()
	machine.OUTPUT_CTR_3V3.Low()

	machine.SPI3.Configure(machine.SPIConfig{
		SCK:       machine.LCD_SCK,
		MOSI:      machine.LCD_MOSI,
		MISO:      machine.LCD_MISO,
		Frequency: 40000000,
	})
}
