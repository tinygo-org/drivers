//go:build !stm32wlx

package main

import (
	"machine"

	"tinygo.org/x/drivers/sx126x"
)

var (
	spi                        = machine.SPI0
	nssPin, busyPin, dio1Pin   = machine.GP17, machine.GP10, machine.GP11
	rxPin, txLowPin, txHighPin = machine.GP13, machine.GP12, machine.GP12
)

func newRadioControl() sx126x.RadioController {
	return sx126x.NewRadioControl(nssPin, busyPin, dio1Pin, rxPin, txLowPin, txHighPin)
}
