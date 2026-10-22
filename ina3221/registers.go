package ina3221

// Address is the I2C address with A0 connected to GND.
// Connect A0 to VS, SDA, or SCL for 0x41, 0x42, or 0x43.
const Address = 0x40

// Registers from the datasheet, section 8.6.
const (
	RegConfig         = 0x00
	RegShuntVoltage1  = 0x01
	RegBusVoltage1    = 0x02
	RegShuntVoltage2  = 0x03
	RegBusVoltage2    = 0x04
	RegShuntVoltage3  = 0x05
	RegBusVoltage3    = 0x06
	RegCriticalLimit1 = 0x07
	RegWarningLimit1  = 0x08
	RegCriticalLimit2 = 0x09
	RegWarningLimit2  = 0x0A
	RegCriticalLimit3 = 0x0B
	RegWarningLimit3  = 0x0C
	RegShuntSum       = 0x0D
	RegShuntSumLimit  = 0x0E
	RegMaskEnable     = 0x0F
	RegPowerValidHigh = 0x10
	RegPowerValidLow  = 0x11
	RegManufacturerID = 0xFE
	RegDieID          = 0xFF
)

const (
	ManufacturerID = 0x5449
	DieID          = 0x3220
)

const (
	configReset = 1 << 15

	maskConversionReady = 1 << 0
)
