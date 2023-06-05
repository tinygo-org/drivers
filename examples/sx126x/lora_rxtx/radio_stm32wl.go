//go:build stm32wlx

package main

import (
	"machine"

	"tinygo.org/x/drivers/sx126x"
)

var spi = machine.SPI3

func newRadioControl() sx126x.RadioController {
	return sx126x.NewRadioControl()
}
