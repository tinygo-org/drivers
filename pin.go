package drivers

// Pin represents a GPIO pin. It is implemented by the machine.Pin type.
type Pin interface {
	// Get gets the current GPIO pin value
	Get() bool

	// Set sets the GPIO pin to either high if true, or low if false
	Set(bool)

	// High sets this GPIO pin to high, assuming it has been configured as an output
	// pin. It is hardware dependent (and often undefined) what happens if you set a
	// pin to high that is not configured as an output pin.
	High()

	// Low sets this GPIO pin to low, assuming it has been configured as an output
	// pin. It is hardware dependent (and often undefined) what happens if you set a
	// pin to low that is not configured as an output pin.
	Low()
}
