//go:build !(baremetal && tinygo)

package legacy

import "tinygo.org/x/drivers"

func PinOutput(pin drivers.PinOutput) drivers.PinOutput {
	return pin
}
