package ina219

// The default I2C address for this device.
const Address = 0x40

const (
	RegConfig       uint8 = 0x0
	RegShuntVoltage uint8 = 0x1
	RegBusVoltage   uint8 = 0x2
	RegPower        uint8 = 0x3
	RegCurrent      uint8 = 0x4
	RegCalibration  uint8 = 0x5
)
