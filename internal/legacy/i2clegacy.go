package legacy

import "tinygo.org/x/drivers"

var i2cBuf []uint8 = make([]uint8, 1)

func ReadRegister(bus drivers.I2C, addr uint8, reg uint8, data []byte) error {
	i2cBuf[0] = reg
	return bus.Tx(uint16(addr), i2cBuf[:1], data)
}

func WriteRegister(bus drivers.I2C, addr uint8, reg uint8, data []byte) error {
	if len(i2cBuf) < 1+len(data) {
		i2cBuf = make([]uint8, 1+len(data))
	}
	i2cBuf[0] = reg
	copy(i2cBuf[1:], data)
	return bus.Tx(uint16(addr), i2cBuf[:1+len(data)], nil)
}
