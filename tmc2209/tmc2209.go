//go:build tinygo

package tmc2209

import (
	"log"
)

// TMC2209 represents a single TMC2209 stepper motor driver on the communication line.
type TMC2209 struct {
	comm    RegisterComm
	address uint8
}

// NewTMC2209 creates a new instance of the TMC2209 driver for a specific address.
func NewTMC2209(comm RegisterComm, address uint8) *TMC2209 {
	return &TMC2209{
		comm:    comm,
		address: address,
	}
}

// Setup initializes the communication interface with the TMC2209.
func (driver *TMC2209) Setup() error {
	// Check if comm is of type *UARTComm
	if uartComm, ok := driver.comm.(*UARTComm); ok {
		// Call Setup only if comm is a *UARTComm
		err := uartComm.Setup()
		if err != nil {
			return CustomError("Failed to setup UART communication: " + err.Error())
		}
	} else {
		// If it's not a UARTComm, log that it's using a different communication method
		log.Println("Using a non-UART communication method")
	}

	return nil
}

// WriteRegister sends a register write command to the TMC2209.
func (driver *TMC2209) WriteRegister(reg uint8, value uint32) error {
	if driver.comm == nil {
		return CustomError("communication interface not set")
	}
	// Use the communication interface (RegisterComm) to write the register
	return driver.comm.WriteRegister(reg, value, driver.address)
}

// ReadRegister sends a register read command to the TMC2209 and returns the read value.
func (driver *TMC2209) ReadRegister(reg uint8) (uint32, error) {
	if driver.comm == nil {
		return 0, CustomError("communication interface not set")
	}
	// Use the communication interface (RegisterComm) to read the register
	return driver.comm.ReadRegister(reg, driver.address)
}
