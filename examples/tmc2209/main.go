package main

import (
	"machine"

	"tinygo.org/x/drivers/tmc2209"
)

func main() {
	uart := machine.UART0
	comm := tmc2209.NewUARTComm(*uart, 0)
	// Create an instance of the TMC2209 with UART communication
	tmc := tmc2209.NewTMC2209(comm, 0x00) // Replace 0x00 with the appropriate address

	// Set up the TMC2209 driver
	err := tmc.Setup()
	if err != nil {
		println("Failed to set up TMC2209: ", err)
	}

	// Write to a register (example: setting a register value)
	err = tmc.WriteRegister(0x10, 0x12345678) // Replace 0x10 with the register address and 0x12345678 with the value
	if err != nil {
		println("Failed to write register:", err)
	}

	// Read from a register (example: reading a register value)
	value, err := tmc.ReadRegister(0x10)
	if err != nil {
		println("Failed to read register: ", err)
	}

	// Output the read value
	println("Register value: ", value)
}
