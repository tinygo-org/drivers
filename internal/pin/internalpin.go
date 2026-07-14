// package pin implements a TinyGo Pin HAL.
// It serves to eliminate machine.Pin from driver constructors
// so that drivers can be used in "big" Go projects where
// there is no machine package.
// This file contains both function and interface-style Pin HAL definitions.
package pin

// OutputFunc is hardware abstraction for a pin which outputs a
// digital signal (high or low level).
//
//	// Code conversion demo: from machine.Pin to pin.OutputFunc
//	led := machine.LED
//	led.Configure(machine.PinConfig{Mode: machine.Output})
//	var pin pin.OutputFunc = led.Set // Going from a machine.Pin to a pin.OutputFunc
//
// This is an alternative to [Output] which is an interface type.
type OutputFunc func(level bool)

// High sets the underlying pin's level to high. This is equivalent to calling PinOutput(true).
func (setPin OutputFunc) High() {
	setPin(true)
}

// Low sets the underlying pin's level to low. This is equivalent to calling PinOutput(false).
func (setPin OutputFunc) Low() {
	setPin(false)
}

// InputFunc is hardware abstraction for a pin which receives a
// digital signal and reads it (high or low level).
//
//	// Code conversion demo: from machine.Pin to pin.InputFunc
//	input := machine.LED
//	input.Configure(machine.PinConfig{Mode: machine.PinInputPulldown}) // or use machine.PinInputPullup or machine.Input
//	var pin pin.InputFunc = input.Get // Going from a machine.Pin to a pin.InputFunc
//
// This is an alternative to [Input] which is an interface type.
type InputFunc func() (level bool)

// // Below is an example on how to define a input/output pin HAL for a
// // pin that must switch between input and output mode:
//
// 	var pinIsOutput bool
// 	var po PinOutputFunc = func(b bool) {
// 		if !pinIsOutput {
// 			pin.Configure(outputMode)
// 			pinIsOutput = true
// 		}
// 		pin.Set(b)
// 	}
//
// 	var pi PinInputFunc = func() bool {
// 		if pinIsOutput {
// 			pin.Configure(inputMode)
// 			pinIsOutput = false
// 		}
// 		return pin.Get()
// 	}

// Output interface represents a pin hardware abstraction layer for a pin that can output a digital signal.
//
// This is an alternative to [OutputFunc] abstraction which is a function type.
type Output interface {
	Set(level bool)
}

// Input interface represents a pin hardware abstraction layer for a pin that can read a digital signal.
//
// This is an alternative to [InputFunc] abstraction which is a function type.
type Input interface {
	Get() (level bool)
}
