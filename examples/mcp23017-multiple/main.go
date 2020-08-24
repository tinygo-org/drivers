// This example demonstrates putting several mcp23017 devices together into
// a single virtual I/O array.
package main

import (
	"machine"

	"tinygo.org/x/drivers/mcp23017"
)

func main() {
	err := machine.I2C0.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_400KHZ,
	})
	if err != nil {
		panic(err)
	}
	// Assume the devices are at addresses 0x20, 0x21
	dev, err := mcp23017.NewI2CDevices(machine.I2C0, 0x20, 0x21)
	if err != nil {
		panic(err)
	}
	// Configure pin 0 for input and all the others for output.
	if err := dev.SetModes([]mcp23017.PinMode{
		mcp23017.Input | mcp23017.Pullup,
		mcp23017.Output,
	}); err != nil {
		panic(err)
	}
	input := dev.Pin(0)
	// Make a mask that represents all the output pins.
	// Note that this leverages the driver behaviour which replicates the highest bit in
	// the last slice element (1 in this case) to all other pins
	outputMask := mcp23017.PinSlice{^mcp23017.Pins(1 << 0)} // All except pin 0
	inputVal, err := input.Get()
	if err != nil {
		panic(err)
	}
	println("input value: ", inputVal)
	// Set the values of all the output pins.
	err = dev.SetPins(mcp23017.PinSlice{0b1011011_01101110, 0b11111101_11100110}, outputMask)
	if err != nil {
		panic(err)
	}
}
