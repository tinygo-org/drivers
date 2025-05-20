//go:build tinygo

package tmc2209

import (
	"machine"
	"time"
)

// CustomError is a lightweight error type used for TinyGo compatibility.
type CustomError string

func (e CustomError) Error() string {
	return string(e)
}

// UARTComm implements RegisterComm for UART-based communication
type UARTComm struct {
	uart    machine.UART
	address uint8
}

// NewUARTComm creates a new UARTComm instance.
func NewUARTComm(uart machine.UART, address uint8) *UARTComm {
	return &UARTComm{
		uart:    uart,
		address: address,
	}
}

// Setup initializes the UART communication with the TMC2209.
func (comm *UARTComm) Setup() error {
	// Check if UART is initialized
	if comm.uart == (machine.UART{}) {
		return CustomError("UART not initialized")
	}

	// Configure the UART interface with the desired baud rate and settings
	err := comm.uart.Configure(machine.UARTConfig{
		BaudRate: 115200,
	})
	if err != nil {
		return CustomError("Failed to configure UART")
	}

	// No built-in timeout in TinyGo, so timeout will be handled in the read/write methods
	return nil
}

// WriteRegister sends a register write command to the TMC2209 with a timeout.
func (comm *UARTComm) WriteRegister(register uint8, value uint32, driverIndex uint8) error {
	buffer := []byte{
		0x05,                       // Sync byte
		comm.address,               // Slave address
		register | 0x80,            // Write command (set MSB to 1 for write)
		byte((value >> 24) & 0xFF), // MSB of value
		byte((value >> 16) & 0xFF), // Middle byte
		byte((value >> 8) & 0xFF),  // Next byte
		byte(value & 0xFF),         // LSB of value
		0,                          // CRC
	}

	buffer[7] = CalculateCRC(buffer[:7])

	// Write the data to the TMC2209
	done := make(chan error, 1)

	go func() {
		comm.uart.Write(buffer)
		done <- nil
	}()

	// Implementing timeout using a 100ms timer
	select {
	case err := <-done:
		return err
	case <-time.After(100 * time.Millisecond): // Timeout after 100ms
		return CustomError("write timeout")
	}
}

// ReadRegister sends a register read command to the TMC2209 with a timeout.
func (comm *UARTComm) ReadRegister(register uint8, driverIndex uint8) (uint32, error) {
	var writeBuffer [4]byte
	writeBuffer[0] = 0x05            // Sync byte
	writeBuffer[1] = 0x00            // Slave address
	writeBuffer[2] = register & 0x7F // Read command (MSB clear for read)
	writeBuffer[3] = CalculateCRC(writeBuffer[:3])

	// Send the read command
	done := make(chan []byte, 1)
	go func() {
		comm.uart.Write(writeBuffer[:])
		readBuffer := make([]byte, 8)
		comm.uart.Read(readBuffer)
		done <- readBuffer
	}()

	// Implementing timeout using a 100ms timer
	select {
	case readBuffer := <-done:
		checksum := CalculateCRC(readBuffer[:7])
		if checksum != readBuffer[7] {
			return 0, CustomError("checksum error")
		}

		// Return the value from the register
		return uint32(readBuffer[3])<<24 | uint32(readBuffer[4])<<16 | uint32(readBuffer[5])<<8 | uint32(readBuffer[6]), nil
	case <-time.After(100 * time.Millisecond): // Timeout after 100ms
		return 0, CustomError("read timeout")
	}
}
