package legacy

import "errors"

// PinOutput represents a pin hardware abstraction layer for a pin that can output a digital signal.
// This is an alternative to drivers.PinOutput abstraction which is a function type. Pros and cons
// of both approaches have been discussed in the [relevant issue]. PinOutput should only be used
// to expose an initialization function of a driver that receives pins of this type. Ideally
// driver developers should also expose the initialization with drivers.Pin type:
//
//	func New(p1, p2, p3 legacy.PinOutput) *Device {
//		return NewWithPinfuncs(p1.Set, p2.Set, p3.Set)
//	}
//
//	func NewWithPinfuncs(p1, p2, p3 drivers.PinOutput) *Device {
//		return &Device{p1:p1, p2:p2, p3:p3}
//	}
//
// [relevant issue]: https://github.com/tinygo-org/drivers/pull/749/files
type PinOutput interface {
	Set(level bool)
}

// PinInput represents a pin hardware abstraction layer. See [PinOutput] for
// more information on why this is "legacy".
type PinInput interface {
	Get() (level bool)
}

// ConfigurePinOut is a legacy function used to configure pins as outputs.
//
// Deprecated: Do not configure pins in drivers.
// This is a legacy feature and should only be used by drivers that
// previously configured pins in initialization to avoid breaking users.
func ConfigurePinOut(po PinOutput) {
	configurePinOut(po)
}

// ConfigurePinInput is a legacy function used to configure pins as inputs.
//
// Deprecated: Do not configure pins in drivers.
// This is a legacy feature and should only be used by drivers that
// previously configured pins in initialization to avoid breaking users.
func ConfigurePinInputPulldown(pi PinInput) {
	configurePinInputPulldown(pi)
}

// ConfigurePinInput is a legacy function used to configure pins as inputs.
//
// Deprecated: Do not configure pins in drivers.
// This is a legacy feature and should only be used by drivers that
// previously configured pins in initialization to avoid breaking users.
func ConfigurePinInput(pi PinInput) {
	configurePinInput(pi)
}

// ConfigurePinInput is a legacy function used to configure pins as inputs.
//
// Deprecated: Do not configure pins in drivers.
// This is a legacy feature and should only be used by drivers that
// previously configured pins in initialization to avoid breaking users.
func ConfigurePinInputPullup(pi PinInput) {
	configurePinInputPullup(pi)
}

var (
	ErrConfigBeforeInstantiated = errors.New("device must be instantiated with New before calling Configure method")
)
