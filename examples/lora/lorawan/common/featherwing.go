//go:build featherwing

package common

import "machine"

var (
	// We assume LoRa Featherwing module with sx127x is connected to PyBadge
	rstPin  = machine.D11
	csPin   = machine.D10
	dio0Pin = machine.D6
	dio1Pin = machine.D9
	spi     = machine.SPI0
)
