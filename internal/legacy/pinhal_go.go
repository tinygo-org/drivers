//go:build !baremetal

package legacy

func configurePinOut(p PinOutput)          {}
func configurePinInput(p PinInput)         {}
func configurePinInputPulldown(p PinInput) {}
func configurePinInputPullup(p PinInput)   {}
func pinIsNoPin(a any) bool                { return false }
