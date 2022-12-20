//go:build feather_m4 || feather_m4_can || feather_nrf52840

package main

import (
	"machine"
)

func init() {
	spi = &machine.SPI0
	sckPin = machine.SPI0_SCK_PIN
	sdoPin = machine.SPI0_SDO_PIN
	sdiPin = machine.SPI0_SDI_PIN
	csPin = machine.D10

	ledPin = machine.LED
}
