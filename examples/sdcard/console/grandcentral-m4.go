//go:build grandcentral_m4

package main

import (
	"machine"
)

func init() {
	spi = &machine.SPI1
	sckPin = machine.SDCARD_SCK_PIN
	sdoPin = machine.SDCARD_SDO_PIN
	sdiPin = machine.SDCARD_SDI_PIN
	csPin = machine.SDCARD_CS_PIN

	ledPin = machine.LED
}
