//go:build !baremetal

package legacy

func configurePinOut(p PinOutput)          {}
func configurePinInput(p PinInput)         {}
func configurePinInputPulldown(p PinInput) {}
func configurePinInputPullup(p PinInput)   {}
