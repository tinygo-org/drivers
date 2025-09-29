package drivers

// TinyGo Pin HAL (exported)
//
// Pins are represented as function types instead of interfaces because they are:
//  - Faster on MCUs
//  - Conceptually simpler than an interface
//    - Very likely easier to teach than interface:
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

// PinOutput is hardware abstraction for a pin which outputs a
// digital signal (high or low level).
//
//	// Code conversion demo: from machine.Pin to drivers.PinOutput
//	led := machine.LED
//	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
//	var pin drivers.PinOutput = led.Set // Going from a machine.Pin to a drivers.PinOutput
type PinOutput func(level bool)

// High sets the underlying pin's level to high. This is equivalent to calling PinOutput(true).
func (setPin PinOutput) High() {
	setPin(true)
}

// Low sets the underlying pin's level to low. This is equivalent to calling PinOutput(false).
func (setPin PinOutput) Low() {
	setPin(false)
}

// PinInput is hardware abstraction for a pin which receives a
// digital signal and reads it (high or low level).
//
//	// Code conversion demo: from machine.Pin to drivers.PinInput
//	input := machine.LED
//	input.Configure(machine.PinConfig{Mode: machine.PinInputPulldown}) // or use machine.PinInputPullup or machine.PinInput
//	var pin drivers.PinInput = input.Get // Going from a machine.Pin to a drivers.PinInput
type PinInput func() (level bool)
