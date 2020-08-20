package lis2mdl

const (
	// Constants/addresses used for I2C.
	MAG_ADDRESS = 0x1E

	// magnetic sensor registers.
	MAG_WHO_AM_I     = 0x4F
	MAG_MR_CFG_REG_A = 0x60
	MAG_MR_CFG_REG_B = 0x61
	MAG_MR_CFG_REG_C = 0x62
	MAG_OUT_X_L_M    = 0x68
	MAG_OUT_X_H_M    = 0x69
	MAG_OUT_Y_L_M    = 0x6A
	MAG_OUT_Y_H_M    = 0x6B
	MAG_OUT_Z_L_M    = 0x6C
	MAG_OUT_Z_H_M    = 0x6D

	// magnetic sensor power mode.
	MAG_POWER_NORMAL = 0x00 // default
	MAG_POWER_LOW    = 0x01

	// magnetic sensor operate mode.
	MAG_SYSTEM_CONTINUOUS = 0x00 // default
	MAG_SYSTEM_SINGLE     = 0x01

	// magnetic sensor data rate
	MAG_DATARATE_10HZ  = 0x00 // default
	MAG_DATARATE_20HZ  = 0x01
	MAG_DATARATE_50HZ  = 0x02
	MAG_DATARATE_100HZ = 0x03
)
