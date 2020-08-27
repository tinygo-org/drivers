package adt7410

// 0x00 Temperature value most significant byte 0x00
// 0x01 Temperature value least significant byte 0x00
// 0x02 Status 0x00
// 0x03 Configuration 0x00
// 0x04 THIGH setpoint most significant byte 0x20 (64°C)
// 0x05 THIGH setpoint least significant byte 0x00 (64°C)
// 0x06 TLOW setpoint most significant byte 0x05 (10°C)
// 0x07 TLOW setpoint least significant byte 0x00 (10°C)
// 0x08 TCRIT setpoint most significant byte 0x49 (147°C)
// 0x09 TCRIT setpoint least significant byte 0x80 (147°C)
// 0x0A THYST setpoint 0x05 (5°C)
// 0x0B ID 0xCX
// 0x0C Reserved 0xXX
// 0x0D Reserved 0xXX
// 0x2E Reserved 0xXX
// 0x2F Software reset 0xXX

const (
	// Address is default I2C address.
	Address = 0x48
	// Address1 is for first device, aka the default.
	Address1 = Address
	// Address2 is for second device.
	Address2 = 0x49
	// Address3 is for third device.
	Address3 = 0x4A
	// Address4 is for fourth device.
	Address4 = 0x4B

	// Temperature Value MSB Register
	RegTempValueMSB = 0x0

	// Temperature Value LSB Register
	RegTempValueLSB = 0x1

	// Status Register
	RegStatus = 0x2

	// Config Register
	RegConfig = 0x3

	// THIGH setpoint most significant byte 0x20 (64°C)
	RegTHIGHMsbReg = 0x4

	// THIGH setpoint least significant byte 0x00 (64°C)
	RegTHIGHLsbReg = 0x5

	// TLOW setpoint most significant byte 0x05 (10°C)
	RegTLOWMsbReg = 0x6

	// TLOW setpoint least significant byte 0x00 (10°C)
	RegTLOWLsbReg = 0x7

	// TCRIT setpoint most significant byte 0x49 (147°C)
	RegTCRITMsbReg = 0x8

	// TCRIT setpoint least significant byte 0x80 (147°C)
	RegTCRITLsbReg = 0x9

	// THYST setpoint 0x05 (5°C)
	RegTHYSTReg = 0xA

	// ID Register (0xCx)
	RegID = 0x0B

	// Software Reset Register
	RegReset = 0x2F
)
