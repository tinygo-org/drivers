package enc28j60

// RCR
func (d *Dev) readCtlReg(addr uint8) uint8 {
	d.CSB.Low()
	var data [3]byte
	addr = ADDR_MASK & addr // first 3 bits are opcode

	// Reading MAC and MII registers requires a dummy read on intermediate byte (see page 28)
	if addr&0x80 != 0 {
		d.Bus.Tx([]byte{addr, 0, 0}, data[:])
		d.CSB.High()
		return data[2]
	}

	d.Bus.Tx([]byte{addr, 0}, data[:1])
	d.CSB.High()
	return data[1]
}

func (d *Dev) writeCtlReg(addr uint8, data []byte) {
	d.CSB.Low()
	addr = ENC28J60_WRITE_CTRL_REG | (ADDR_MASK & addr)

	d.Bus.Tx(append([]byte{addr}, data...), nil)

	d.CSB.High()
}

func (d *Dev) Reset() {
	d.CSB.Low()
	d.Bus.Tx([]byte{0xff}, nil)
	d.CSB.High()
}
