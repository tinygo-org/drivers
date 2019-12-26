package adt7410

const (
	// Default I2C address
	Address = 0x48

	// Temperature Value MSB Register
	RegTempValueMSB = 0x0

	// Temperature Value LSB Register
	RegTempValueLSB = 0x1

	// Status Register
	RegStatus = 0x2

	// Config Register
	RegConfig = 0x3

	// ID Register
	RegID = 0x0B

	// Software Reset Register
	RegReset = 0x2F
)
