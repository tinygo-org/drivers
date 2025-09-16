package pin

import "tinygo.org/x/drivers"

// Here to aid relevant documentation links of [drivers.PinOutput] and [drivers.PinInput].
var _ drivers.PinOutput

// Output represents a pin hardware abstraction layer for a pin that can output a digital signal.
//
// This is an alternative to [drivers.PinOutput] abstraction which is a function type and has
// not been standardized as of yet as a standard HAL in the drivers package,
// [discussion ongoing here].
//
// [discussion ongoing here]: https://github.com/orgs/tinygo-org/discussions/5043
type Output interface {
	Set(level bool)
}

// Input represents a pin hardware abstraction layer.
// See [Output] for more information on why this type exists separate to drivers.
type Input interface {
	Get() (level bool)
}
