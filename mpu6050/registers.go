package mpu6050

// read reads register reg and writes n bytes to b where
// n is the length of b.
func (p *Device) read(reg uint8, buff []byte) error {
	buf := [1]byte{reg}
	return p.conn.Tx(uint16(p.address), buf[:1], buff)
}

// write8 writes a registry byte
func (p *Device) write8(reg uint8, datum byte) error {
	var buff [2]byte
	buff[0] = reg
	buff[1] = datum
	return p.conn.Tx(uint16(p.address), buff[:], nil)
}

// MPU 6050 REGISTER ADRESSES
const (
	_SMPRT_DIV    uint8 = 0x19
	_CONFIG       uint8 = 0x1A
	_GYRO_CONFIG  uint8 = 0x1B
	_ACCEL_CONFIG uint8 = 0x1C
	_FIFO_EN      uint8 = 0x23
)

// MPU 6050 MASKS
const (
	_G_FS_SEL     uint8 = 0x18
	_AFS_SEL      uint8 = 0x18
	_CLK_SEL_MASK uint8 = 0x07
	_SLEEP_MASK   uint8 = 0x40
)

// MPU 6050 SHIFTS
const (
	_AFS_SHIFT   uint8 = 3
	_G_FS_SHIFT  uint8 = 3
	_SLEEP_SHIFT uint8 = 6
)

// Gyroscope ranges for Init configuration
const (
	// 250°/s
	RangeGyro250 RangeGyro = iota
	// 500°/s
	RangeGyro500
	// 1000°/s
	RangeGyro1000
	// 2000°/s
	RangeGyro2000
)

// Accelerometer ranges for Init configuration
const (
	// 2g
	RangeAccel2 RangeAccel = iota
	// 4g
	RangeAccel4
	// 8g
	RangeAccel8
	// 16g
	RangeAccel16
)

// Registers. Names, addresses and comments copied from the datasheet.
const (
	// Self test registers
	_SELF_TEST_X = 0x0D
	_SELF_TEST_Y = 0x0E
	_SELF_TEST_Z = 0x0F
	_SELF_TEST_A = 0x10

	_SMPLRT_DIV = 0x19 // Sample rate divider

	// I2C pass-through configuration
	_I2C_MST_CTRL   = 0x24
	_I2C_SLV0_ADDR  = 0x25
	_I2C_SLV0_REG   = 0x26
	_I2C_SLV0_CTRL  = 0x27
	_I2C_SLV1_ADDR  = 0x28
	_I2C_SLV1_REG   = 0x29
	_I2C_SLV1_CTRL  = 0x2A
	_I2C_SLV2_ADDR  = 0x2B
	_I2C_SLV2_REG   = 0x2C
	_I2C_SLV2_CTRL  = 0x2D
	_I2C_SLV3_ADDR  = 0x2E
	_I2C_SLV3_REG   = 0x2F
	_I2C_SLV3_CTRL  = 0x30
	_I2C_SLV4_ADDR  = 0x31
	_I2C_SLV4_REG   = 0x32
	_I2C_SLV4_DO    = 0x33
	_I2C_SLV4_CTRL  = 0x34
	_I2C_SLV4_DI    = 0x35
	_I2C_MST_STATUS = 0x36

	// Interrupt configuration
	_INT_PIN_CFG = 0x37 // Interrupt pin/bypass enable configuration
	_INT_ENABLE  = 0x38 // Interrupt enable
	_INT_STATUS  = 0x3A // Interrupt status

	// Accelerometer measurements
	_ACCEL_XOUT_H = 0x3B
	_ACCEL_XOUT_L = 0x3C
	_ACCEL_YOUT_H = 0x3D
	_ACCEL_YOUT_L = 0x3E
	_ACCEL_ZOUT_H = 0x3F
	_ACCEL_ZOUT_L = 0x40

	// Temperature measurement
	_TEMP_OUT_H = 0x41
	_TEMP_OUT_L = 0x42

	// Gyroscope measurements
	_GYRO_XOUT_H = 0x43
	_GYRO_XOUT_L = 0x44
	_GYRO_YOUT_H = 0x45
	_GYRO_YOUT_L = 0x46
	_GYRO_ZOUT_H = 0x47
	_GYRO_ZOUT_L = 0x48

	// External sensor data
	_EXT_SENS_DATA_00 = 0x49
	_EXT_SENS_DATA_01 = 0x4A
	_EXT_SENS_DATA_02 = 0x4B
	_EXT_SENS_DATA_03 = 0x4C
	_EXT_SENS_DATA_04 = 0x4D
	_EXT_SENS_DATA_05 = 0x4E
	_EXT_SENS_DATA_06 = 0x4F
	_EXT_SENS_DATA_07 = 0x50
	_EXT_SENS_DATA_08 = 0x51
	_EXT_SENS_DATA_09 = 0x52
	_EXT_SENS_DATA_10 = 0x53
	_EXT_SENS_DATA_11 = 0x54
	_EXT_SENS_DATA_12 = 0x55
	_EXT_SENS_DATA_13 = 0x56
	_EXT_SENS_DATA_14 = 0x57
	_EXT_SENS_DATA_15 = 0x58
	_EXT_SENS_DATA_16 = 0x59
	_EXT_SENS_DATA_17 = 0x5A
	_EXT_SENS_DATA_18 = 0x5B
	_EXT_SENS_DATA_19 = 0x5C
	_EXT_SENS_DATA_20 = 0x5D
	_EXT_SENS_DATA_21 = 0x5E
	_EXT_SENS_DATA_22 = 0x5F
	_EXT_SENS_DATA_23 = 0x60

	// I2C peripheral data out
	_I2C_PER0_DO      = 0x63
	_I2C_PER1_DO      = 0x64
	_I2C_PER2_DO      = 0x65
	_I2C_PER3_DO      = 0x66
	_I2C_MST_DELAY_CT = 0x67

	// Clock settings
	_CLOCK_INTERNAL               = 0x00
	_CLOCK_PLL_XGYRO              = 0x01
	_CLOCK_PLL_YGYRO              = 0x02
	_CLOCK_PLL_ZGYRO              = 0x03
	_CLOCK_PLL_EXTERNAL_32_768_KZ = 0x04
	_CLOCK_PLL_EXTERNAL_19_2_MHZ  = 0x05
	_CLOCK_RESERVED               = 0x06
	_CLOCK_STOP                   = 0x07

	// Accelerometer settings
	_AFS_RANGE_2G  = 0x00
	_AFS_RANGE_4G  = 0x01
	_AFS_RANGE_8G  = 0x02
	_AFS_RANGE_16G = 0x03

	// Gyroscope settings
	_FS_RANGE_250  = 0x00
	_FS_RANGE_500  = 0x01
	_FS_RANGE_1000 = 0x02
	_FS_RANGE_2000 = 0x03

	// other registers
	_SIGNAL_PATH_RES = 0x68 // Signal path reset
	_USER_CTRL       = 0x6A // User control
	_PWR_MGMT_1      = 0x6B // Power Management 1
	_PWR_MGMT_2      = 0x6C // Power Management 2
	_FIFO_COUNTH     = 0x72 // FIFO count registers (high bits)
	_FIFO_COUNTL     = 0x73 // FIFO count registers (low bits)
	_FIFO_R_W        = 0x74 // FIFO read/write
	_WHO_AM_I        = 0x75 // Who am I register
)
