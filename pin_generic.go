//go:build !tinygo

// Generic no-op implementations of SafePin, SafePinInput and SafePinOutput.
// It's custom implementation's responsibility to configure the pin modes correctly when needed.

package drivers

func SafePinInput(pin PinInput) PinInput {
	return pin
}

func SafePinOutput(pin PinOutput) PinOutput {
	return pin
}

func SafePin(pin Pin) Pin {
	return pin
}
