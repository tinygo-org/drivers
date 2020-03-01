package main

import (
	"time"

	"tinygo.org/x/drivers/shiftregister"
)

func main() {
	d := shiftregister.New(
		shiftregister.EIGHT_BITS,
		latch, // D12 Pin latch connected to ST_CP of 74HC595 (12)
		clock, // D11 Pin clock connected to SH_CP of 74HC595 (11)
		data,  // D10 Pin data connected to DS of 74HC595 (14)
	)
	d.Configure()

	for {
		d.WriteMask(0x50)
		time.Sleep(500 * time.Millisecond)
		d.WriteMask(0xA0)
		time.Sleep(500 * time.Millisecond)
		d.WriteMask(0x05)
		time.Sleep(500 * time.Millisecond)
		d.WriteMask(0x0A)
		time.Sleep(500 * time.Millisecond)
	}
}
