// Package bmp388 provides a driver for the BMP388 digital temperature & pressure sensor by Bosch.
//
// Datasheet: https://www.bosch-sensortec.com/media/boschsensortec/downloads/datasheets/bst-bmp388-ds001.pdf
package bmp388

const (
	ADDRESS      byte = 0x77 // default I2C address
	REG_CHIP_ID  byte = 0x00 // useful for checking the connection
	REG_CALI     byte = 0x31 // pressure & temperature compensation calibration coefficients
	REG_PRESS    byte = 0x04 // start of pressure data registers
	REG_TEMP     byte = 0x07 // start of temperature data registers
	REG_PWR_CTRL byte = 0x1B // measurement mode & pressure/temperature sensor power register
	REG_OSR      byte = 0x1C // oversampling settings register
	REG_CMD      byte = 0x7E // miscellaneous command register
	REG_STAT     byte = 0x03 // sensor status register
	REG_ERR      byte = 0x02 // error status register
)

const (
	CHIP_ID    byte = 0x50 // correct response if reading from chip id register
	PWR_PRESS  byte = 0x01 // power on pressure sensor
	PWR_TEMP   byte = 0x02 // power on temperature sensor
	SOFT_RESET byte = 0xB6 // command to reset all user configuration
	DRDY_PRESS byte = 0x20 // for checking if pressure data is ready
	DRDY_TEMP  byte = 0x40 // for checking if pressure data is ready
)

const (
	NORMAL Mode = 0x30
	FORCED Mode = 0x16
	SLEEP  Mode = 0x00
)

const (
	SAMPLING_1X Oversampling = iota
	SAMPLING_2X
	SAMPLING_4X
	SAMPLING_8X
	SAMPLING_16X
	SAMPLING_32X
)
