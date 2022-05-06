package lsm6ds3tr

// Constants/addresses used for I2C.

// The I2C address which this device listens to.
const Address = 0x6A

const (
	WHO_AM_I             = 0x0F
	STATUS               = 0x1E
	CTRL1_XL             = 0x10
	CTRL2_G              = 0x11
	CTRL3_C              = 0x12
	CTRL4_C              = 0x13
	CTRL5_C              = 0x14
	CTRL6_C              = 0x15
	CTRL7_G              = 0x16
	CTRL8_XL             = 0x17
	CTRL9_XL             = 0x18
	CTRL10_C             = 0x19
	OUTX_L_G             = 0x22
	OUTX_H_G             = 0x23
	OUTY_L_G             = 0x24
	OUTY_H_G             = 0x25
	OUTZ_L_G             = 0x26
	OUTZ_H_G             = 0x27
	OUTX_L_XL            = 0x28
	OUTX_H_XL            = 0x29
	OUTY_L_XL            = 0x2A
	OUTY_H_XL            = 0x2B
	OUTZ_L_XL            = 0x2C
	OUTZ_H_XL            = 0x2D
	OUT_TEMP_L           = 0x20
	OUT_TEMP_H           = 0x21
	BW_SCAL_ODR_DISABLED = 0x00
	BW_SCAL_ODR_ENABLED  = 0x80
	STEP_TIMESTAMP_L     = 0x49
	STEP_TIMESTAMP_H     = 0x4A
	STEP_COUNTER_L       = 0x4B
	STEP_COUNTER_H       = 0x4C
	STEP_COUNT_DELTA     = 0x15
	TAP_CFG              = 0x58
	INT1_CTRL            = 0x0D

	ACCEL_2G  AccelRange = 0x00
	ACCEL_4G  AccelRange = 0x08
	ACCEL_8G  AccelRange = 0x0C
	ACCEL_16G AccelRange = 0x04

	ACCEL_SR_OFF  AccelSampleRate = 0x00
	ACCEL_SR_13   AccelSampleRate = 0x10
	ACCEL_SR_26   AccelSampleRate = 0x20
	ACCEL_SR_52   AccelSampleRate = 0x30
	ACCEL_SR_104  AccelSampleRate = 0x40
	ACCEL_SR_208  AccelSampleRate = 0x50
	ACCEL_SR_416  AccelSampleRate = 0x60
	ACCEL_SR_833  AccelSampleRate = 0x70
	ACCEL_SR_1666 AccelSampleRate = 0x80
	ACCEL_SR_3332 AccelSampleRate = 0x90
	ACCEL_SR_6664 AccelSampleRate = 0xA0

	GYRO_125DPS  GyroRange = 0x02
	GYRO_245DPS  GyroRange = 0x00
	GYRO_500DPS  GyroRange = 0x04
	GYRO_1000DPS GyroRange = 0x08
	GYRO_2000DPS GyroRange = 0x0C

	GYRO_SR_OFF  GyroSampleRate = 0x00
	GYRO_SR_13   GyroSampleRate = 0x10
	GYRO_SR_26   GyroSampleRate = 0x20
	GYRO_SR_52   GyroSampleRate = 0x30
	GYRO_SR_104  GyroSampleRate = 0x40
	GYRO_SR_208  GyroSampleRate = 0x50
	GYRO_SR_416  GyroSampleRate = 0x60
	GYRO_SR_833  GyroSampleRate = 0x70
	GYRO_SR_1666 GyroSampleRate = 0x80
	GYRO_SR_3332 GyroSampleRate = 0x90
	GYRO_SR_6664 GyroSampleRate = 0xA0
)
