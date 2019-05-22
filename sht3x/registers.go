package sht3x

// Constants/addresses used for I2C.

// The I2C address which this device listens to.
const (
	AddressA = 0x44
	AddressB = 0x45
)

const (
	// single shot, high repeatability
	MEASUREMENT_COMMAND_MSB = 0x24
	MEASUREMENT_COMMAND_LSB = 0x00
)
