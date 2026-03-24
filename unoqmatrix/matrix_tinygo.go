//go:build baremetal

package unoqmatrix

import (
	"machine"

	pin "tinygo.org/x/drivers/internal/pin"
)

// NewFromBasePin creates a Device from a base machine.Pin.
// It constructs 11 CharlieplexPin values from consecutive pins starting at basePin.
// Each pin lazily switches between output and input mode as needed.
func NewFromBasePin(basePin machine.Pin) Device {
	var pins [numPins]CharlieplexPin
	for i := range pins {
		p := basePin + machine.Pin(i)
		var isOutput bool
		pins[i] = CharlieplexPin{
			Set: pin.OutputFunc(func(level bool) {
				if !isOutput {
					p.Configure(machine.PinConfig{Mode: machine.PinOutput})
					isOutput = true
				}
				p.Set(level)
			}),
			Float: func() {
				if isOutput {
					p.Configure(machine.PinConfig{Mode: machine.PinInput})
					isOutput = false
				}
			},
		}
	}
	return New(pins)
}
