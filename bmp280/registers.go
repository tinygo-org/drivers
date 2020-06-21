// Package bmp280 provides a driver for the BMP280 digital temperature & pressure sensor by Bosch.
//
// Datasheet: https://www.bosch-sensortec.com/media/boschsensortec/downloads/datasheets/bst-bmp280-ds001.pdf
package bmp280

// The I2C address which this device listens to.
const Address = 0x77

// Registers
const (
	REG_ID        = 0xD0 // WHO_AM_I
	REG_RESET     = 0xE0
	REG_STATUS    = 0xF3
	REG_CTRL_MEAS = 0xF4
	REG_CONFIG    = 0xF5
	REG_TEMP      = 0xFA
	REG_PRES      = 0xF7
	REG_CALI      = 0x88

	CHIP_ID   = 0x58
	CMD_RESET = 0xB6
)

const (
	SAMPLING_SKIPPED Oversampling = iota
	SAMPLING_1X
	SAMPLING_2X
	SAMPLING_4X
	SAMPLING_8X
	SAMPLING_16X
)

const (
	MODE_SLEEP  Mode = 0x00
	MODE_FORCED Mode = 0x01
	MODE_NORMAL Mode = 0x03
)

const (
	STANDBY_1MS Standby = iota
	STANDBY_63MS
	STANDBY_125MS
	STANDBY_250MS
	STANDBY_500MS
	STANDBY_1000MS
	STANDBY_2000MS
	STANDBY_4000MS
)

const (
	FILTER_OFF Filter = iota
	FILTER_2X
	FILTER_4X
	FILTER_8X
	FILTER_16X
)
