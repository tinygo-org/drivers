package shtc3

// Constants used for I2C.

const (
	SHTC3_ADDRESS        = 0x70
	SHTC3_CMD_WAKEUP     = "\x35\x17" // Wake up
	SHTC3_CMD_MEASURE_HP = "\x7C\xA2" // Read sensor in high power mode with clock stretching
	SHTC3_CMD_SLEEP      = "\xB0\x98" // Sleep
	SHTC3_CMD_SOFT_RESET = "\x80\x5D" // Soft Reset
)
