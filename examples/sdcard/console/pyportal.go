// +build pyportal

package main

import (
	"machine"
)

func init() {
	spi = machine.SPI0
	spi.Configure(machine.SPIConfig{
		SCK:       machine.SPI0_SCK_PIN,
		SDO:       machine.SPI0_SDO_PIN,
		SDI:       machine.SPI0_SDI_PIN,
		Frequency: 24000000,
		LSBFirst:  false,
		Mode:      0, // phase=0, polarity=0
	})

	csPin = machine.D32 // SD_CS

	ledPin = machine.LED
}
