package enc28j60

import (
	"machine"
	"time"
)

var errTest ErrorCode = 255

// this file stores test functions the client can run in order to verify
// the enc28j60's SPI/Ethernet connection and functioning. Results are printed
// to serial
func TestSPI(csb machine.Pin, spi machine.SPI, frequency uint32) error {
	e, err := New(csb, spi, frequency)
	if err != nil {
		println("spi config fail")
		return errTest
	}
	// e.Reset()
	// e.configure()
	// e.configure([]byte{0, 0, 0, 0, 0, 0})
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
	// d.write(ERXSTL, &0xFF)
	// d.write(ERXSTH, RXSTART_INIT>>8)
	cnt := 0
	var old, new uint8
	// rx := [2]byte{}
	for i := range ops {
		// Set Bank
		bank := (ops[i].addr & BANK_MASK) >> 5
		// e.enableCS()
		// e.Bus.Tx([]byte{ECON1 | BIT_FIELD_CLR, BANK_MASK >> 5}, nil)
		// e.disableCS()
		// set bitField set bankmask
		econCleared := e.readOp(READ_CTL_REG, ECON1) &^ 0b11
		e.writeOp(WRITE_CTL_REG, ops[i].addr, econCleared|bank)

		readbank := uint8(e.readOp(READ_CTL_REG, ECON1)&BANK_MASK) >> 5
		// Read previous address data
		old = e.readOp(READ_CTL_REG, ops[i].addr)

		e.writeOp(WRITE_CTL_REG, ops[i].addr, ops[i].new)

		new = e.readOp(READ_CTL_REG, ops[i].addr)

		time.Sleep(time.Microsecond * 50)
		if new != ops[i].new {
			cnt++
			println("addr:", "0x"+string(byteToHex(ops[i].addr&ADDR_MASK)),
				", wrote ", "0x"+string(byteToHex(ops[i].new)),
				", read back", "0x"+string(byteToHex(new)),
				", old was ", "0x"+string(byteToHex(old)),
				", read bank:", readbank,
				", bankmask used:", bank)
		}
	}
	if cnt > 0 {
		println("some inconsistencies were found in SPI test. Rev", e.GetRev())
		return errTest
	}

	println("All tests passed! Rev", e.GetRev())
	return nil
}
