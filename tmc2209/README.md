# TMC2209 Go Package

This package provides a lightweight interface to communicate with the TMC2209 stepper motor driver via UART. It is designed to be used with microcontrollers and supports both default UART communication and custom interface implementations for flexibility.
Currently this package only implements the communications with TMC2209. Functions to handle specific operations such as EnableStealthChop() etc. will need to be implemented by the user. 
## Features
- **UART Communication:** Standard communication through UART for controlling the TMC2209 driver.
- **Custom Communication Interfaces:** Allows custom implementations of communication interfaces (e.g., USB, SPI, etc.) via the `RegisterComm` interface. The usecase for this is if you have host system that can "talk" to the microcontroller connected to the TMC2209
- **Error Handling:** Lightweight error handling for TinyGo compatibility.
- **Register Access:** Provides methods to read and write registers on the TMC2209.

## Setup

### Prerequisites
- **TinyGo**: This package is optimized for TinyGo, which is suitable for running Go on microcontrollers and embedded systems.
- **UART Communication**: For microcontrollers with UART support, such as ESP32, STM32, and Raspberry Pi Pico.


## Usage
### Using with Microcontrollers (Default UART)

To use the package with a microcontroller, you need to initialize the UART communication and configure the TMC2209 driver.
Example:

```aiignore
package main

import (
	"fmt"
	"log"
	"github.com/yourusername/tmc2209"
	"machine"
)

func main() {
	uart := machine.UART0

	// Create an instance of the TMC2209 with UART communication
	tmc := tmc2209.NewTMC2209(uart, 0x00) // Replace 0x00 with the appropriate address

	// Set up the TMC2209 driver
	err := tmc.Setup()
	if err != nil {
		log.Fatalf("Failed to set up TMC2209: %v", err)
	}

	// Write to a register (example: setting a register value)
	err = tmc.WriteRegister(0x10, 0x12345678) // Replace 0x10 with the register address and 0x12345678 with the value
	if err != nil {
		log.Fatalf("Failed to write register: %v", err)
	}

	// Read from a register (example: reading a register value)
	value, err := tmc.ReadRegister(0x10)
	if err != nil {
		log.Fatalf("Failed to read register: %v", err)
	}

	// Output the read value
	fmt.Printf("Register value: 0x%X\n", value)
}

```
## Microcontroller Notes

- **TinyGo Support:** This code is optimized for use with TinyGo on supported microcontrollers.
- **UART Configuration:** Ensure the UART instance is configured correctly for your microcontroller. 
- The machine.UART0 in the example is for the default UART on TinyGo-compatible devices like the Raspberry Pi Pico. Check your microcontroller's documentation for the correct UART instance and pin configuration.

## 2. Custom Interface Implementation

If you want to use a custom communication interface (e.g., USB, SPI, etc.), you can implement the RegisterComm interface.
Custom Interface Example (e.g., USB Communication):

```
package main

import (
	"fmt"
	"log"
	"github.com/yourusername/tmc2209"
	"yourcustompackage" // Custom package for communication
)

type CustomComm struct {
	// Implement your custom communication method here
}

func (c *CustomComm) ReadRegister(register uint8, driverIndex uint8) (uint32, error) {
	// Implement the register read logic using your custom interface (USB, SPI, etc.)
	// For example, send the register read command over USB and read the response
	return 0, nil
}

func (c *CustomComm) WriteRegister(register uint8, value uint32, driverIndex uint8) error {
	// Implement the register write logic using your custom interface (USB, SPI, etc.)
	// For example, send the register write command over USB
	return nil
}

func main() {
	// Create an instance of the TMC2209 with your custom communication interface
	customComm := &CustomComm{}
	tmc := tmc2209.NewTMC2209(customComm, 0x00) // Replace 0x00 with the appropriate address

	// Set up the TMC2209 driver
	err := tmc.Setup()
	if err != nil {
		log.Fatalf("Failed to set up TMC2209: %v", err)
	}

	// Write to a register (example: setting a register value)
	err = tmc.WriteRegister(0x10, 0x12345678) // Replace 0x10 with the register address and 0x12345678 with the value
	if err != nil {
		log.Fatalf("Failed to write register: %v", err)
	}

	// Read from a register (example: reading a register value)
	value, err := tmc.ReadRegister(0x10)
	if err != nil {
		log.Fatalf("Failed to read register: %v", err)
	}

	// Output the read value
	fmt.Printf("Register value: 0x%X\n", value)
}

```

### Custom Interface Notes:

- RegisterComm Interface: Your custom communication interface (e.g., USB, SPI) should implement the RegisterComm interface. This ensures that the TMC2209 driver can interact with your custom interface for reading and writing registers.
- Error Handling: Implement proper error handling in your custom interface methods (ReadRegister and WriteRegister).

#### Functions
```
NewTMC2209(comm RegisterComm, address uint8) *TMC2209
```
Creates a new TMC2209 instance with the provided communication interface (comm) and driver address (address).

```Setup() error```

Initializes the communication interface. This is required before interacting with the TMC2209 driver. The method checks if the communication interface is UART and sets it up accordingly.

```WriteRegister(register uint8, value uint32) error```

Writes a value to a specific register on the TMC2209 driver.

```ReadRegister(register uint8) (uint32, error)```

Reads a value from a specific register on the TMC2209 driver.

### Error Handling

This package uses a lightweight custom error type (CustomError) to ensure compatibility with TinyGo and reduce external dependencies.
Example Error:

```CustomError("communication interface not set")```


Created by Amken3d
info@amken3d.us

 
