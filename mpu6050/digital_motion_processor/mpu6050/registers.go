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

	// other data values (bits)
	BANK_SEL              = 0x6D // DMP bank selection
	MEM_START_ADDR        = 0x6E // Register -
	MEM_R_W               = 0x6F // Register - Memory
	TC_OTP_BNK_VLD_BIT    = 0x00 // bit 0
	XG_OFFS_TC            = 0x00 // Register - [7] PWR_MODE, [6:1] XG_OFFS_TC, [0] OTP_BNK_VLD
	DMP_CODE_SIZE         = 1929
	DMP_FIFO_RATE_DIVISOR = 0x01
	DMP_MEMORY_CHUNK_SIZE = 16
	WIRE_BUFFER_LENGTH    = 32

	FIFO_DEFAULT_TIMEOUT = 11000

	PWR_MGMT_1_DEVICE_RESET_BIT = 0x07
	PWR_MGMT_1_SLEEP_BIT        = 0x06
	PWR_MGMT_1_CYCLE_BIT        = 0x05
	PWR_MGMT_1_TEMP_DIS_BIT     = 0x03
	PWR_MGMT_1_CLKSEL_BIT       = 0x02
	PWR_MGMT_1_CLKSEL_LENGTH    = 0x03

	USER_CTRL_DMP_EN_BIT         = 0x07
	USER_CTRL_FIFO_EN_BIT        = 0x06
	USER_CTRL_I2C_MST_EN_BIT     = 0x05
	USER_CTRL_I2C_IF_DIS_BIT     = 0x04
	USER_CTRL_DMP_RESET_BIT      = 0x03
	USER_CTRL_FIFO_RESET_BIT     = 0x02
	USER_CTRL_I2C_MST_RESET_BIT  = 0x01
	USER_CTRL_SIG_COND_RESET_BIT = 0x00

	CONFIG_EXT_SYNC_SET_BIT    = 0x05
	CONFIG_EXT_SYNC_SET_LENGTH = 0x03
	CONFIG_DLPF_CFG_BIT        = 0x02
	CONFIG_DLPF_CFG_LENGTH     = 0x03

	GYRO_CONFIG_FS_SEL_BIT    = 0x04
	GYRO_CONFIG_FS_SEL_LENGTH = 0x02

	SYNC_DISABLED     = 0x0
	SYNC_TEMP_OUT_L   = 0x1
	SYNC_GYRO_XOUT_L  = 0x2
	SYNC_GYRO_YOUT_L  = 0x3
	SYNC_GYRO_ZOUT_L  = 0x4
	SYNC_ACCEL_XOUT_L = 0x5
	SYNC_ACCEL_YOUT_L = 0x6
	SYNC_ACCEL_ZOUT_L = 0x7

	DLPF_BW_256 = 0x00
	DLPF_BW_188 = 0x01
	DLPF_BW_98  = 0x02
	DLPF_BW_42  = 0x03
	DLPF_BW_20  = 0x04
	DLPF_BW_10  = 0x05
	DLPF_BW_5   = 0x06

	GYRO_FS_250  = 0x00
	GYRO_FS_500  = 0x01
	GYRO_FS_1000 = 0x02
	GYRO_FS_2000 = 0x03

	DMP_CFG_1 = 0x70
	DMP_CFG_2 = 0x71

	MOT_THR   = 0x1F
	MOT_DUR   = 0x20
	ZRMOT_THR = 0x21
	ZRMOT_DUR = 0x22

	INTERRUPT_FF_BIT          = 0x07
	INTERRUPT_MOT_BIT         = 0x06
	INTERRUPT_ZMOT_BIT        = 0x05
	INTERRUPT_FIFO_OFLOW_BIT  = 0x04
	INTERRUPT_I2C_MST_INT_BIT = 0x03
	INTERRUPT_PLL_RDY_INT_BIT = 0x02
	INTERRUPT_DMP_INT_BIT     = 0x01
	INTERRUPT_DATA_RDY_BIT    = 0x00

	ACCEL_CONFIG_XA_ST_BIT        = 0x07
	ACCEL_CONFIG_YA_ST_BIT        = 0x06
	ACCEL_CONFIG_ZA_ST_BIT        = 0x05
	ACCEL_CONFIG_AFS_SEL_BIT      = 0x04
	ACCEL_CONFIG_AFS_SEL_LENGTH   = 0x02
	ACCEL_CONFIG_ACCEL_HPF_BIT    = 0x02
	ACCEL_CONFIG_ACCEL_HPF_LENGTH = 0x03

	WHO_AM_I_BIT    = 0x06
	WHO_AM_I_LENGTH = 0x06

	XG_OFFS_USRH = 0x13 //[15:0] XG_OFFS_USR
	XG_OFFS_USRL = 0x14
	YG_OFFS_USRH = 0x15 //[15:0] YG_OFFS_USR
	YG_OFFS_USRL = 0x16
	ZG_OFFS_USRH = 0x17 //[15:0] ZG_OFFS_USR
	ZG_OFFS_USRL = 0x18

	XA_OFFS_H    = 0x06 //[15:0] XA_OFFS
	XA_OFFS_L_TC = 0x07
	YA_OFFS_H    = 0x08 //[15:0] YA_OFFS
	YA_OFFS_L_TC = 0x09
	ZA_OFFS_H    = 0x0A //[15:0] ZA_OFFS
	ZA_OFFS_L_TC = 0x0B
)
