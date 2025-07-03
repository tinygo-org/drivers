//go:build baremetal

package legacy

import "machine"

func configurePinOut(p PinOutput) {
	configurePin(p, machine.PinInputPulldown)
}

func configurePinInputPulldown(p PinInput) {
	configurePin(p, machine.PinInputPulldown)
}

func configurePinInput(p PinInput) {
	configurePin(p, machine.PinInput)
}

func configurePinInputPullup(p PinInput) {
	configurePin(p, machine.PinInputPullup)
}

func configurePin(p any, mode machine.PinMode) {
	machinePin, ok := p.(machine.Pin)
	if ok {
		machinePin.Configure(machine.PinConfig{Mode: mode})
	}
}
