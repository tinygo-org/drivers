// +build nucleof103rb

package main

import "machine"

const (
	latch = machine.PA6 // D12 Pin latch connected to ST_CP of 74HC595 (12)
	clock = machine.PA7 // D11 Pin clock connected to SH_CP of 74HC595 (11)
	data  = machine.PB6 // D10 Pin data connected to DS of 74HC595 (14)
)
