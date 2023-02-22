package bmp180

// Constants/addresses used for I2C.

// The I2C address which this device listens to.
const Address = 0x77

// Registers. Names, addresses and comments copied from the datasheet.
const (
	AC1_MSB          = 0xAA // Calibration coefficients start at 0xAA ends at 0xBF
	CMD_TEMP         = 0x2E
	CMD_PRESSURE     = 0x34
	REG_CTRL         = 0xF4
	REG_TEMP_MSB     = 0xF6
	REG_PRESSURE_MSB = 0xF6

	WHO_AM_I = 0xD0
	CHIP_ID  = 0x55
)

const (
	// ULTRALOWPOWER is the lowest oversampling mode of the pressure measurement.
	ULTRALOWPOWER OversamplingMode = iota
	// BSTANDARD is the standard oversampling mode of the pressure measurement.
	STANDARD
	// HIGHRESOLUTION is a high oversampling mode of the pressure measurement.
	HIGHRESOLUTION
	// ULTRAHIGHRESOLUTION is the highest oversampling mode of the pressure measurement.
	ULTRAHIGHRESOLUTION
)

const (
	SEALEVEL_PRESSURE float32 = 1013.25 // in hPa
)
