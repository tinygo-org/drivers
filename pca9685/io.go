package pca9685

import "tinygo.org/x/drivers"

func (d *Dev) readReg(reg uint8, data []byte) error {
	return d.bus.Tx(uint16(d.addr), []byte{reg}, data)
}

func (d *Dev) writeReg(reg uint8, data []byte) error {
	var (
		bus  drivers.I2C = d.bus
		addr uint8       = d.addr
	)
	buf := make([]uint8, len(data)+1)
	buf[0] = reg
	copy(buf[1:], data)
	return bus.Tx(uint16(addr), buf, nil)
}
