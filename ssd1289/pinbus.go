package ssd1289

import (
	"tinygo.org/x/drivers/internal/legacy"
	"tinygo.org/x/drivers/internal/pin"
)

type pinBus struct {
	pins [16]pin.Output
}

func NewPinBus(pins [16]pin.Output) pinBus {

	// configure GPIO pins (on baremetal targets only, for backwards compatibility)
	for i := 0; i < 16; i++ {
		legacy.ConfigurePinOut(pins[i])
	}

	return pinBus{
		pins: pins,
	}
}

func (b pinBus) Set(data uint16) {
	for i := 15; i >= 0; i-- {
		b.pins[i].Set((data & (1 << i)) != 0)
	}
}
