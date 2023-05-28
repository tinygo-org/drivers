package legacy

import "tinygo.org/x/drivers"

type I2C struct {
	i2c drivers.I2C
}

func ReadRegister(i2c drivers.I2C, addr uint8, reg uint8, buf []byte) error {
	return i2c.Tx(uint16(addr), []byte{reg}, buf)
}

func WriteRegister(i2c drivers.I2C, addr uint8, reg uint8, buf []byte) error {
	return i2c.Tx(uint16(addr), append([]byte{reg}, buf...), nil)
}
