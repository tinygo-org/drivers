//go:build qtpy

package main

import "machine"

func init() {
	pwr := machine.NEOPIXELS_POWER
	pwr.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pwr.High()
}
