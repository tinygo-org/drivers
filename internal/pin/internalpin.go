package pin

// OutputFn is a function that sets the underlying pin's level.
type OutputFn func(level bool)

// High sets the underlying pin's level to high. This is equivalent to calling PinOutput(true).
func (setPin OutputFn) High() {
	setPin(true)
}

// Low sets the underlying pin's level to low. This is equivalent to calling PinOutput(false).
func (setPin OutputFn) Low() {
	setPin(false)
}

// InputFn is a function that reads the underlying pin's level.
type InputFn func() (level bool)
