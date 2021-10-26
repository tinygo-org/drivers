// +build feather_stm32f405

package main

import (
	"machine"

	"tinygo.org/x/drivers/ili9341"
)

var (
	csPin   = machine.D12
	dcPin   = machine.D10
	display = ili9341.NewSPI(
		machine.SPI0,
		dcPin,
		csPin,
		machine.D8, // if wired to 3.3V, pick an unused pin
	)

	// ILI9341's LED pin. set this to an unused pin (but not NoPin!) if
	// wired via resistor straight to 3.3V. the boing example tries to
	// set this pin and will panic if NoPin is used.
	backlight = machine.D13
)

func init() {
	machine.SPI0.Configure(machine.SPIConfig{
		SCK:       machine.SPI0_SCK_PIN,
		SDO:       machine.SPI0_SDO_PIN,
		SDI:       machine.SPI0_SDI_PIN,
		Frequency: 40000000,
	})
}
