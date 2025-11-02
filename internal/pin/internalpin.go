package pin

// TinyGo Pin HAL
//
// Pins are represented as function types instead of interfaces because they are:
//  - Faster on MCUs
//  - Conceptually simpler than an interface
//    - Easier to teach than interface for someone who is new to programming.
//  - Cleaner at call sites (e.g., isBusy := d.isBusy() vs d.isBusy.Get()).
//  - Enable inline funcs at the usage site- less boilerplate than defining
//     an interface type + method (see example below).
//  - Less prone to “method creep”: the API surface stays small while
//    advanced behavior can be composed in user code.
//    Example: a pin that can act as both output and input without adding methods.
//
//    var pinIsOutput bool
//    var po PinOutput = func(b bool) {
//    	if !pinIsOutput {
//    		pin.Configure(outputMode)
//    		pinIsOutput = true
//    	}
//    	pin.Set(b)
//    }
//    var pi PinInput = func() bool {
//    	if pinIsOutput {
//    		pin.Configure(inputMode)
//    		pinIsOutput = false
//    	}
//    	return pin.Get()
//    }
//
// See https://github.com/tinygo-org/drivers/pull/753 for detailed commentary.
// See https://github.com/orgs/tinygo-org/discussions/5043 for ongoing discussion

// Output is hardware abstraction for a pin which outputs a
// digital signal (high or low level).
// See [ongoing discussion here].
//
//	// Code conversion demo: from machine.Pin to pin.Output
//	led := machine.LED
//	led.Configure(machine.PinConfig{Mode: machine.Output})
//	var pin pin.Output = led.Set // Going from a machine.Pin to a pin.Output
//
// [ongoing discussion here]: https://github.com/orgs/tinygo-org/discussions/5043
type Output func(level bool)

// High sets the underlying pin's level to high. This is equivalent to calling PinOutput(true).
func (setPin Output) High() {
	setPin(true)
}

// Low sets the underlying pin's level to low. This is equivalent to calling PinOutput(false).
func (setPin Output) Low() {
	setPin(false)
}

// Input is hardware abstraction for a pin which receives a
// digital signal and reads it (high or low level).
//
//	// Code conversion demo: from machine.Pin to pin.Input
//	input := machine.LED
//	input.Configure(machine.PinConfig{Mode: machine.PinInputPulldown}) // or use machine.PinInputPullup or machine.Input
//	var pin pin.Input = input.Get // Going from a machine.Pin to a pin.Input
type Input func() (level bool)

// This file contains interface-style Pin HAL definition.
// It serves to eliminate machine.Pin from driver constructors
// so that drivers can be used in "big" Go projects where
// there is no machine package.

// OutputInterface represents a pin hardware abstraction layer for a pin that can output a digital signal.
//
// This is an alternative to [Output] abstraction which is a function type and has
// not been standardized as of yet as a standard HAL in the drivers package,
// [discussion ongoing here].
//
// [discussion ongoing here]: https://github.com/orgs/tinygo-org/discussions/5043
type OutputInterface interface {
	Set(level bool)
}

// InputInterface represents a pin hardware abstraction layer.
// See [OutputInterface] for more information on why this type exists separate to drivers.
type InputInterface interface {
	Get() (level bool)
}
