//go:build baremetal && fe310

package legacy

import "machine"

const (
	pulldown = machine.PinInput
	pullup   = machine.PinInput
)
