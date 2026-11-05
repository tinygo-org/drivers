//go:build teensy36

package main

import "machine"

func init() {
	// The Teensy 3.6 has no onboard LEDs, so drive an external strip.
	neo = machine.D02
}
