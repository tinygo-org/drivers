// +build wioterminal

package main

import (
	"machine"
)

func init() {
	spi = machine.SPI2
	spi.Configure(machine.SPIConfig{
		SCK:       machine.SCK2,
		SDO:       machine.SDO2,
		SDI:       machine.SDI2,
		Frequency: 24000000,
		LSBFirst:  false,
		Mode:      0, // phase=0, polarity=0
	})

	csPin = machine.SS2

	ledPin = machine.LED
}
