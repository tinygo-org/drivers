//go:build baremetal && fe310

package pin

import "machine"

const (
	pulldown = machine.PinInput
	pullup   = machine.PinInput
)
