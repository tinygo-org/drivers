package drivers

// PinOutput represents a pin hardware abstraction layer for a pin that can output a digital signal.
type PinOutput interface {
	Set(level bool)
}

// PinInput represents a pin hardware abstraction layer.
type PinInput interface {
	Get() (level bool)
}

// PinOutputFunc is hardware abstraction for a function that causes pin to output a
// digital signal (high or low level).
//
//	// Code conversion demo: from machine.Pin to drivers.PinOutputFunc
//	led := machine.LED
//	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
//	var pin drivers.PinOutputFunc = led.Set // Going from a machine.Pin to a drivers.PinOutputFunc
type PinOutputFunc func(level bool)

// High sets the underlying pin's level to high. This is equivalent to calling PinOutput(true).
func (po PinOutputFunc) High() {
	po(true)
}

// Low sets the underlying pin's level to low. This is equivalent to calling PinOutput(false).
func (po PinOutputFunc) Low() {
	po(false)
}

// PinInputFunc is hardware abstraction for a pin which receives a
// digital signal and reads it (high or low level).
//
//	// Code conversion demo: from machine.Pin to drivers.PinInputFunc
//	input := machine.LED
//	input.Configure(machine.PinConfig{Mode: machine.PinInputPulldown}) // or use machine.PinInputPullup or machine.PinInput
//	var pin drivers.PinInputFunc = input.Get // Going from a machine.Pin to a drivers.PinInputFunc
type PinInputFunc func() (level bool)
