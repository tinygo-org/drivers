//go:build arduino_uno_r4_wifi

package unor4matrix

import "machine"

// NewOnboard returns a device connected to the UNO R4 WiFi's built-in matrix.
func NewOnboard() Device {
	// Matrix line order from the UNO R4 WiFi variant.
	// https://github.com/arduino/ArduinoCore-renesas/blob/424e86eff92d37f72123c2b641dd8bbf06a38b47/variants/UNOWIFIR4/variant.cpp
	lines := [...]machine.Pin{
		machine.P0_03, machine.P0_04, machine.P0_11,
		machine.P0_12, machine.P0_13, machine.P0_15,
		machine.P2_04, machine.P2_05, machine.P2_06,
		machine.P2_12, machine.P2_13,
	}
	var pins [numPins]CharlieplexPin
	for i, p := range lines {
		pins[i] = CharlieplexPin{
			Set: func(level bool) {
				// Preload PODR before enabling PDR to avoid driving the previous level.
				// RA4M1 R01UH0887EJ0110, sections 19.2.1, 19.2.3, 19.2.5, pp. 357, 359, 361-362.
				p.Set(level)
				p.Configure(machine.PinConfig{Mode: machine.PinOutput})
			},
			Float: func() { p.Configure(machine.PinConfig{Mode: machine.PinInput}) },
		}
	}
	return New(pins)
}
