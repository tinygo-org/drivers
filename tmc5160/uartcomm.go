//go:build uart

package tmc5160

import (
	"machine"
	"time"
)

// UARTComm implements RegisterComm for UART-based communication with Driver.
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

// Setup initializes the UART communication with the Driver.
func (comm *UARTComm) Setup() error {
	if comm.uart == (machine.UART{}) {
		return CustomError("UART not initialized")
	}
	err := comm.uart.Configure(machine.UARTConfig{
		BaudRate: 115200,
	})
	if err != nil {
		return CustomError("Failed to configure UART")
	}
	return nil
}

// WriteRegister sends a register write command to the Driver.
// Prepare the data packet (sync byte + address + register + data + checksum)
func (comm *UARTComm) WriteRegister(register uint8, value uint32, driverIndex uint8) error {

	buffer := []byte{
		0x05,                       // Sync byte
		comm.address,               // Slave address
		register | 0x80,            // Write command (MSB set to 1 for write)
		byte((value >> 24) & 0xFF), // MSB of value
		byte((value >> 16) & 0xFF), // Middle byte
		byte((value >> 8) & 0xFF),  // Next byte
		byte(value & 0xFF),         // LSB of value
	}
	checksum := byte(0)
	for _, b := range buffer[:7] {
		checksum ^= b
	}
	buffer[7] = checksum // Set checksum byte

	// Write the data to the Driver
	done := make(chan error, 1)

	go func() {
		comm.uart.Write(buffer)
		done <- nil
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(100 * time.Millisecond): // Timeout after 100ms
		return CustomError("write timeout")
	}
}

// ReadRegister sends a register read command to the Driver.
func (comm *UARTComm) ReadRegister(register uint8, driverIndex uint8) (uint32, error) {
	// Prepare the read command (sync byte + address + register + checksum)
	var writeBuffer [4]byte
	writeBuffer[0] = 0x05                                             // Sync byte
	writeBuffer[1] = comm.address                                     // Slave address
	writeBuffer[2] = register & 0x7F                                  // Read command (MSB clear for read)
	writeBuffer[3] = writeBuffer[0] ^ writeBuffer[1] ^ writeBuffer[2] // Checksum
	done := make(chan []byte, 1)
	go func() {
		comm.uart.Write(writeBuffer[:])
		readBuffer := make([]byte, 8) // Prepare the buffer to read 8 bytes
		comm.uart.Read(readBuffer)
		done <- readBuffer
	}()
	select {
	case readBuffer := <-done:
		checksum := byte(0)
		for i := 0; i < 7; i++ {
			checksum ^= readBuffer[i]
		}
		if checksum != readBuffer[7] {
			return 0, CustomError("checksum error")
		}
		return uint32(readBuffer[3])<<24 | uint32(readBuffer[4])<<16 | uint32(readBuffer[5])<<8 | uint32(readBuffer[6]), nil
	case <-time.After(100 * time.Millisecond): // Timeout after 100ms
		return 0, CustomError("read timeout")
	}
}
