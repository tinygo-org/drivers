//go:build baremetal && tinygo

package legacy

import (
	"machine"

	"tinygo.org/x/drivers"
)

func PinOutput(pin drivers.PinOutput) drivers.PinOutput {
	if p, ok := pin.(machine.Pin); ok {
		p.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
	return pin
}
