// +build teensy40

package main

import (
	"machine"

	"tinygo.org/x/drivers/ili9341"
)

var (
	display = ili9341.NewSPI(
		machine.SPI1,
		machine.D9,          // DC
		machine.SPI1_CS_PIN, // CS
		machine.D4,          // RST
	)

	backlight = machine.D2
)

func init() {
	machine.SPI1.Configure(machine.SPIConfig{
		Frequency: 34000000,
		SDI:       machine.SPI1_SDI_PIN,
		SDO:       machine.SPI1_SDO_PIN,
		SCK:       machine.SPI1_SCK_PIN,
		CS:        machine.SPI1_CS_PIN,
	})
}
