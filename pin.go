package drivers

// PinOutput is hardware abstraction for a pin which outputs a
// digital signal (high or low level).
//
//	// Code conversion demo: from machine.Pin to drivers.PinOutput
//	led := machine.LED
//	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
//	var pin drivers.PinOutput = led.Set // Going from a machine.Pin to a drivers.PinOutput
type PinOutput func(level bool)

// PinInput is hardware abstraction for a pin which receives a
// digital signal and reads it (high or low level).
//
//	// Code conversion demo: from machine.Pin to drivers.PinInput
//	input := machine.LED
//	input.Configure(machine.PinConfig{Mode: machine.PinInputPulldown}) // or use machine.PinInputPullup or machine.PinInput
//	var pin drivers.PinInput = input.Get // Going from a machine.Pin to a drivers.PinInput
type PinInput func() (level bool)
