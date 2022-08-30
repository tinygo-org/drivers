package vl6180x

// The I2C address which this device listens to.
const Address = 0x29

// Registers
const (
	CHIP_ID                            = 0xB4
	WHO_AM_I                           = 0x0000
	SYSTEM_INTERRUPT_CONFIG            = 0x0014
	SYSTEM_INTERRUPT_CLEAR             = 0x0015
	SYSTEM_FRESH_OUT_OF_RESET          = 0x0016
	SYSRANGE_START                     = 0x0018
	SYSRANGE_PART_TO_PART_RANGE_OFFSET = 0x0024
	SYSALS_START                       = 0x0038
	SYSALS_ANALOGUE_GAIN               = 0x003F
	SYSALS_INTEGRATION_PERIOD_HI       = 0x0040
	SYSALS_INTEGRATION_PERIOD_LO       = 0x0041
	RESULT_RANGE_STATUS                = 0x004d
	RESULT_INTERRUPT_STATUS_GPIO       = 0x004f
	RESULT_ALS_VAL                     = 0x0050
	RESULT_RANGE_VAL                   = 0x0062
	I2C_SLAVE_DEVICE_ADDRESS           = 0x0212
	RANGING_INTERMEASUREMENT_PERIOD    = 0x001b
	ALS_INTERMEASUREMENT_PERIOD        = 0x003e

	ALS_GAIN_1    = 0x06 ///< 1x gain
	ALS_GAIN_1_25 = 0x05 ///< 1.25x gain
	ALS_GAIN_1_67 = 0x04 ///< 1.67x gain
	ALS_GAIN_2_5  = 0x03 ///< 2.5x gain
	ALS_GAIN_5    = 0x02 ///< 5x gain
	ALS_GAIN_10   = 0x01 ///< 10x gain
	ALS_GAIN_20   = 0x00 ///< 20x gain
	ALS_GAIN_40   = 0x07 ///< 40x gain

)

const (
	NONE VL6180XError = iota
	SYSERR_1
	SYSERR_5
	ECEFAIL
	NOCONVERGE
	RANGEIGNORE
	SNR
	RAWUFLOW
	RAWOFLOW
	RANGEUFLOW
	RANGEOFLOW
)
