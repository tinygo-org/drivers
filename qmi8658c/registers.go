package qmi8656c

// The I2C address that the sensor listens to.
const Address = 0x6B

const (
	// Who am I
	WHO_AM_I   = 0x00
	IDENTIFIER = 0x05

	// Configuration registers
	CTRL1 = 0x02 // SPI Modes
	CTRL2 = 0x03 // Accelerometer config
	CTRL3 = 0x04 // Gyro config
	CTRL4 = 0x05 // Magnetometer config (ignored)
	CTRL5 = 0x06 // Sensor DSP config
	CTRL6 = 0x07 // Motion on Demand (ignored)
	CTRL7 = 0x08 // Sensors config

	// Interface config (CTRL1)
	SPI_4_WIRE        = 0x00
	SPI_3_WIRE        = 0x80
	SPI_NOT_AUTO_INC  = 0x00
	SPI_AUTO_INC      = 0x40
	SPI_LITTLE_ENDIAN = 0x00
	SPI_BIG_ENDIAN    = 0x20

	// Accelerometer scale config (CTRL2-H)
	ACC_SELF_TEST = 0x80

	// Accelerometer scale config (CTRL2-H)
	ACC_2G  = 0x00
	ACC_4G  = 0x10
	ACC_8G  = 0x20
	ACC_16G = 0x30

	// Accelerometer output data rate (ODR) config (CTRL2-L)
	ACC_NORMAL_8000HZ   = 0x00
	ACC_NORMAL_4000HZ   = 0x01
	ACC_NORMAL_2000HZ   = 0x02
	ACC_NORMAL_1000HZ   = 0x03
	ACC_NORMAL_500HZ    = 0x04
	ACC_NORMAL_250HZ    = 0x05
	ACC_NORMAL_125HZ    = 0x06
	ACC_NORMAL_62HZ     = 0x07
	ACC_NORMAL_31HZ     = 0x08
	ACC_LOW_POWER_128HZ = 0x0C
	ACC_LOW_POWER_21HZ  = 0x0D
	ACC_LOW_POWER_11HZ  = 0x0E
	ACC_LOW_POWER_3HZ   = 0x0F

	// Gyro scale config (CTRL3-H)
	GYRO_SELF_TEST = 0x80

	// Gyro scale config (CTRL3-H)
	GYRO_16DPS   = 0x00
	GYRO_32DPS   = 0x10
	GYRO_64DPS   = 0x20
	GYRO_128DPS  = 0x30
	GYRO_256DPS  = 0x40
	GYRO_512DPS  = 0x50
	GYRO_1024DPS = 0x60
	GYRO_2048DPS = 0x70

	// Gyro output data rate (ODR) config (CTRL3-L)
	GYRO_8000HZ = 0x00
	GYRO_4000HZ = 0x01
	GYRO_2000HZ = 0x02
	GYRO_1000HZ = 0x03
	GYRO_500HZ  = 0x04
	GYRO_250HZ  = 0x05
	GYRO_125HZ  = 0x06
	GYRO_62HZ   = 0x07
	GYRO_31HZ   = 0x08

	// Gyro DSP config (CTRL4-H)
	GYRO_LOW_PASS_OFF  = 0x00 // Disabled
	GYRO_LOW_PASS_2_62 = 0x10 // 2.62% of output data rate (ODR)
	GYRO_LOW_PASS_3_59 = 0x30 // 3.59% of output data rate (ODR)
	GYRO_LOW_PASS_5_32 = 0x50 // 5.32% of output data rate (ODR)
	GYRO_LOW_PASS_14   = 0x70 // 14% of output data rate (ODR)

	// Accelerometer DSP config (CTRL4-L)
	ACC_LOW_PASS_OFF  = 0x00 // Disabled
	ACC_LOW_PASS_2_62 = 0x01 // 2.62% of output data rate (ODR)
	ACC_LOW_PASS_3_59 = 0x03 // 3.59% of output data rate (ODR)
	ACC_LOW_PASS_5_32 = 0x05 // 5.32% of output data rate (ODR)
	ACC_LOW_PASS_14   = 0x07 // 14% of output data rate (ODR)

	// Motion on demand (MOD) (CTRL6)
	MOD_DISABLE = 0x00
	MOD_ENABLE  = 0x80

	// Enable sensors (CTRL7)
	GYRO_DISABLE       = 0x00
	GYRO_FULL_ENABLE   = 0x02
	GYRO_SNOOZE_ENABLE = 0x12
	ACC_DISABLE        = 0x00
	ACC_ENABLE         = 0x01

	// Timestamp Outputs Register Adresses
	TIMESTAMP_OUT_L = 0x30
	TIMESTAMP_OUT_M = 0x31
	TIMESTAMP_OUT_H = 0x32

	// Temperature Outputs Register Adresses
	TEMP_OUT_L = 0x33
	TEMP_OUT_H = 0x34

	// Acceleration Outputs Register Adresses
	ACC_XOUT_L = 0x35
	ACC_XOUT_H = 0x36
	ACC_YOUT_L = 0x37
	ACC_YOUT_H = 0x38
	ACC_ZOUT_L = 0x39
	ACC_ZOUT_H = 0x3A

	// Angular Rate Outputs Register Adresses
	GYRO_XOUT_L = 0x3B
	GYRO_XOUT_H = 0x3C
	GYRO_YOUT_L = 0x3D
	GYRO_YOUT_H = 0x3E
	GYRO_ZOUT_L = 0x3F
	GYRO_ZOUT_H = 0x40

	// Quaternion Outputs Register Adresses
	DELTA_QUAT_WOUT_L = 0x49
	DELTA_QUAT_WOUT_H = 0x4A
	DELTA_QUAT_XOUT_L = 0x4B
	DELTA_QUAT_XOUT_H = 0x4C
	DELTA_QUAT_YOUT_L = 0x4D
	DELTA_QUAT_YOUT_H = 0x4E
	DELTA_QUAT_ZOUT_L = 0x4F
	DELTA_QUAT_ZOUT_H = 0x50

	// Delta Velocity Outputs Register Adresses
	DELTA_VEL_XOUT_L = 0x51
	DELTA_VEL_XOUT_H = 0x52
	DELTA_VEL_YOUT_L = 0x53
	DELTA_VEL_YOUT_H = 0x54
	DELTA_VEL_ZOUT_L = 0x55
	DELTA_VEL_ZOUT_H = 0x56
)
