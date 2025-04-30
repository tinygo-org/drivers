//go:build baremetal

package legacy

import "machine"

func configurePinOut(p PinOutput) {
	machinePin, ok := p.(machine.Pin)
	if ok {
		machinePin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
}
