package drivers

// PinOutput is hardware abstraction for a pin which outputs a
// digital signal (high or low voltage).
type PinOutput func(level bool)

// PinInput is hardware abstraction for a pin which receives a
// digital signal and reads it (high or low voltage).
type PinInput func() (level bool)
