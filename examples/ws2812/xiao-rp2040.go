//go:build xiao_rp2040

package main

import "machine"

var neo = machine.D2 // Note, D10 does not work to drive WS2812 on this board
var led = machine.LED

// XIAO RP2040 has RGB led onboard that must be powered explicitly; we use its red led for default heartbeat.
func init() {
	machine.NEOPIXEL_POWER.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.NEOPIXEL_POWER.High()
}
