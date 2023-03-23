//	Product: https://ams.com/as5601
//	Datasheet: https://ams.com/documents/20143/36005/AS5601_DS000395_3-00.pdf

package as560x // import tinygo.org/x/drivers/ams560x

import "tinygo.org/x/drivers"

// AS5601Device represents an ams AS5601 device driver accessed over I2C
type AS5601Device struct {
	BaseDevice // promote base device
}

// NewAS5601 creates a new AS5601Device given an I2C bus
func NewAS5601(bus drivers.I2C) AS5601Device {
	// Create base device
	baseDev := newBaseDevice(bus)
	// Add AS5601 specific registers
	baseDev.registers[ABN] = newI2CRegister(ABN, 0, 0b1111, 1, reg_read|reg_write|reg_program)
	baseDev.registers[PUSHTHR] = newI2CRegister(PUSHTHR, 0, 0xff, 1, reg_read|reg_write|reg_program)
	// Return the device
	return AS5601Device{baseDev}
}
