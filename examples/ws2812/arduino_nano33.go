//go:build arduino_nano33

// This example does not really work in the moment of writing.
// Some issue seem to exist in code for SAMD21 chip.
// See also example for XIAO board (does not work either).

package main

import "machine"

var neo = machine.D2
var led = machine.LED
