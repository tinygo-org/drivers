//go:build wioterminal

package main

import (
	"machine"
)

func init() {
	spi = &machine.SPI2
	sckPin = machine.SCK2
	sdoPin = machine.SDO2
	sdiPin = machine.SDI2
	csPin = machine.SS2

	ledPin = machine.LED
}
