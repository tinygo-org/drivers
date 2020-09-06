// +build atsamd21

package main

import (
	"machine"

	"tinygo.org/x/drivers/ili9341"
)

var (
	display = ili9341.NewSPI(
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
		SDO:       machine.SPI0_SDO_PIN,
		SDI:       machine.SPI0_SDI_PIN,
		Frequency: 24000000,
	})
}
