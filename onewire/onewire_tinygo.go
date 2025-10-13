package onewire

import (
	"machine"

	"tinygo.org/x/drivers/internal/legacy"
)

// New creates a new GPIO 1-Wire connection.
// The pin must be pulled up to the VCC via a resistor greater than 500 ohms (default 4.7k).
func New(p machine.Pin) Device {
	isOut := false
	return Device{
		set: func(level bool) {
			if !isOut {
				legacy.ConfigurePinOut(p)
			}
			p.Set(level)
		},
		get: func() (level bool) {
			if isOut {
				legacy.ConfigurePinInputPullup(p)
			}
			return p.Get()
		},
	}
}
