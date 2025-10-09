package legacy

import (
	"errors"

	"tinygo.org/x/drivers"
)

// The pingconfig group of files serve to abstract away
// pin configuration calls on the machine.Pin type.
// It was observed this way of developing drivers was
// non-portable and unusable on "big" Go projects so
// future projects should NOT configure pins in driver code.
// Users must configure pins before passing them as arguments
// to drivers.

// ConfigurePinOut is a legacy function used to configure pins as outputs.
//
// Deprecated: Do not configure pins in drivers.
// This is a legacy feature and should only be used by drivers that
// previously configured pins in initialization to avoid breaking users.
func ConfigurePinOut(po drivers.OutputPin) {
	configurePinOut(po)
}

// ConfigurePinInput is a legacy function used to configure pins as inputs.
//
// Deprecated: Do not configure pins in drivers.
// This is a legacy feature and should only be used by drivers that
// previously configured pins in initialization to avoid breaking users.
func ConfigurePinInputPulldown(pi drivers.InputPin) {
	configurePinInputPulldown(pi)
}

// ConfigurePinInput is a legacy function used to configure pins as inputs.
//
// Deprecated: Do not configure pins in drivers.
// This is a legacy feature and should only be used by drivers that
// previously configured pins in initialization to avoid breaking users.
func ConfigurePinInput(pi drivers.InputPin) {
	configurePinInput(pi)
}

// ConfigurePinInput is a legacy function used to configure pins as inputs.
//
// Deprecated: Do not configure pins in drivers.
// This is a legacy feature and should only be used by drivers that
// previously configured pins in initialization to avoid breaking users.
func ConfigurePinInputPullup(pi drivers.InputPin) {
	configurePinInputPullup(pi)
}

// PinIsNoPin returns true if the argument is a machine.Pin type and is the machine.NoPin predeclared type.
//
// Deprecated: Drivers do not require pin knowledge from now on.
func PinIsNoPin(pin drivers.Pin) bool {
	return pinIsNoPin(pin)
}

var (
	ErrConfigBeforeInstantiated = errors.New("device must be instantiated with New before calling Configure method")
)
