//go:build thingplus_rp2040

package main

import "machine"

// This is the pin assignment for the internal neopixel of the
// Sparkfun thingplus rp2040.
// Replace neo and led in the code below to match the pin
// that you are using if different.
var neo machine.Pin = machine.GPIO8
var led = machine.LED
