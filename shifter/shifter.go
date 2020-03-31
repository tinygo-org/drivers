// Package shifter is for 8bit shift register, most common are 74HC165 and 74165
package shifter // import "tinygo.org/x/drivers/shifter"

import (
	"errors"
	"machine"
)

const (
	EIGHT_BITS     NumberBit = 8
	SIXTEEN_BITS   NumberBit = 16
	THIRTYTWO_BITS NumberBit = 32
)

type NumberBit int8

// Device holds the Pins.
type Device struct {
	latch machine.Pin
	clk   machine.Pin
	out   machine.Pin
	Pins  []ShiftPin
	bits  NumberBit
}

// ShiftPin is the implementation of the ShiftPin interface.
type ShiftPin struct {
	pin     machine.Pin
	d       *Device
	pressed bool
}

// New returns a new shifter driver given the correct pins.
func New(numBits NumberBit, latch, clk, out machine.Pin) Device {
	return Device{
		latch: latch,
		clk:   clk,
		out:   out,
		Pins:  make([]ShiftPin, int(numBits)),
		bits:  numBits,
	}
}

// Configure here just for interface compatibility.
func (d *Device) Configure() {
	d.latch.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.clk.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.out.Configure(machine.PinConfig{Mode: machine.PinInput})
	for i := 0; i < int(d.bits); i++ {
		d.Pins[i] = d.GetShiftPin(i)
	}
}

// GetShiftPin returns an ShiftPin for a specific input.
func (d *Device) GetShiftPin(input int) ShiftPin {
	return ShiftPin{pin: machine.Pin(input), d: d}
}

// Read8Input updates the internal pins' states and returns it as an uint8.
func (d *Device) Read8Input() (uint8, error) {
	if d.bits != EIGHT_BITS {
		return 0, errors.New("wrong amount of registers")
	}
	return uint8(d.readInput(EIGHT_BITS)), nil
}

// Read16Input updates the internal pins' states and returns it as an uint16.
func (d *Device) Read16Input() (uint16, error) {
	if d.bits != SIXTEEN_BITS {
		return 0, errors.New("wrong amount of registers")
	}
	return uint16(d.readInput(SIXTEEN_BITS)), nil
}

// Read32Input updates the internal pins' states and returns it as an uint32.
func (d *Device) Read32Input() (uint32, error) {
	if d.bits != THIRTYTWO_BITS {
		return 0, errors.New("wrong amount of registers")
	}
	return d.readInput(THIRTYTWO_BITS), nil
}

// Get the pin's state for a specific ShiftPin.
// Read{8|16|32}Input should be called before to update the state. Read{8|16|32}Input updates
// all the pins, no need to call it for each pin individually.
func (p ShiftPin) Get() bool {
	return p.pressed
}

// Configure here just for interface compatibility.
func (p ShiftPin) Configure() {
}

// readInput reads howMany bits from the shift register and updates the internal pins' states.
func (d *Device) readInput(howMany NumberBit) uint32 {
	d.latch.High()
	var data uint32
	for i := howMany - 1; i >= 0; i-- {
		d.clk.Low()
		if d.out.Get() {
			data |= 1 << i
			d.Pins[i].pressed = true
		} else {
			d.Pins[i].pressed = false
		}
		d.clk.High()
	}
	d.latch.Low()
	return data
}
