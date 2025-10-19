//go:build !digispark && !arduino && !lint

package main

import "machine"

func init() {
	// Replace neo in the code below to match the pin
	// that you are using if different.
	neo = machine.WS2812
}
