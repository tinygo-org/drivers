//go:build macropad_rp2040

package main

import (
	"machine"

	"tinygo.org/x/drivers/encoders"
)

var (
	enc = encoders.NewQuadratureViaInterrupt(machine.ROT_A, machine.ROT_B)
)

func main() {

	enc.Configure(encoders.QuadratureConfig{
		Precision: 4,
	})

	for oldValue := 0; ; {
		if newValue := enc.Position(); newValue != oldValue {
			println("value: ", newValue)
			oldValue = newValue
		}
	}

}
