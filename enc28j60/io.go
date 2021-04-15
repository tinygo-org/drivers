package enc28j60

import (
	"runtime/interrupt"
	"time"
)

// the ENC28J60 has 4 banks (0 through 3). It can only read/write to
// one at a time, and much switch between them by writing to ECON1 register.
func (d *Dev) setBank(address uint8) {
	bank := address & BANK_MASK
	if bank != d.Bank {
		d.writeOp(BIT_FIELD_CLR, ECON1, ECON1_BSEL1|ECON1_BSEL0)
		d.writeOp(BIT_FIELD_SET, ECON1, bank>>5)
		d.Bank = bank
	}
}

// readOp reads from a register defined in registers.go. It requires
// the ENC28J60 Bank be set beforehand.
func (d *Dev) readOp(op, address uint8) uint8 {
	var read [2]byte

	d.enableCS()
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

func (d *Dev) read(address uint8) uint8 {
	d.setBank(address)
	return d.readOp(READ_CTL_REG, address)
}

func (d *Dev) write(address, data uint8) {
	d.setBank(address)
	d.writeOp(WRITE_CTL_REG, address, data)
}

func (d *Dev) phyWrite(address uint8, data uint16) {
	// set the PHY register address
	d.write(MIREGADR, address)
	// write the PHY data
	d.write(MIWRL, uint8(data))
	d.write(MIWRH, uint8(data>>8))
	// wait until the PHY write completes
	for d.read(MISTAT)&MISTAT_BUSY != 0 {
		time.Sleep(time.Microsecond * 15)
	}
}

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
