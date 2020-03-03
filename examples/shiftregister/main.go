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

		// Examples using masks. This method writes all pins state at once.

		// All pins High
		d.WriteMask(0xFF)
		delay()

		// All pins Low
		d.WriteMask(0x00)
		delay()

		// Some fun with masks
		for _, pattern := range patterns {
			d.WriteMask(pattern)
			shortDelay()
		}
		delay()
		d.WriteMask(0x00)

		// Examples using individually addressable pin API. This method is slower than using mask
		// because all register's pins state are send for is p.Set() call.

		// Set register's pin #4
		d.GetShiftPin(4).High()
		delay()
		d.GetShiftPin(4).Low()
		delay()

		// Get an individual pin and use it
		pin := d.GetShiftPin(7)
		pin.High()
		delay()
		pin.Low()
		delay()

		// Prepare an array of pin attached to the register
		pins := [8]*shiftregister.ShiftPin{}
		for p := 0; p < 8; p++ {
			pins[p] = d.GetShiftPin(p)
		}

		for p := 7; p >= 0; p-- {
			pins[p].Low()
			shortDelay()
			pins[p].High()
		}

		for p := 7; p >= 0; p-- {
			pins[p].High()
			time.Sleep(100 * time.Millisecond)
			pins[p].Low()
		}
		delay()
	}
}

func delay() {
	time.Sleep(500 * time.Millisecond)
}
func shortDelay() {
	time.Sleep(100 * time.Millisecond)
}

var patterns = []uint32{
	0b00000001,
	0b00000010,
	0b00000100,
	0b00001000,
	0b00010000,
	0b00100000,
	0b01000000,
	0b10000000,
	0b10000001,
	0b10000010,
	0b10000100,
	0b10001000,
	0b10010000,
	0b10100000,
	0b11000000,
	0b11000001,
	0b11000010,
	0b11000100,
	0b11001000,
	0b11010000,
	0b11100000,
	0b11100001,
	0b11100010,
	0b11100100,
	0b11101000,
	0b11110000,
	0b11110001,
	0b11110010,
	0b11110100,
	0b11111000,
	0b11111001,
	0b11111010,
	0b11111100,
	0b11111101,
	0b11111110,
	0b11111111,
	0b00000000,
	0b11111111,
	0b00000000,
	0b11111111,
}
