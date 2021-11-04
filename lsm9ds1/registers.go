package lsm9ds1

// Constants/addresses used for I2C.

const (

	// Constants/addresses used for I2C.
	ACCEL_ADDRESS = 0x6B
	MAG_ADDRESS   = 0x1E

	// Table 21. Accelerometer and gyroscope register address map
	WHO_AM_I     = 0x0F // value 0x68
	CTRL_REG1_G  = 0x10
	OUT_X_L_G    = 0x18
	OUT_X_H_G    = 0x19
	OUT_Y_L_G    = 0x1A
	OUT_Y_H_G    = 0x1B
	OUT_Z_L_G    = 0x1C
	OUT_Z_H_G    = 0x1D
	OUT_TEMP_L   = 0x15
	OUT_TEMP_H   = 0x16
	CTRL_REG6_XL = 0x20
	STATUS_REG   = 0x27
	OUT_X_L_XL   = 0x28
	OUT_X_H_XL   = 0x29
	OUT_Y_L_XL   = 0x2A
	OUT_Y_H_XL   = 0x2B
	OUT_Z_L_XL   = 0x2C
	OUT_Z_H_XL   = 0x2D

	// Table 22. Magnetic sensor register address map
	OFFSET_X_REG_L_M = 0x05
	OFFSET_X_REG_H_M = 0x06
	OFFSET_Y_REG_L_M = 0x07
	OFFSET_Y_REG_H_M = 0x08
	OFFSET_Z_REG_L_M = 0x09
	OFFSET_Z_REG_H_M = 0x0A
	WHO_AM_I_M       = 0x0F // value 0x3D
	CTRL_REG1_M      = 0x20 // TEMP_COMP OM1 OM0 DO2 DO1 DO0 FAST_ODR ST
	CTRL_REG2_M      = 0x21 // 0 FS1 FS0 0 REBOOT SOFT_RST 0 0
	CTRL_REG3_M      = 0x22 // 0 LP 0 0 SIM MD1 MD0
	CTRL_REG4_M      = 0x23 // 0 0 0 0 OMZ1 OMZ0 BLE 0
	STATUS_REG_M     = 0x27
	OUT_X_L_M        = 0x28
	OUT_X_H_M        = 0x29
	OUT_Y_L_M        = 0x2A
	OUT_Y_H_M        = 0x2B
	OUT_Z_L_M        = 0x2C
	OUT_Z_H_M        = 0x2D

	// Table 67. CTRL_REG6_XL register description
	ACCEL_2G  AccelRange = 0b00
	ACCEL_4G  AccelRange = 0b10
	ACCEL_8G  AccelRange = 0b11
	ACCEL_16G AccelRange = 0b01

	// Table 68. ODR register setting (accelerometer only mode)
	ACCEL_SR_OFF AccelSampleRate = 0b000
	ACCEL_SR_10  AccelSampleRate = 0b001
	ACCEL_SR_50  AccelSampleRate = 0b010
	ACCEL_SR_119 AccelSampleRate = 0b011
	ACCEL_SR_238 AccelSampleRate = 0b100
	ACCEL_SR_476 AccelSampleRate = 0b101
	ACCEL_SR_952 AccelSampleRate = 0b110

	// Table 67. CTRL_REG6_XL register description
	ACCEL_BW_50  AccelBandwidth = 0b11
	ACCEL_BW_105 AccelBandwidth = 0b10
	ACCEL_BW_211 AccelBandwidth = 0b01
	ACCEL_BW_408 AccelBandwidth = 0b00

	// Table 45. CTRL_REG1_G register description
	GYRO_250DPS  GyroRange = 0b00
	GYRO_500DPS  GyroRange = 0b01
	GYRO_2000DPS GyroRange = 0b11

	// Table 9. Gyroscope operating modes
	// Table 46. ODR and BW configuration setting (after LPF1)
	GYRO_SR_OFF GyroSampleRate = 0b000
	GYRO_SR_15  GyroSampleRate = 0b001
	GYRO_SR_60  GyroSampleRate = 0b010
	GYRO_SR_119 GyroSampleRate = 0b011
	GYRO_SR_238 GyroSampleRate = 0b100
	GYRO_SR_476 GyroSampleRate = 0b101
	GYRO_SR_952 GyroSampleRate = 0b110

	// Table 114. Full-scale selection
	MAG_4G  MagRange = 0b00
	MAG_8G  MagRange = 0b01
	MAG_12G MagRange = 0b10
	MAG_16G MagRange = 0b11

	// Table 111. Output data rate configuration
	MAG_SR_06 MagSampleRate = 0b000
	MAG_SR_1  MagSampleRate = 0b001
	MAG_SR_2  MagSampleRate = 0b010
	MAG_SR_5  MagSampleRate = 0b011
	MAG_SR_10 MagSampleRate = 0b100
	MAG_SR_20 MagSampleRate = 0b101
	MAG_SR_40 MagSampleRate = 0b110
	MAG_SR_80 MagSampleRate = 0b111
)
