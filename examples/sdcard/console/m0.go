//go:build atsamd21 && !p1am_100

package main

import (
	"machine"
)

func init() {
	spi = &machine.SPI0
	sckPin = machine.SPI0_SCK_PIN
	sdoPin = machine.SPI0_SDO_PIN
	sdiPin = machine.SPI0_SDI_PIN
	csPin = machine.D2

	ledPin = machine.LED
}
