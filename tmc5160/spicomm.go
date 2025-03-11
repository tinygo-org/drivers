//go:build tinygo

package tmc5160

import (
	"machine"
	"time"
)

// CustomError is a lightweight error type used for TinyGo compatibility.
type CustomError string

func (e CustomError) Error() string {
	return string(e)
}

// SPIComm implements RegisterComm for SPI-based communication
type SPIComm struct {
	spi    *machine.SPI
	CsPins map[uint8]machine.Pin // Map to store CS pin for each Driver by its address
}

// NewSPIComm creates a new SPIComm instance.
func NewSPIComm(spi *machine.SPI, csPins map[uint8]machine.Pin) *SPIComm {
	return &SPIComm{
		spi:    spi,
		CsPins: csPins,
	}
}

// Setup initializes the SPI communication with the Driver and configures all CS pins.
func (comm *SPIComm) Setup() error {
	// Check if SPI is initialized
	if comm.spi == nil {
		return CustomError("SPI not initialized")
	}

	// Configure all CS pins (make them output and set them high)
	for _, csPin := range comm.CsPins {
		csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		csPin.High() // Set all CS pins high initially
	}

	// Configure the SPI interface with the desired settings
	err := comm.spi.Configure(machine.SPIConfig{
		LSBFirst: false,
		Mode:     3,
	})
	if err != nil {
		return CustomError("Failed to configure SPI")
	}

	return nil
}

// WriteRegister sends a register write command to the TMC5160.
func (comm *SPIComm) WriteRegister(register uint8, value uint32, driverAddress uint8) error {
	// Assert the chip select pin (set CS low to start communication)
	csPin, exists := comm.CsPins[driverAddress]
	if !exists {
		return CustomError("Invalid driver address")
	}
	csPin.Low()

	// Set the register address with WRITE_ACCESS (0x80)
	addressWithWriteAccess := register | 0x80

	// Send the address and the data to write (split into 4 bytes)
	_, err := spiTransfer40(comm.spi, addressWithWriteAccess, value)
	if err != nil {
		csPin.High()
		return CustomError("Failed to write register")
	}

	// Deassert the chip select pin (set CS high to end communication)
	csPin.High()

	return nil
}

// ReadRegister sends a register read command to the TMC5160.
func (comm *SPIComm) ReadRegister(register uint8, driverAddress uint8) (uint32, error) {
	// Assert the chip select pin (set CS low to start communication)
	csPin, exists := comm.CsPins[driverAddress]
	if !exists {
		return 0, CustomError("Invalid driver address")
	}
	csPin.Low()

	// Step 1: Send a dummy write operation to begin the read sequence
	_, err := spiTransfer40(comm.spi, register, 0x00) // Send dummy data
	if err != nil {
		csPin.High()
		return 0, CustomError("Failed to send dummy write")
	}
	csPin.High()
	time.Sleep(176 * time.Nanosecond)
	csPin.Low()
	// Step 2: Send the register read request again to get the actual value
	response, err := spiTransfer40(comm.spi, register, 0x00) // Send again to get actual register data
	if err != nil {
		csPin.High()
		return 0, CustomError("Failed to read register")
	}

	// Deassert the chip select pin (set CS high to end communication)
	csPin.High()

	return response, nil
}

func spiTransfer40(spi *machine.SPI, register uint8, txData uint32) (uint32, error) {
	// Prepare the 5-byte buffer for transmission (1 byte address + 4 bytes data)
	tx := []byte{
		register,           // Address byte
		byte(txData >> 24), // Upper 8 bits of data
		byte(txData >> 16), // Middle 8 bits of data
		byte(txData >> 8),  // Next 8 bits of data
		byte(txData),       // Lower 8 bits of data
	}

	rx := make([]byte, 5)

	// Perform the SPI transaction
	err := spi.Tx(tx, rx)
	if err != nil {
		return 0, err
	}

	// Combine the received bytes into a 32-bit response, ignore the address byte
	rxData := uint32(rx[1])<<24 | uint32(rx[2])<<16 | uint32(rx[3])<<8 | uint32(rx[4])

	return rxData, nil
}
