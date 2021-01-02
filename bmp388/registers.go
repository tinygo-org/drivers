// Package bmp388 provides a driver for Bosch's BMP388 digital temperature & pressure sensor.
// The datasheet can be found here: https://www.bosch-sensortec.com/media/boschsensortec/downloads/datasheets/bst-bmp388-ds001.pdf
package bmp388

const ADDRESS byte = 0x77 // default I2C address

const (
	REG_CHIP_ID  byte = 0x00 // useful for checking the connection
	REG_CALI     byte = 0x31 // pressure & temperature compensation calibration coefficients
	REG_PRESS    byte = 0x04 // start of pressure data registers
	REG_TEMP     byte = 0x07 // start of temperature data registers
	REG_PWR_CTRL byte = 0x1B // measurement mode & pressure/temperature sensor power register
	REG_OSR      byte = 0x1C // oversampling settings register
	REG_ODR      byte = 0x1D //
	REG_CMD      byte = 0x7E // miscellaneous command register
	REG_STAT     byte = 0x03 // sensor status register
	REG_ERR      byte = 0x02 // error status register
	REG_IIR      byte = 0x1F
)

const (
	CHIP_ID    byte = 0x50 // correct response if reading from chip id register
	PWR_PRESS  byte = 0x01 // power on pressure sensor
	PWR_TEMP   byte = 0x02 // power on temperature sensor
	SOFT_RESET byte = 0xB6 // command to reset all user configuration
	DRDY_PRESS byte = 0x20 // for checking if pressure data is ready
	DRDY_TEMP  byte = 0x40 // for checking if pressure data is ready
)

// The difference between forced and normal mode is the bmp388 goes to sleep after taking a measurement in forced mode.
// Set it to forced if you intend to take measurements sporadically and want to save power. The driver will handle
// waking the sensor up when the sensor is in forced mode.
const (
	NORMAL Mode = 0x30
	FORCED Mode = 0x16
	SLEEP  Mode = 0x00
)

// Increasing sampling rate increases precision but also the wait time for measurements. The datasheet has a table of
// suggested values for oversampling, output data rates, and iir filter coefficients by use case.
const (
	SAMPLING_1X Oversampling = iota
	SAMPLING_2X
	SAMPLING_4X
	SAMPLING_8X
	SAMPLING_16X
	SAMPLING_32X
)

// Output data rates in Hz. If increasing the sampling rates you need to decrease the output data rates, else the bmp388
// will freeze and Configure() will return a configuration error message. In that case keep decreasing the data rate
// until the bmp is happy
const (
	ODR_200 OutputDataRate = iota
	ODR_100
	ODR_50
	ODR_25
	ODR_12p5
	ODR_6p25
	ODR_3p1
	ODR_1p5
	ODR_0p78
	ODR_0p39
	ODR_0p2
	ODR_0p1
	ODR_0p05
	ODR_0p02
	ODR_0p01
	ODR_0p006
	ODR_0p003
	ODR_0p0015
)

// IIR filter coefficients, higher values means steadier measurements but slower reaction times
const (
	COEF_0 FilterCoefficient = iota
	COEF_1
	COEF_3
	COEF_7
	COEF_15
	COEF_31
	COEF_63
	COEF_127
)
