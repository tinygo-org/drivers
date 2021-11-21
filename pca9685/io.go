package pca9685

func (d *Dev) readReg(reg uint8, data []byte) error {
	return d.bus.ReadRegister(d.addr, reg, data)
}

func (d *Dev) writeReg(reg uint8, data []byte) error {
	return d.bus.WriteRegister(d.addr, reg, data)
}
