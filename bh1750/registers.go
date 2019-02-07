package bh1750

// Constants/addresses used for I2C.

// The I2C address which this device listens to.
const Address = 0x23

// Registers. Names, addresses and comments copied from the datasheet.
const (
	POWER_DOWN                              = 0x00
	POWER_ON                                = 0x01
	RESET                                   = 0x07
	CONTINUOUS_HIGH_RES_MODE   SamplingMode = 0x10
	CONTINUOUS_HIGH_RES_MODE_2 SamplingMode = 0x11
	CONTINUOUS_LOW_RES_MODE    SamplingMode = 0x13
	ONE_TIME_HIGH_RES_MODE     SamplingMode = 0x20
	ONE_TIME_HIGH_RES_MODE_2   SamplingMode = 0x21
	ONE_TIME_LOW_RES_MODE      SamplingMode = 0x23

	// resolution in 10*lx
	HIGH_RES  = 10
	HIGH_RES2 = 5
	LOW_RES   = 40
)
