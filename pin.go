package drivers

// OutputPin represents a pin hardware abstraction layer for a pin that can output a digital signal.
type OutputPin interface {
	Set(level bool)
}

// InputPin represents a pin hardware abstraction layer for a pin that can read a digital input signal.
type InputPin interface {
	Get() (level bool)
}

// Pin represents a pin hardware abstraction layer that can both input and output digital signals.
type Pin interface {
	OutputPin
	InputPin
}
