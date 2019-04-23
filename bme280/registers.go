package bme280

// Constants/addresses used for I2C.

// The I2C address which this device listens to.
const Address = 0x76

// Registers. Names, addresses and comments copied from the datasheet.
const (
	CTRL_MEAS_ADDR     = 0xF4
	CTRL_HUMIDITY_ADDR = 0xF2
	CTRL_CONFIG        = 0xF5
	REG_PRESSURE       = 0xF7
	REG_TEMPERATURE    = 0xFA
	REG_HUMIDITY       = 0xFD
	REG_CALIBRATION    = 0x88
	CMD_RESET          = 0xE0

	WHO_AM_I = 0xD0
	CHIP_ID  = 0x60
)
