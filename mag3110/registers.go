package mag3110

// Constants/addresses used for I2C.

// The I2C address which this device listens to.
const Address = 0x0E

// Registers. Names, addresses and comments are copied from the datasheet.
const (
	DR_STATUS = 0x00 // Data ready status per axis
	OUT_X_MSB = 0x01 // Bits [15:8] of X measurement
	OUT_X_LSB = 0x02 // Bits [7:0] of X measurement
	OUT_Y_MSB = 0x03 // Bits [15:8] of Y measurement
	OUT_Y_LSB = 0x04 // Bits [7:0] of Y measurement
	OUT_Z_MSB = 0x05 // Bits [15:8] of Z measurement
	OUT_Z_LSB = 0x06 // Bits [7:0] of Z measurement
	WHO_AM_I  = 0x07 // Device ID Number
	SYSMOD    = 0x08 // Current System Mode
	OFF_X_MSB = 0x09 // Bits [14:7] of user X offset
	OFF_X_LSB = 0x0A // Bits [6:0] of user X offset
	OFF_Y_MSB = 0x0B // Bits [14:7] of user Y offset
	OFF_Y_LSB = 0x0C // Bits [6:0] of user Y offset
	OFF_Z_MSB = 0x0D // Bits [14:7] of user Z offset
	OFF_Z_LSB = 0x0E // Bits [6:0] of user Z offset
	DIE_TEMP  = 0x0F // Temperature, signed 8 bits in °C
	CTRL_REG1 = 0x10 // Operation modes
	CTRL_REG2 = 0x11 // Operation modes
)
