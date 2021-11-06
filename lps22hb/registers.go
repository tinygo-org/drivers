package lps22hb

const (

	// I2C address
	LPS22HB_ADDRESS = 0x5C

	// control/status registers
	LPS22HB_WHO_AM_I_REG  = 0x0F
	LPS22HB_CTRL1_REG     = 0x10
	LPS22HB_CTRL2_REG     = 0x11
	LPS22HB_STATUS_REG    = 0x27
	LPS22HB_PRESS_OUT_REG = 0x28
	LPS22HB_TEMP_OUT_REG  = 0x2B
)
