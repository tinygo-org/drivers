package enc28j60

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/net"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/encoding/hex"
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
			println("addr:", "0x"+string(hex.Byte(ops[i].addr&ADDR_MASK)),
				", wrote ", "0x"+string(hex.Byte(ops[i].new)),
				", read back", "0x"+string(hex.Byte(new)),
				", old was ", "0x"+string(hex.Byte(old)),
				", read bank:", readbank,
				", bankmask used:", (ops[i].addr&BANK_MASK)>>5)
		}
	}
	failures += testReg16(e)

	if failures > 0 {
		println("some inconsistencies were found in SPI test. Rev", e.GetRev())
		return errTest
	}

	println("All tests passed! Rev", e.GetRev())
	return nil
}

func testReg16(e *Dev) (failures int) {
	var ops = []struct {
		addr uint8
		// default, new values
		def, new uint16
	}{
		{addr: ERXSTL, def: 0x05FA, new: RXSTART_INIT},
		{addr: ERDPTL, def: 0x05FA, new: 0x1fff},
	}
	var old, new uint16
	for i := range ops {
		old = e.read16(ops[i].addr)
		// write new data
		e.write16(ops[i].addr, ops[i].new)
		// read new data to check if all was written OK
		new = e.read16(ops[i].addr)
		if new != ops[i].new {
			failures++
			println("addrL:", "0x"+string(hex.Byte(ops[i].addr&ADDR_MASK)),
				", wrote ", "0x"+string(hex.Byte(uint8(ops[i].new>>8)))+string(hex.Byte(uint8(ops[i].new))),
				", read back", "0x"+string(hex.Byte(uint8(new>>8)))+string(hex.Byte(uint8(new))),
				", old was ", "0x"+string(hex.Byte(uint8(old>>8)))+string(hex.Byte(uint8(old))))
		}
	}
	return
}

// Prints data recieved
func TestConn(csb machine.Pin, spi drivers.SPI) error {
	const plen = 300
	ebuff := [plen]byte{}
	var (
		// // Hardware address of ENC28J60
		macAddr = net.HardwareAddr{0xDE, 0xAD, 0xBE, 0xEF, 0xFE, 0xFF}
	)
	// macaddr := []byte{0xde, 0xad, 0xfe, 0xfe, 0xfe, 0xfe}
	// macaddr := []byte{0xde, 0xad, 0xfe, 0xff, 0xef, 0xee}
	buff := [plen]byte{}
	e := New(csb, spi)
	err := e.Init(ebuff[:], macAddr)
	if err != nil {
		return err
	}
	println("macaddr:", string(hex.Bytes(macAddr)))
	println("recieving data...")
	// TODO figure out how buffer works
	for {
		time.Sleep(time.Second)
		plen := e.PacketRecieve(buff[:])
		if plen == 0 {
			continue
		}
		println("recieve data:")
		println(string(hex.Bytes(buff[:plen])))
	}
	return nil
}
