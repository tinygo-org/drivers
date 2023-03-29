//go:build p1am_100

package main

import (
	"machine"
)

func init() {
	spi = &machine.SDCARD_SPI
	sckPin = machine.SDCARD_SCK_PIN
	sdoPin = machine.SDCARD_SDO_PIN
	sdiPin = machine.SDCARD_SDI_PIN
	csPin = machine.SDCARD_SS_PIN

	ledPin = machine.LED
}
