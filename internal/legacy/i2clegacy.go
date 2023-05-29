package legacy

import "tinygo.org/x/drivers"

func ReadRegister(bus drivers.I2C, addr uint8, reg uint8, data []byte) error {
	return bus.Tx(uint16(addr), []byte{reg}, data)
}

func WriteRegister(bus drivers.I2C, addr uint8, reg uint8, data []byte) error {
	buf := make([]uint8, len(data)+1)
	buf[0] = reg
	copy(buf[1:], data)
	return bus.Tx(uint16(addr), buf, nil)
}
