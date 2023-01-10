//go:build thingplus_rp2040

package main

import (
	"machine"
)

func init() {
	spi = machine.SPI1
	sckPin = machine.SPI1_SCK_PIN
	sdoPin = machine.SPI1_SDO_PIN
	sdiPin = machine.SPI1_SDI_PIN
	csPin = machine.GPIO9

	ledPin = machine.LED
}
