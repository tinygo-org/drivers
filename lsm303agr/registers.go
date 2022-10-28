package lsm303agr

const (

	// Constants/addresses used for I2C.
	ACCEL_ADDRESS = 0x19
	MAG_ADDRESS   = 0x1E

	// i2C 8-bit subaddress (SUB): the 7 LSb represent the actual register address
	// while the MSB enables address auto increment.
	// If the MSb of the SUB field is 1, the SUB (register address) is
	// automatically increased to allow multiple data read/writes.
	ADDR_AUTO_INC_MASK = 0x80

	// accelerometer registers.
	ACCEL_WHO_AM_I     = 0x0F
	ACCEL_CTRL_REG1_A  = 0x20
	ACCEL_CTRL_REG4_A  = 0x23
	ACCEL_OUT_X_L_A    = 0x28
	ACCEL_OUT_X_H_A    = 0x29
	ACCEL_OUT_Y_L_A    = 0x2A
	ACCEL_OUT_Y_H_A    = 0x2B
	ACCEL_OUT_Z_L_A    = 0x2C
	ACCEL_OUT_Z_H_A    = 0x2D
	ACCEL_OUT_AUTO_INC = ACCEL_OUT_X_L_A | ADDR_AUTO_INC_MASK

	// magnetic sensor registers.
	MAG_WHO_AM_I     = 0x4F
	MAG_MR_REG_M     = 0x60
	MAG_OUT_X_L_M    = 0x68
	MAG_OUT_X_H_M    = 0x69
	MAG_OUT_Y_L_M    = 0x6A
	MAG_OUT_Y_H_M    = 0x6B
	MAG_OUT_Z_L_M    = 0x6C
	MAG_OUT_Z_H_M    = 0x6D
	MAG_OUT_AUTO_INC = MAG_OUT_X_L_M | ADDR_AUTO_INC_MASK

	// temperature sensor registers.
	TEMP_CFG_REG_A    = 0x1F
	OUT_TEMP_L_A      = 0x0C
	OUT_TEMP_H_A      = 0x0D
	OUT_TEMP_AUTO_INC = OUT_TEMP_L_A | ADDR_AUTO_INC_MASK

	// accelerometer power mode.
	ACCEL_POWER_NORMAL = 0x00 // default
	ACCEL_POWER_LOW    = 0x08

	// accelerometer range.
	ACCEL_RANGE_2G  = 0x00 // default
	ACCEL_RANGE_4G  = 0x01
	ACCEL_RANGE_8G  = 0x02
	ACCEL_RANGE_16G = 0x03

	// accelerometer data rate.
	ACCEL_DATARATE_1HZ    = 0x01
	ACCEL_DATARATE_10HZ   = 0x02
	ACCEL_DATARATE_25HZ   = 0x03
	ACCEL_DATARATE_50HZ   = 0x04
	ACCEL_DATARATE_100HZ  = 0x05 // default
	ACCEL_DATARATE_200HZ  = 0x06
	ACCEL_DATARATE_400HZ  = 0x07
	ACCEL_DATARATE_1344HZ = 0x09 // 5376Hz in low-power mode

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
