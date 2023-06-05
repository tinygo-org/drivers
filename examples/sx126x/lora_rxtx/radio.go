//go:build !stm32wlx

package main

import (
	"machine"

	"tinygo.org/x/drivers/sx126x"
)

var (
	spi                        = machine.SPI1
	nssPin, busyPin, dio1Pin   = machine.GP13, machine.GP6, machine.GP7
	rxPin, txLowPin, txHighPin = machine.GP9, machine.GP8, machine.GP8
)

func newRadioControl() sx126x.RadioController {
	return sx126x.NewRadioControl(nssPin, busyPin, dio1Pin, rxPin, txLowPin, txHighPin)
}

func init() {
	spi.Configure(machine.SPIConfig{
		Mode:      0,
		Frequency: 8 * 1e6,
		SDO:       machine.SPI1_SDO_PIN,
		SDI:       machine.SPI1_SDI_PIN,
		SCK:       machine.SPI1_SCK_PIN,
	})
}
