//go:build digispark

package main

import "machine"

func init() {
	// This is the pin assignment for the Digispark only.
	// Replace neo and led in the code below to match the pin
	// that you are using if different.
	neo = machine.Pin(0)
}
