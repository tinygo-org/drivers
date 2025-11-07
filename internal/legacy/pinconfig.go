package legacy

import (
	"errors"

	"tinygo.org/x/drivers/internal/pin"
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
func ConfigurePinOut(po pin.Output) {
	configurePinOut(po)
}

// ConfigurePinInput is a legacy function used to configure pins as inputs.
//
// Deprecated: Do not configure pins in drivers.
// This is a legacy feature and should only be used by drivers that
// previously configured pins in initialization to avoid breaking users.
func ConfigurePinInputPulldown(pi pin.Input) {
	configurePinInputPulldown(pi)
}

// ConfigurePinInput is a legacy function used to configure pins as inputs.
//
// Deprecated: Do not configure pins in drivers.
// This is a legacy feature and should only be used by drivers that
// previously configured pins in initialization to avoid breaking users.
func ConfigurePinInput(pi pin.Input) {
	configurePinInput(pi)
}

// ConfigurePinInput is a legacy function used to configure pins as inputs.
//
// Deprecated: Do not configure pins in drivers.
// This is a legacy feature and should only be used by drivers that
// previously configured pins in initialization to avoid breaking users.
func ConfigurePinInputPullup(pi pin.Input) {
	configurePinInputPullup(pi)
}

// PinIsNoPin returns true if the argument is a machine.Pin type and is the machine.NoPin predeclared type.
//
// Deprecated: Drivers do not require pin knowledge from now on.
func PinIsNoPin(pin any) bool {
	return pinIsNoPin(pin)
}

var (
	ErrConfigBeforeInstantiated = errors.New("device must be instantiated with New before calling Configure method")
)
