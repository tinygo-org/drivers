package mpu6050

// Constants/addresses used for I2C.

// The I2C address which this device listens to.
const Address = 0x68

// Registers. Names, addresses and comments copied from the datasheet.
const (
	// Self test registers
	SELF_TEST_X = 0x0D
	SELF_TEST_Y = 0x0E
	SELF_TEST_Z = 0x0F
	SELF_TEST_A = 0x10

	SMPLRT_DIV   = 0x19 // Sample rate divider
	CONFIG       = 0x1A // Configuration
	GYRO_CONFIG  = 0x1B // Gyroscope configuration
	ACCEL_CONFIG = 0x1C // Accelerometer configuration
	FIFO_EN      = 0x23 // FIFO enable

	// I2C pass-through configuration
	I2C_MST_CTRL   = 0x24
	I2C_SLV0_ADDR  = 0x25
	I2C_SLV0_REG   = 0x26
	I2C_SLV0_CTRL  = 0x27
	I2C_SLV1_ADDR  = 0x28
	I2C_SLV1_REG   = 0x29
	I2C_SLV1_CTRL  = 0x2A
	I2C_SLV2_ADDR  = 0x2B
	I2C_SLV2_REG   = 0x2C
	I2C_SLV2_CTRL  = 0x2D
	I2C_SLV3_ADDR  = 0x2E
	I2C_SLV3_REG   = 0x2F
	I2C_SLV3_CTRL  = 0x30
	I2C_SLV4_ADDR  = 0x31
	I2C_SLV4_REG   = 0x32
	I2C_SLV4_DO    = 0x33
	I2C_SLV4_CTRL  = 0x34
	I2C_SLV4_DI    = 0x35
	I2C_MST_STATUS = 0x36

	// Interrupt configuration
	INT_PIN_CFG = 0x37 // Interrupt pin/bypass enable configuration
	INT_ENABLE  = 0x38 // Interrupt enable
	INT_STATUS  = 0x3A // Interrupt status

	// Accelerometer measurements
	ACCEL_XOUT_H = 0x3B
	ACCEL_XOUT_L = 0x3C
	ACCEL_YOUT_H = 0x3D
	ACCEL_YOUT_L = 0x3E
	ACCEL_ZOUT_H = 0x3F
	ACCEL_ZOUT_L = 0x40

	// Temperature measurement
	TEMP_OUT_H = 0x41
	TEMP_OUT_L = 0x42

	// Gyroscope measurements
	GYRO_XOUT_H = 0x43
	GYRO_XOUT_L = 0x44
	GYRO_YOUT_H = 0x45
	GYRO_YOUT_L = 0x46
	GYRO_ZOUT_H = 0x47
	GYRO_ZOUT_L = 0x48

	// External sensor data
	EXT_SENS_DATA_00 = 0x49
	EXT_SENS_DATA_01 = 0x4A
	EXT_SENS_DATA_02 = 0x4B
	EXT_SENS_DATA_03 = 0x4C
	EXT_SENS_DATA_04 = 0x4D
	EXT_SENS_DATA_05 = 0x4E
	EXT_SENS_DATA_06 = 0x4F
	EXT_SENS_DATA_07 = 0x50
	EXT_SENS_DATA_08 = 0x51
	EXT_SENS_DATA_09 = 0x52
	EXT_SENS_DATA_10 = 0x53
	EXT_SENS_DATA_11 = 0x54
	EXT_SENS_DATA_12 = 0x55
	EXT_SENS_DATA_13 = 0x56
	EXT_SENS_DATA_14 = 0x57
	EXT_SENS_DATA_15 = 0x58
	EXT_SENS_DATA_16 = 0x59
	EXT_SENS_DATA_17 = 0x5A
	EXT_SENS_DATA_18 = 0x5B
	EXT_SENS_DATA_19 = 0x5C
	EXT_SENS_DATA_20 = 0x5D
	EXT_SENS_DATA_21 = 0x5E
	EXT_SENS_DATA_22 = 0x5F
	EXT_SENS_DATA_23 = 0x60

	// I2C peripheral data out
	I2C_PER0_DO      = 0x63
	I2C_PER1_DO      = 0x64
	I2C_PER2_DO      = 0x65
	I2C_PER3_DO      = 0x66
	I2C_MST_DELAY_CT = 0x67

	// Clock settings
	CLOCK_INTERNAL               = 0x00
	CLOCK_PLL_XGYRO              = 0x01
	CLOCK_PLL_YGYRO              = 0x02
	CLOCK_PLL_ZGYRO              = 0x03
	CLOCK_PLL_EXTERNAL_32_768_KZ = 0x04
	CLOCK_PLL_EXTERNAL_19_2_MHZ  = 0x05
	CLOCK_RESERVED               = 0x06
	CLOCK_STOP                   = 0x07

	// Accelerometer settings
	AFS_RANGE_2G  = 0x00
	AFS_RANGE_4G  = 0x01
	AFS_RANGE_8G  = 0x02
	AFS_RANGE_16G = 0x03

	// Gyroscope settings
	FS_RANGE_250  = 0x00
	FS_RANGE_500  = 0x01
	FS_RANGE_1000 = 0x02
	FS_RANGE_2000 = 0x03

	// other registers
	SIGNAL_PATH_RES = 0x68 // Signal path reset
	USER_CTRL       = 0x6A // User control
	PWR_MGMT_1      = 0x6B // Power Management 1
	PWR_MGMT_2      = 0x6C // Power Management 2
	FIFO_COUNTH     = 0x72 // FIFO count registers (high bits)
	FIFO_COUNTL     = 0x73 // FIFO count registers (low bits)
	FIFO_R_W        = 0x74 // FIFO read/write
	WHO_AM_I        = 0x75 // Who am I
)
