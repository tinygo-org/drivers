package enc28j60

import (
	"machine"
	"time"
)

var errTest ErrorCode = 255

// this file stores test functions the client can run in order to verify
// the enc28j60's SPI/Ethernet connection and functioning. Results are printed
// to serial
func TestSPI(csb machine.Pin, spi machine.SPI) error {
	e := New(csb, spi)
	e.Reset()
	time.Sleep(time.Millisecond * 50)

	var ops = []struct {
		addr uint8
		// default, new values
		def, new uint8
	}{
		// first read All-bank registers
		// {addr: EIE, def: 0, new: 0b11},
		{addr: ECON1, def: 0, new: 0b11},
		{addr: ECON1, def: 0, new: 0b00},
		{addr: ECON1, def: 0, new: 0b10},
		// {addr: ESTAT, def: 0, new: 0b01},
		// commence reading registers in banks 0-3
		{addr: ERXSTL, def: 0b1111_1010, new: RXSTART_INIT},
		{addr: ERXSTH, def: 0b0_0101, new: RXSTART_INIT >> 8},
		{addr: MACON1, def: 0b0_0000, new: 0b1},
	}

	failures := 0
	var old, new uint8

	for i := range ops {
		// Set Bank
		e.setBank(ops[i].addr)
		// read back bank for debugging purposes
		readbank := uint8(e.readOp(READ_CTL_REG, ECON1)&BANK_MASK) >> 5

		// Read previous address data
		old = e.readOp(READ_CTL_REG, ops[i].addr)
		// write new data
		e.writeOp(WRITE_CTL_REG, ops[i].addr, ops[i].new)
		// read new data to check if all was written OK
		new = e.readOp(READ_CTL_REG, ops[i].addr)
		if new != ops[i].new {
			failures++
			println("addr:", "0x"+string(byteToHex(ops[i].addr&ADDR_MASK)),
				", wrote ", "0x"+string(byteToHex(ops[i].new)),
				", read back", "0x"+string(byteToHex(new)),
				", old was ", "0x"+string(byteToHex(old)),
				", read bank:", readbank,
				", bankmask used:", (ops[i].addr&BANK_MASK)>>5)
		}
	}
	if failures > 0 {
		println("some inconsistencies were found in SPI test. Rev", e.GetRev())
		return errTest
	}

	println("All tests passed! Rev", e.GetRev())
	return nil
}
