package pin

import "errors"

// OutputFunc is hardware abstraction for a pin which outputs a
// digital signal (high or low level).
//
//	// Code conversion demo: from machine.Pin to pin.OutputFunc
//	led := machine.LED
//	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
//	var p pin.OutputFuncFunc = led.Set // Going from a machine.Pin to a pin.OutputFunc
type OutputFunc func(level bool)

func (o OutputFunc) High() {
	o(true)
}

func (o OutputFunc) Low() {
	o(false)
}

// InputFunc is hardware abstraction for a pin which receives a
// digital signal and reads it (high or low level).
//
//	// Code conversion demo: from machine.Pin to pin.InputFunc
//	input := machine.LED
//	input.Configure(machine.PinConfig{Mode: machine.PinInputPulldown}) // or use machine.PinInputPullup or machine.PinInput
//	var p pin.InputFunc = input.Get // Going from a machine.Pin to a drivers.PinInput
type InputFunc func() (level bool)

// PinOutput represents a pin hardware abstraction layer for a pin that can output a digital signal.
// This is an wrapper to pin.OutputFunc abstraction which is a function type.
//
//	func New(p1, p2, p3 pin.Output) *Device {
//		return NewWithPinfuncs(p1.Set, p2.Set, p3.Set)
//	}
//
//	func NewWithPinfuncs(p1, p2, p3 pin.OutputFunc) *Device {
//		return &Device{p1:p1, p2:p2, p3:p3}
//	}
//
// [relevant issue]: https://github.com/tinygo-org/drivers/pull/749/files
type Output interface {
	Set(level bool)
}

// PinInput represents a pin hardware abstraction layer. See [PinOutput] for
// more information on why this is "legacy".
type Input interface {
	Get() (level bool)
}

// ConfigureOutput is a legacy function used to configure pins as outputs.
//
// Deprecated: You should not configure pins in drivers.
// This is a legacy feature and should only be used by drivers that
// previously configured pins in initialization to avoid breaking users.
func ConfigureOutput(po Output) {
	configureOutput(po)
}

// ConfigureInput is a legacy function used to configure pins as inputs.
//
// Deprecated: You should not configure pins in drivers.
// This is a legacy feature and should only be used by drivers that
// previously configured pins in initialization to avoid breaking users.
func ConfigureInputPulldown(pi Input) {
	configureInputPulldown(pi)
}

// ConfigureInput is a legacy function used to configure pins as inputs.
//
// Deprecated: You should not configure pins in drivers.
// This is a legacy feature and should only be used by drivers that
// previously configured pins in initialization to avoid breaking users.
func ConfigureInput(pi Input) {
	configureInput(pi)
}

// ConfigureInputPullup is a legacy function used to configure pins as inputs.
//
// Deprecated: You should not configure pins in drivers.
// This is a legacy feature and should only be used by drivers that
// previously configured pins in initialization to avoid breaking users.
func ConfigureInputPullup(pi Input) {
	configureInputPullup(pi)
}

// IsNotPin returns true if the argument is a machine.Pin type and is the machine.NoPin predeclared type.
//
// Deprecated: Drivers should not require pin knowledge.
func IsNotPin(pin any) bool {
	return isNotPin(pin)
}

var (
	ErrConfigBeforeInstantiated = errors.New("device must be instantiated with New before calling Configure method")
)
