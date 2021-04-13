package enc28j60

import (
	"runtime/interrupt"
	"time"
)

func (d *Dev) readOp(op, address uint8) uint8 {
	cmd := [1]byte{op | (address & ADDR_MASK)}
	var read [1]byte

	d.enableCS()

	err := d.Bus.Tx(cmd[:], read[:])
	dbp("read addr:", []byte{address})
	dbp("got:", read[:])
	if err != nil {
		dbp("error read addr:", []byte{address})
		dbp(err.Error(), []byte{address})
	}
	// do dummy read if needed (for mac and mii, see datasheet page 29)
	if address&0x80 != 0 {
		d.Bus.Tx(d.dummy[0:1], nil)
	}
	d.disableCS()
	return read[0]
}

func (d *Dev) writeOp(op, address, data uint8) {
	d.enableCS()
	cmd := [2]byte{op | (address & ADDR_MASK), data}
	err := d.Bus.Tx(cmd[:], nil)
	dbp("write addr:", []byte{address})
	if err != nil {
		dbp(err.Error(), []byte{op})
	}
	d.disableCS()
}

// RCR
func (d *Dev) readCtlReg(addr uint8) uint8 {
	d.enableCS()
	var data [3]byte
	addr = ADDR_MASK & addr // first 3 bits are opcode

	// Reading MAC and MII registers requires a dummy read on intermediate byte (see page 28)
	if addr&0x80 != 0 {
		d.Bus.Tx([]byte{addr, 0, 0}, data[:])
		d.disableCS()
		return data[2]
	}

	d.Bus.Tx([]byte{addr, 0}, data[:1])
	d.disableCS()
	return data[1]
}

func (d *Dev) writeCtlReg(addr uint8, data []byte) {
	d.enableCS()
	addr = ENC28J60_WRITE_CTRL_REG | (ADDR_MASK & addr)

	d.Bus.Tx(append([]byte{addr}, data...), nil)

	d.disableCS()
}

func (d *Dev) Reset() {
	d.enableCS()
	d.Bus.Tx([]byte{ENC28J60_SOFT_RESET}, nil)
	d.disableCS()
}

// TODO remove when certian not a problem
const setTime = time.Millisecond * 2

// enableCS enables SPI communication on bus. Disables Interrupts.
// do not call enableCS twice before calling disable
func (d *Dev) enableCS() {
	d.is = interrupt.Disable()
	d.CSB.Low()
}

// disableCS ends SPI communication on bus
// always call disableCS after calling enable once
// critical part done
func (d *Dev) disableCS() {
	d.CSB.High()
	interrupt.Restore(d.is)
}
