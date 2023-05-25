//go:build !digispark && !arduino && !qtpy && !m5stamp_c3 && !thingplus_rp2040 && !nano_rp2040 && !xiao && !xiao_rp2040 && !arduino_nano33

package main

import "machine"

// Replace neo and led in the code below to match the pin
// that you are using if different.
var neo machine.Pin = machine.WS2812
var led = machine.LED
