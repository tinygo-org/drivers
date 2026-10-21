package bme68x

const (
	// REG_CTRL_MEAS is the CTRL_MEAS address
	REG_CTRL_MEAS uint8 = 0x74
	// REG_CTRL_HUM is the CTRL_HUM address
	REG_CTRL_HUM uint8 = 0x72
	// REG_CONFIG is the CONFIG address
	REG_CONFIG uint8 = 0x75
	// MODE_MSK is the mask for operation mode
	MODE_MSK uint8 = 0x03

	// REG_SOFT_RESET is the soft reset address
	REG_SOFT_RESET uint8 = 0xE0
	// CMD_RESET is the soft reset command
	CMD_RESET uint8 = 0xB6

	// LEN_INTERLEAVE_BUFF is the length of the interleave buffer
	LEN_INTERLEAVE_BUFF uint8 = 20

	// REG_COEFF1 is the register address for 1st group of coefficients
	REG_COEFF1 uint8 = 0x8A
	// REG_COEFF2 is the register address for 2nd group of coefficients
	REG_COEFF2 uint8 = 0xE1
	// REG_COEFF3 is the register address for 3rd group of coefficients
	REG_COEFF3 uint8 = 0x00

	// REG_IDAC_HEAT0 is the 0th current DAC address
	REG_IDAC_HEAT0 uint8 = 0x50 // idac_heat0
	// GAS_WAIT0 is the 0th gas wait address
	REG_GAS_WAIT0 uint8 = 0x64 // gas_wait_0
	// REG_RES_HEAT0 is the 0th resistance heat address
	REG_RES_HEAT0 uint8 = 0x5A // res_heat_0
	// REG_CTRL_GAS_0 is the CTRL_GAS_0 address
	REG_CTRL_GAS_0 uint8 = 0x70 // ctrl_gas_0
	// REG_CTRL_GAS_1 is the CTRL_GAS_1 address
	REG_CTRL_GAS_1 uint8 = 0x71 // ctrl_gas_1
	// REG_VARIANT_ID is the variant ID address
	REG_VARIANT_ID uint8 = 0xF0 // variant_id

	// MEAS_STATUS_0 is the measurement status address
	MEAS_STATUS_0 uint8 = 0x1D
	// NEW_DATA_MSK is the mask for new data
	NEW_DATA_MSK uint8 = 0x80
	// GAS_INDEX_MSK is the mask for gas index
	GAS_INDEX_MSK uint8 = 0x0F
	// GAS_RANGE_MSK is the mask for gas range
	GAS_RANGE_MSK uint8 = 0x0F
	// GASM_VALID_MSK is the mask for gas measurement valid
	GASM_VALID_MSK uint8 = 0x20
	// HEAT_STAB_MSK is the mask for heater stability
	HEAT_STAB_MSK uint8 = 0x10
	// VARIANT_GAS_HIGH is the high gas variant
	VARIANT_GAS_HIGH uint8 = 0x01
	// OSH_MSK is the mask for humidity oversampling
	OSH_MSK uint8 = 0x07
	// OSP_MSK is the mask for pressure oversampling
	OSP_MSK uint8 = 0x1C
	// OST_MSK is the mask for temperature oversampling
	OST_MSK uint8 = 0xE0
	// FILTER_MSK is the mask for IIR filter
	FILTER_MSK uint8 = 0x1C
	// ODR20_MSK is the mask for ODR[2:0]
	ODR20_MSK uint8 = 0xE0
	// ODR3_MSK is the mask for ODR[3]
	ODR3_MSK uint8 = 0x80
	// ENABLE_HEATER enables heater
	ENABLE_HEATER uint8 = 0x00
	// DISABLE_HEATER disables heater
	DISABLE_HEATER uint8 = 0x01
	// ENABLE_GAS_MEAS_H enables gas measurement high
	ENABLE_GAS_MEAS_H uint8 = 0x02
	// ENABLE_GAS_MEAS_L enables gas measurement low
	ENABLE_GAS_MEAS_L uint8 = 0x01
	// DISABLE_GAS_MEAS disables gas measurement
	DISABLE_GAS_MEAS uint8 = 0x00
	// HCTRL_MSK is the mask for heater control
	HCTRL_MSK uint8 = 0x08
	// NBCONV_MSK is the mask for number of conversions
	NBCONV_MSK uint8 = 0x0F
	// RUN_GAS_MSK is the mask for run gas
	RUN_GAS_MSK uint8 = 0x30
	// HCTRL_POS is the heater control bit position
	HCTRL_POS uint8 = 3
	// RUN_GAS_POS is the run gas bit position
	RUN_GAS_POS uint8 = 4
	// OSP_POS is the pressure oversampling bit position
	OSP_POS uint8 = 2
	// OST_POS is the temperature oversampling bit position
	OST_POS uint8 = 5
	// FILTER_POS is the filter bit position
	FILTER_POS uint8 = 2
	// ODR20_POS is the ODR[2:0] bit position
	ODR20_POS uint8 = 5
	// ODR3_POS is the ODR[3] bit position
	ODR3_POS uint8 = 7

	// CHIP_ID is the unique chip identifier
	CHIP_ID uint8 = 0x61
	// REG_CHIP_ID is the chip ID address
	REG_CHIP_ID uint8 = 0xD0
)

// Mode is the mode of the sensor.
type Mode byte

const (
	// ModeSleep is the sleep mode. The sensor will not take any measurements.
	ModeSleep Mode = 0x00
	// ModeForced is the forced mode. The sensor will take a measurement and store it in the
	// sensor's memory.
	ModeForced Mode = 0x01
)

// FilterCoefficient is the filter coefficient used for the sensor.
// The filter coefficient is used to filter the sensor data. Higher values means steadier
// measurements but slower reaction times.
type FilterCoefficient byte

const (
	// Coeff0 is no filter.
	Coeff0 FilterCoefficient = iota
	Coeff2
	Coeff4
	Coeff8
	Coeff16
	Coeff32
	Coeff64
	Coeff128
)

// Oversampling is the oversampling coefficient used for the sensor.
// The oversampling coefficient is used to increase the resolution of the sensor data.
type Oversampling byte

const (
	// SamplingOff turns off sensor reading.
	SamplingOff Oversampling = iota
	Sampling1X
	Sampling2X
	Sampling4X
	Sampling8X
	Sampling16X
)

// ODR is the output data rate of the sensor.
// The output data rate is the rate at which the sensor will take measurements.
type ODR byte

const (
	// Standby time of 0.59ms
	ODR_0_59 ODR = iota
	// Standby time of 62.5ms
	ODR_62_5
	// Standby time of 125ms
	ODR_125
	// Standby time of 250ms
	ODR_250
	// Standby time of 500ms
	ODR_500
	// Standby time of 1s
	ODR_1000
	// Standby time of 10ms
	ODR_10
	// Standby time of 20ms
	ODR_20
	// No standby time
	ODR_NONE
)
