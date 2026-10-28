//go:build stm32wlx

package main

import (
	"machine"

	"tinygo.org/x/drivers/sx126x"
)

var (
	spi    = machine.SPI3
	rstPin = machine.NoPin
)

func newRadioControl() sx126x.RadioController {
	return sx126x.NewRadioControl()
}
