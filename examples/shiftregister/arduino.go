// +build arduino

package main

import "machine"

const (
	latch = machine.Pin(12) // D12 Pin latch connected to ST_CP of 74HC595 (12)
	clock = machine.Pin(11) // D11 Pin clock connected to SH_CP of 74HC595 (11)
	data  = machine.Pin(10) // D10 Pin data connected to DS of 74HC595 (14)
)
