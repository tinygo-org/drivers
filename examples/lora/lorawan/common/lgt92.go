//go:build lgt92

package common

import "machine"

var (
	rstPin  = machine.PB0
	csPin   = machine.PA15
	dio0Pin = machine.PC13
	dio1Pin = machine.PB10
	spi     = machine.SPI0
)
