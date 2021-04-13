package enc28j60

import (
	"runtime/interrupt"
	"time"
)

func (d *Dev) readOp(op, address uint8) uint8 {
	cmd := [1]byte{op | (address & ADDR_MASK)}
	var read [1]byte

	d.enable()

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
	d.disable()
	return read[0]
}

func (d *Dev) writeOp(op, address, data uint8) {
	d.enable()
	cmd := [2]byte{op | (address & ADDR_MASK), data}
	err := d.Bus.Tx(cmd[:], nil)
	dbp("write addr:", []byte{address})
	if err != nil {
		dbp(err.Error(), []byte{op})
	}
	d.disable()
}

// RCR
func (d *Dev) readCtlReg(addr uint8) uint8 {
	d.enable()
	var data [3]byte
	addr = ADDR_MASK & addr // first 3 bits are opcode

	// Reading MAC and MII registers requires a dummy read on intermediate byte (see page 28)
	if addr&0x80 != 0 {
		d.Bus.Tx([]byte{addr, 0, 0}, data[:])
		d.disable()
		return data[2]
	}

	d.Bus.Tx([]byte{addr, 0}, data[:1])
	d.disable()
	return data[1]
}

func (d *Dev) writeCtlReg(addr uint8, data []byte) {
	d.enable()
	addr = ENC28J60_WRITE_CTRL_REG | (ADDR_MASK & addr)

	d.Bus.Tx(append([]byte{addr}, data...), nil)

	d.disable()
}

func (d *Dev) Reset() {
	d.enable()
	d.Bus.Tx([]byte{ENC28J60_SOFT_RESET}, nil)
	d.disable()
}

// TODO remove when certian not a problem
const setTime = time.Millisecond * 2

// enable enables SPI communication on bus. Disables Interrupts.
// do not call enable twice before calling disable
func (d *Dev) enable() {
	d.is = interrupt.Disable()
	d.CSB.Low()
}

// disable ends SPI communication on bus
// always call disable after calling enable once
// critical part done
func (d *Dev) disable() {
	d.CSB.High()
	interrupt.Restore(d.is)
}
