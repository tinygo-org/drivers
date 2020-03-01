package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/shiftregister"
)

func main() {
	d := shiftregister.New(
		shiftregister.EIGHT_BITS,
		machine.PA6, // D12 Pin latch connected to ST_CP of 74HC595 (12)
		machine.PA7, // D11 Pin clock connected to SH_CP of 74HC595 (11)
		machine.PB6, // D10 Pin data connected to DS of 74HC595 (14)
	)
	d.Configure()

	for {
		d.WriteMask(0x55)
		time.Sleep(100 * time.Millisecond)
		d.WriteMask(0xAA)
		time.Sleep(100 * time.Millisecond)
	}
}
