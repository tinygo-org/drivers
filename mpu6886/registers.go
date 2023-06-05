package mpu6886

// Constants/addresses used for I2C.

// The I2C address which this device listens to.
const (
	DefaultAddress   = 0x68
	SecondaryAddress = 0x69
)

// Registers. Names, addresses and comments copied from the datasheet.
const (
	XG_OFFS_TC_H = 0x04
	XG_OFFS_TC_L = 0x05
	YG_OFFS_TC_H = 0x07
	YG_OFFS_TC_L = 0x08
	ZG_OFFS_TC_H = 0x0A
	ZG_OFFS_TC_L = 0x0B

	// Self test registers
	SELF_TEST_X_ACCEL = 0x0D
	SELF_TEST_Y_ACCEL = 0x0E
	SELF_TEST_Z_ACCEL = 0x0F

	XG_OFFS_USRH = 0x13
	XG_OFFS_USRL = 0x14
	YG_OFFS_USRH = 0x15
	YG_OFFS_USRL = 0x16
	ZG_OFFS_USRH = 0x17
	ZG_OFFS_USRL = 0x18

	SMPLRT_DIV      = 0x19
	CONFIG          = 0x1A
	GYRO_CONFIG     = 0x1B
	ACCEL_CONFIG    = 0x1C
	ACCEL_CONFIG_2  = 0x1D
	LP_MODE_CFG     = 0x1E
	ACCEL_WOM_X_THR = 0x20
	ACCEL_WOM_Y_THR = 0x21
	ACCEL_WOM_Z_THR = 0x22
	FIFO_EN         = 0x23
	FSYNC_INT       = 0x36

	// Interrupt configuration
	INT_PIN_CFG        = 0x37
	INT_ENABLE         = 0x38
	FIFO_WM_INT_STATUS = 0x39
	INT_STATUS         = 0x3A

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

	SELF_TEST_X_GYRO = 0x50
	SELF_TEST_Y_GYRO = 0x51
	SELF_TEST_Z_GYRO = 0x52

	E_ID0 = 0x53
	E_ID1 = 0x54
	E_ID2 = 0x55
	E_ID3 = 0x56
	E_ID4 = 0x57
	E_ID5 = 0x58
	E_ID6 = 0x59

	FIFO_WM_TH1       = 0x60
	FIFO_WM_TH2       = 0x61
	SIGNAL_PATH_RESET = 0x68
	ACCEL_INTEL_CTRL  = 0x69
	USER_CTRL         = 0x6A
	PWR_MGMT_1        = 0x6B
	PWR_MGMT_2        = 0x6C
	I2C_IF            = 0x70
	FIFO_COUNTH       = 0x72
	FIFO_COUNTL       = 0x73
	FIFO_R_W          = 0x74
	WHO_AM_I          = 0x75

	XA_OFFSET_H = 0x77
	XA_OFFSET_L = 0x78
	YA_OFFSET_H = 0x7A
	YA_OFFSET_L = 0x7B
	ZA_OFFSET_H = 0x7D
	ZA_OFFSET_L = 0x7E
)

// Accelerometer and gyroscope ranges
const (
	AFS_RANGE_2_G = iota
	AFS_RANGE_4_G
	AFS_RANGE_8_G
	AFS_RANGE_16_G
)

const (
	GFS_RANGE_250 = iota
	GFS_RANGE_500
	GFS_RANGE_1000
	GFS_RANGE_2000
)
