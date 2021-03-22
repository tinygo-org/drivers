package enc28j60

import (
	"machine"
	"time"

	"tinygo.org/x/drivers"
)

// this file stores test functions the client can run in order to verify
// the enc28j60's SPI/Ethernet connection and functioning. Results are printed
// to serial

func TestSPI(csb machine.Pin, spi drivers.SPI) {
	e := New(csb, spi)
	e.Reset()
	var ops = []struct {
		addr     uint8
		def, new uint8
	}{ // TXPAUS allow MAC to transmit pause control frames (needed for flow control in full duplex)
		{addr: MACON1, new: MACON1_MARXEN | MACON1_TXPAUS | MACON1_RXPAUS},
		// bring MAC out of reset
		{addr: MACON3, new: MACON3_ZPADCRC | MACON3_TXCRCEN | MACON3_FRMLNEN},
		{addr: MACON4, new: MACON4_DEFER},
		{addr: MABBIPG, new: 0x12},
		// MAC address
		{addr: MAADR5, new: 0xde},
		{addr: MAADR4, new: 0xad},
		{addr: MAADR3, new: 0xbe},
		{addr: MAADR2, new: 0xef},
		{addr: MAADR1, new: 0xfe},
		{addr: MAADR0, new: 0xff},
	}

	cnt := 0
	var old, new uint8
	for i := range ops {
		old = e.readCtlReg(ops[i].addr)
		// old = e.read(ops[i].addr)
		e.writeCtlReg(ops[i].addr, []byte{ops[i].new})
		// e.write(ops[i].addr, ops[i].new)
		time.Sleep(time.Microsecond * 50)
		new = e.read(ops[i].addr)
		if new != ops[i].new {
			cnt++

			println("addr:", "0x"+string(byteToHex(ops[i].addr)),
				", wrote ", "0x"+string(byteToHex(ops[i].new)),
				", read back", "0x"+string(byteToHex(new)),
				", old was ", "0x"+string(byteToHex(old)))
		}
	}
	if cnt > 0 {
		println("some inconsistencies were found in SPI test")
	}
}
