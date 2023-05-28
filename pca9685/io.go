package pca9685

import "tinygo.org/x/drivers/internal/legacy"

func (d *Dev) readReg(reg uint8, data []byte) error {
	return legacy.ReadRegister(d.bus, d.addr, reg, data)
}

func (d *Dev) writeReg(reg uint8, data []byte) error {
	return legacy.WriteRegister(d.bus, d.addr, reg, data)
}
