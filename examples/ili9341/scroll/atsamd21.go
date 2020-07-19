// +build atsamd21

package main

import (
	"machine"

	"tinygo.org/x/drivers/ili9341"
)

var (
	display = ili9341.NewSpi(
		machine.SPI0,
		machine.D0,
		machine.D1,
		machine.D2,
	)

	backlight = machine.D3
)

func init() {
	machine.SPI0.Configure(machine.SPIConfig{
		SCK:       machine.SPI0_SCK_PIN,
		MOSI:      machine.SPI0_MOSI_PIN,
		MISO:      machine.SPI0_MISO_PIN,
		Frequency: 24000000,
	})
}
