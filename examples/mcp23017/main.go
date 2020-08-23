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
	dev, err := mcp23017.NewI2C(machine.I2C0, 0x20)
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
	outputMask := ^mcp23017.Pins(1 << 0) // All except pin 0
	inputVal, err := input.Get()
	if err != nil {
		panic(err)
	}
	println("input value: ", inputVal)
	// Set the values of all the output pins.
	err = dev.SetPins(0b1011011_01101110, outputMask)
	if err != nil {
		panic(err)
	}
}
