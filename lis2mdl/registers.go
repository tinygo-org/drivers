package lis2mdl

const (
	// Constants/addresses used for I2C.
	ADDRESS = 0x1E

	// magnetic sensor registers.
	OFFSET_X_REG_L = 0x45
	OFFSET_X_REG_H = 0x46
	OFFSET_Y_REG_L = 0x47
	OFFSET_Y_REG_H = 0x48
	OFFSET_Z_REG_L = 0x49
	OFFSET_Z_REG_H = 0x4A
	WHO_AM_I       = 0x4F
	CFG_REG_A      = 0x60
	CFG_REG_B      = 0x61
	CFG_REG_C      = 0x62
	INT_CRTL_REG   = 0x63
	INT_SOURCE_REG = 0x64
	INT_THS_L_REG  = 0x65
	INT_THS_H_REG  = 0x66
	STATUS_REG     = 0x67
	OUTX_L_REG     = 0x68
	OUTX_H_REG     = 0x69
	OUTY_L_REG     = 0x6A
	OUTY_H_REG     = 0x6B
	OUTZ_L_REG     = 0x6C
	OUTZ_H_REG     = 0x6D
	TEMP_OUT_L_REG = 0x6E
	TEMP_OUT_H_REG = 0x6F

	// magnetic sensor power mode.
	POWER_NORMAL = 0x00 // default
	POWER_LOW    = 0x01

	// magnetic sensor operate mode.
	SYSTEM_CONTINUOUS = 0x00 // default
	SYSTEM_SINGLE     = 0x01

	// magnetic sensor data rate
	DATARATE_10HZ  = 0x00 // default
	DATARATE_20HZ  = 0x01
	DATARATE_50HZ  = 0x02
	DATARATE_100HZ = 0x03
)
