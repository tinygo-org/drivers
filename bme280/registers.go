package bme280

// Constants/addresses used for I2C.

// The I2C address which this device listens to.
const Address = 0x76

// Registers. Names, addresses and comments copied from the datasheet.
const (
	CTRL_MEAS_ADDR        = 0xF4
	CTRL_HUMIDITY_ADDR    = 0xF2
	CTRL_CONFIG           = 0xF5
	REG_PRESSURE          = 0xF7
	REG_CALIBRATION       = 0x88
	REG_CALIBRATION_H1    = 0xA1
	REG_CALIBRATION_H2LSB = 0xE1
	CMD_RESET             = 0xE0

	WHO_AM_I = 0xD0
	CHIP_ID  = 0x60
)

// Increasing sampling rate increases precision but also the wait time for measurements. The datasheet has a table of
// suggested values for oversampling, output data rates, and iir filter coefficients by use case.
const (
	SamplingOff Oversampling = iota
	Sampling1X
	Sampling2X
	Sampling4X
	Sampling8X
	Sampling16X
)

// In normal mode (the default) the sensor takes masurements periodically.  In forced
// mode, the sensor takes a measurement only when requested.
//
// For use-cases with infrequent sampling, forced mode is more power efficient.
const (
	ModeNormal Mode = 0x03
	ModeForced Mode = 0x01
	ModeSleep  Mode = 0x00
)

// IIR filter coefficients, higher values means steadier measurements but slower reaction times
const (
	Coeff0 FilterCoefficient = iota
	Coeff2
	Coeff4
	Coeff8
	Coeff16
)

// Period of standby in normal mode which controls how often measurements are taken
//
// Note Period10ms and Period20ms are out of sequence, but are per the datasheet
const (
	Period0_5ms  Period = 0b000
	Period62_5ms        = 0b001
	Period125ms         = 0b010
	Period250ms         = 0b011
	Period500ms         = 0b100
	Period1000ms        = 0b101
	Period10ms          = 0b110
	Period20ms          = 0b111
)

const (
	SEALEVEL_PRESSURE float32 = 1013.25 // in hPa
)
