//go:build baremetal

package pin

import "machine"

func configureOutput(p Output) {
	configure(p, machine.PinOutput)
}

func configureInputPulldown(p Input) {
	configure(p, pulldown) // some chips do not have pull down, in which case pulldown==machine.PinInput.
}

func configureInput(p Input) {
	configure(p, machine.PinInput)
}

func configureInputPullup(p Input) {
	configure(p, pullup) // some chips do not have pull up, in which case pullup==machine.PinInput.
}

func isNotPin(a any) bool {
	p, ok := a.(machine.Pin)
	return ok && p == machine.NoPin
}

func configure(p any, mode machine.PinMode) {
	machinePin, ok := p.(machine.Pin)
	if ok {
		machinePin.Configure(machine.PinConfig{Mode: mode})
	}
}
