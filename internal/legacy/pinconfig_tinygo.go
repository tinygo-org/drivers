//go:build baremetal

package legacy

import (
	"machine"

	"tinygo.org/x/drivers/internal/pin"
)

func configurePinOut(po pin.Output) {
	configurePin(po, machine.PinOutput)
}

func configurePinInputPulldown(pi pin.Input) {
	configurePin(pi, pulldown) // some chips do not have pull down, in which case pulldown==machine.PinInput.
}

func configurePinInput(pi pin.Input) {
	configurePin(pi, machine.PinInput)
}

func configurePinInputPullup(pi pin.Input) {
	configurePin(pi, pullup) // some chips do not have pull up, in which case pullup==machine.PinInput.
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
