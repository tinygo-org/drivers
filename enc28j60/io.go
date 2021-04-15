package enc28j60

import (
	"runtime/interrupt"
	"time"
)

func (d *Dev) readOp(op, address uint8) uint8 {
	d.enableCS()
	var read [2]byte
	d.Bus.Tx([]byte{op | (address & ADDR_MASK), 0}, read[:])
	dbp("RD addr, got:", []byte{address & ADDR_MASK}, read[1:])
	// do dummy read if needed (for mac and mii, see datasheet page 29)
	if address&SPRD_MASK != 0 {
		d.Bus.Tx(d.dummy[0:1], nil)
	}
	d.disableCS()
	return read[1]
}

func (d *Dev) writeOp(op, address, data uint8) {
	d.enableCS()
	err := d.Bus.Tx([]byte{op | (address & ADDR_MASK), data}, nil)
	dbp("WR addr, data:", []byte{address & ADDR_MASK}, []byte{data})
	if err != nil {
		dbp(err.Error(), []byte{op})
	}
	d.disableCS()
}

func (d *Dev) Reset() {
	d.enableCS()
	d.Bus.Tx([]byte{SOFT_RESET}, nil)
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
