//go:build baremetal

package legacy

import "machine"

func configurePinOut(p PinOutput) {
	configurePin(p, machine.PinOutput)
}

func configurePinInputPulldown(p PinInput) {
	configurePin(p, pulldown) // some chips do not have pull down, in which case pulldown==machine.PinInput.
}

func configurePinInput(p PinInput) {
	configurePin(p, machine.PinInput)
}

func configurePinInputPullup(p PinInput) {
	configurePin(p, pullup) // some chips do not have pull up, in which case pullup==machine.PinInput.
}

func pinIsNoPin(a any) bool {
	p, ok := a.(machine.Pin)
	return ok && p == machine.NoPin
}

func configurePin(p any, mode machine.PinMode) {
	machinePin, ok := p.(machine.Pin)
	if ok {
		machinePin.Configure(machine.PinConfig{Mode: mode})
	}
}
