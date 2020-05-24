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
		machine.LCD_SS_PIN,
		machine.LCD_RESET,
	)

	backlight = machine.LCD_BACKLIGHT
)

func init() {
	machine.OUTPUT_CTR_5V.Configure(machine.PinConfig{machine.PinOutput})
	machine.OUTPUT_CTR_3V3.Configure(machine.PinConfig{machine.PinOutput})

	machine.OUTPUT_CTR_5V.High()
	machine.OUTPUT_CTR_3V3.Low()

	machine.SPI3.Configure(machine.SPIConfig{
		SCK:       machine.LCD_SCK_PIN,
		MOSI:      machine.LCD_MOSI_PIN,
		MISO:      machine.LCD_MISO_PIN,
		Frequency: 40000000,
	})
}
