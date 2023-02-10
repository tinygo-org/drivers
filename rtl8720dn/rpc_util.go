package rtl8720dn

import (
	"fmt"
	"io"
)

var (
	headerBuf            [4]byte
	readBuf              [4]byte
	startWriteMessageBuf [1024]byte
	payload              [2048]byte
)

const (
	xVersion = 1
)

func startWriteMessage(msgType, service, requestNumber, sequence uint32) []byte {
	startWriteMessageBuf[0] = byte(msgType)
	startWriteMessageBuf[1] = byte(requestNumber)
	startWriteMessageBuf[2] = byte(service)
	startWriteMessageBuf[3] = byte(xVersion)

	startWriteMessageBuf[4] = byte(sequence)
	startWriteMessageBuf[5] = byte(sequence >> 8)
	startWriteMessageBuf[6] = byte(sequence >> 16)
	startWriteMessageBuf[7] = byte(sequence >> 24)

	return startWriteMessageBuf[:8]
}

func (r *rtl8720dn) performRequest(msg []byte) {
	crc := computeCRC16(msg)
	headerBuf[0] = byte(len(msg))
	headerBuf[1] = byte(len(msg) >> 8)
	headerBuf[2] = byte(crc)
	headerBuf[3] = byte(crc >> 8)

	if r.debug {
		fmt.Printf("tx : %2d : ", len(headerBuf))
		dumpHex(headerBuf[:])
		fmt.Printf("\r\n")
	}

	r.uart.Write(headerBuf[:])

	if r.debug {
		fmt.Printf("tx : %2d : ", len(msg))
		dumpHex(msg)
		fmt.Printf("\r\n")
	}
	r.uart.Write(msg)
}

func dumpHex(b []byte) {
	for i := range b {
		if i == 0 {
			fmt.Printf("%02X", b[i])
		} else {
			fmt.Printf(" %02X", b[i])
		}
	}
}

func (r *rtl8720dn) read() {
	for {
		n, _ := io.ReadFull(r.uart, readBuf[:4])
		if n == 0 {
			continue
		}

		if r.debug {
			fmt.Printf("rx : %2d : ", n)
			dumpHex(readBuf[:n])
			fmt.Printf("\r\n")
		}

		length := uint16(readBuf[0]) + uint16(readBuf[1])<<8
		crc := uint16(readBuf[2]) + uint16(readBuf[3])<<8

		n, _ = io.ReadFull(r.uart, payload[:length])
		if r.debug {
			fmt.Printf("rx : %2d : ", length)
			dumpHex(payload[0:n])
			fmt.Printf("\r\n")
		}

		n = int(length)

		crcNew := computeCRC16(payload[:n])
		if g, e := crcNew, crc; g != e {
			fmt.Printf("err CRC16: got %04X want %04X\r\n", g, e)
		}
		if payload[0] == 0x02 || payload[0] == 0x00 {
			return
		}
	}
}
