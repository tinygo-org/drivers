// +build feather_stm32f405

package main

import (
	"machine"

	"tinygo.org/x/drivers/ili9341"
)

var (
	csPin   = machine.D12
	dcPin   = machine.D11
	display = ili9341.NewSpi(
		machine.SPI0,
		dcPin,
		csPin,
		machine.D8,
	)

	backlight = machine.D9
)

func init() {
	machine.SPI0.Configure(machine.SPIConfig{
		SCK:       machine.SPI0_SCK_PIN,
		SDO:       machine.SPI0_SDO_PIN,
		SDI:       machine.SPI0_SDI_PIN,
		Frequency: 40000000,
	})
}
