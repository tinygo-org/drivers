package rtl8720dn

import (
	"encoding/binary"
	"fmt"
	"io"
	"machine"
	"runtime"
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

type blockingUARTReader struct {
	uart *machine.UART
}

func (r blockingUARTReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	for {
		n, err := r.uart.Read(p)
		if n != 0 || err != nil {
			return n, err
		}
		runtime.Gosched()
	}
}

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
	if r.rpcErr != nil {
		*r.rpcErr = nil
	}
	crc := computeCRC16(msg)
	headerBuf[0] = byte(len(msg))
	headerBuf[1] = byte(len(msg) >> 8)
	headerBuf[2] = byte(crc)
	headerBuf[3] = byte(crc >> 8)

	for _, b := range [2][]byte{headerBuf[:], msg} {
		if r.debug {
			fmt.Printf("tx : %2d : ", len(b))
			dumpHex(b)
			fmt.Printf("\r\n")
		}
		n, err := r.uart.Write(b)
		if r.rpcErr != nil && (err != nil || n != len(b)) {
			if err == nil {
				err = io.ErrShortWrite
			}
			*r.rpcErr = err
			return
		}
	}
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
	if r.rpcErr != nil {
		defer func() {
			if *r.rpcErr != nil {
				clear(payload[:12])
			}
		}()
		if *r.rpcErr != nil {
			return
		}
		reader := blockingUARTReader{uart: r.uart}
		for {
			if _, err := io.ReadFull(reader, readBuf[:]); err != nil {
				*r.rpcErr = err
				return
			}
			length := binary.LittleEndian.Uint16(readBuf[:2])
			if length < 8 || int(length) > len(payload) {
				*r.rpcErr = fmt.Errorf("Invalid TLS RPC frame length %d", length)
				return
			}
			if _, err := io.ReadFull(reader, payload[:length]); err != nil {
				*r.rpcErr = err
				return
			}
			if computeCRC16(payload[:length]) != binary.LittleEndian.Uint16(readBuf[2:]) {
				*r.rpcErr = fmt.Errorf("TLS RPC response CRC mismatch")
				return
			}
			if payload[3] != xVersion {
				*r.rpcErr = fmt.Errorf("Invalid TLS RPC response header")
				return
			}
			switch payload[0] {
			case 0x02:
			case 0x01, 0x03:
				continue
			case 0x00:
				*r.rpcErr = fmt.Errorf("unsupported incoming RPC invocation")
				return
			default:
				*r.rpcErr = fmt.Errorf("Invalid TLS RPC message type %d", payload[0])
				return
			}
			if length != 12 {
				*r.rpcErr = fmt.Errorf("Invalid TLS RPC response length %d", length)
				return
			}
			if payload[1] != startWriteMessageBuf[1] ||
				payload[2] != startWriteMessageBuf[2] ||
				binary.LittleEndian.Uint32(payload[4:8]) != uint32(r.seq) {
				*r.rpcErr = fmt.Errorf("Invalid TLS RPC response header")
			}
			return
		}
	}
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
