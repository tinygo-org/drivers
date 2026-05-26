package bme688

// I2C addresses
const (
	Address    uint16 = 0x77
	AddressAlt uint16 = 0x76
)

// Chip identifiers
const (
	CHIP_ID        = 0x61
	VARIANT_BME680 = 0x00
	VARIANT_BME688 = 0x01
)

// Registers
const (
	REG_COEFF_3       = 0x00 // res_heat_val, res_heat_range, range_sw_err
	REG_CTRL_GAS_0    = 0x70 // heat_off
	REG_CTRL_GAS_1    = 0x71 // run_gas, nb_conv
	REG_CTRL_HUM      = 0x72 // osrs_h
	REG_CTRL_MEAS     = 0x74 // osrs_t, osrs_p, mode
	REG_CONFIG        = 0x75 // filter
	REG_FIELD_0       = 0x1D // start of 17-byte measurement field
	REG_MEAS_STATUS_0 = 0x1D // measurement status
	REG_RES_HEAT_0    = 0x5A // target heater resistance (profile 0..9 at 0x5A..0x63)
	REG_GAS_WAIT_0    = 0x64 // heater on duration (profile 0..9 at 0x64..0x6D)
	REG_COEFF_1       = 0x8A // T2, T3, P1..P10 calibration coefficients
	REG_COEFF_2       = 0xE1 // H2, H1, H3..H7, T1, GH2, GH1, GH3
	REG_CHIP_ID       = 0xD0
	REG_VARIANT_ID    = 0xF0
	REG_RESET         = 0xE0
)

const (
	LEN_COEFF_1 = 23
	LEN_COEFF_2 = 14
	LEN_COEFF_3 = 5
	LEN_FIELD   = 17
)

// Reset command
const RESET_CMD = 0xB6

// Measurement status bits (REG_MEAS_STATUS_0 / field byte 0)
const (
	NEW_DATA_MSK  = 0x80 // new data available
	GAS_MEAS_MSK  = 0x40 // gas measurement ongoing
	MEASURING_MSK = 0x20 // TPH measurement ongoing
	GAS_IDX_MSK   = 0x0F // gas heater profile index used
)

// Gas result bits (field byte 14, i.e. field[0x1D + 14])
const (
	GAS_VALID_MSK = 0x20 // gas measurement result valid
	HEAT_STAB_MSK = 0x10 // heater temperature stable
	GAS_RANGE_MSK = 0x0F // gas ADC range
)

// ctrl_gas_0 bits
const HEAT_OFF_MSK = 0x08 // set to disable heater

// ctrl_gas_1 run_gas bit per variant
const (
	RUN_GAS_680 = 0x10 // bit 4 for BME680
	RUN_GAS_688 = 0x20 // bit 5 for BME688 (low-range gas measurement)
)

// Oversampling settings for temperature, pressure, and humidity
type Oversampling byte

const (
	SamplingOff Oversampling = 0b000
	Sampling1X  Oversampling = 0b001
	Sampling2X  Oversampling = 0b010
	Sampling4X  Oversampling = 0b011
	Sampling8X  Oversampling = 0b100
	Sampling16X Oversampling = 0b101
)

// Mode settings
type Mode byte

const (
	ModeSleep  Mode = 0x00
	ModeForced Mode = 0x01
)

// IIR filter coefficients
type Filter byte

const (
	FilterOff Filter = 0b000
	Filter1   Filter = 0b001
	Filter3   Filter = 0b010
	Filter7   Filter = 0b011
	Filter15  Filter = 0b100
	Filter31  Filter = 0b101
	Filter63  Filter = 0b110
	Filter127 Filter = 0b111
)

// Gas resistance lookup tables from Bosch BME68x API
var gasLookupTable1 = [16]uint32{
	2147483647, 2147483647, 2147483647, 2147483647,
	2147483647, 2126008810, 2147483647, 2130303777,
	2147483647, 2147483647, 2143188679, 2136746228,
	2147483647, 2126008810, 2147483647, 2147483647,
}

var gasLookupTable2 = [16]uint32{
	4096000000, 2048000000, 1024000000, 512000000,
	255744255, 127110228, 64000000, 32258064,
	16016016, 8000000, 4000000, 2000000,
	1000000, 500000, 250000, 125000,
}
