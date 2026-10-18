package ina3221

// Address is the I2C address with A0 connected to GND.
// Connect A0 to VS, SDA, or SCL for 0x41, 0x42, or 0x43.
const Address = 0x40

// Registers from the datasheet, section 8.6.
const (
	RegConfig         uint8 = 0x00
	RegShuntVoltage1  uint8 = 0x01
	RegBusVoltage1    uint8 = 0x02
	RegShuntVoltage2  uint8 = 0x03
	RegBusVoltage2    uint8 = 0x04
	RegShuntVoltage3  uint8 = 0x05
	RegBusVoltage3    uint8 = 0x06
	RegCriticalLimit1 uint8 = 0x07
	RegWarningLimit1  uint8 = 0x08
	RegCriticalLimit2 uint8 = 0x09
	RegWarningLimit2  uint8 = 0x0A
	RegCriticalLimit3 uint8 = 0x0B
	RegWarningLimit3  uint8 = 0x0C
	RegShuntSum       uint8 = 0x0D
	RegShuntSumLimit  uint8 = 0x0E
	RegMaskEnable     uint8 = 0x0F
	RegPowerValidHigh uint8 = 0x10
	RegPowerValidLow  uint8 = 0x11
	RegManufacturerID uint8 = 0xFE
	RegDieID          uint8 = 0xFF
)

const (
	ManufacturerID = 0x5449
	DieID          = 0x3220
)

const (
	configReset = 1 << 15

	maskConversionReady = 1 << 0
)
