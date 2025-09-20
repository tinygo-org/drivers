//go:build !baremetal

package legacy

import (
	"tinygo.org/x/drivers"
)

func configurePinOut(p drivers.PinOutput)          {}
func configurePinInput(p drivers.PinInput)         {}
func configurePinInputPulldown(p drivers.PinInput) {}
func configurePinInputPullup(p drivers.PinInput)   {}
func pinIsNoPin(a any) bool                        { return false }
