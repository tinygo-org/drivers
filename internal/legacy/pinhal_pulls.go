//go:build baremetal && !fe310

package legacy

import "machine"

const (
	pulldown = machine.PinInputPulldown
	pullup   = machine.PinInputPullup
)
