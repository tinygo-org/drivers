//go:build xiao

// This example does not really work in the moment of writing.
// Some issue seem to exist in code for SAMD21 chip.
// See also example for Arduino Nano 33 IoT board (does not work either).

package main

import "machine"

var neo = machine.D2
var led = machine.LED
