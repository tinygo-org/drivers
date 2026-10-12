package rtl8720dn

import (
	"encoding/binary"
	"fmt"
	"io"
)

var (
	headerBuf            [4]byte
	readBuf              [4]byte
	startWriteMessageBuf [1024]byte
	payload              [2048]byte
)

type rpcState struct {
	uart          io.ReadWriter
	seq           uint64
	debug         bool
	expected      [8]byte
	rpcErr        error
	payloadLength int
}

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

func (r *rtl8720dn) request(msg []byte) error {
	if err := r.writeRequest(msg); err != nil {
		return err
	}
	return r.readResponse()
}

func (r *rtl8720dn) performRequest(msg []byte) {
	r.rpcErr = r.writeRequest(msg)
}

func (r *rtl8720dn) writeRequest(msg []byte) error {
	r.payloadLength = 0
	if len(msg) < 8 || len(msg) > 65535 {
		return fmt.Errorf("invalid RPC request length %d", len(msg))
	}
	copy(r.expected[:], msg[:8])
	binary.LittleEndian.PutUint16(headerBuf[:2], uint16(len(msg)))
	binary.LittleEndian.PutUint16(headerBuf[2:], computeCRC16(msg))
	for _, data := range [][]byte{headerBuf[:], msg} {
		if r.debug {
			fmt.Printf("tx : %2d : ", len(data))
			dumpHex(data)
			fmt.Printf("\r\n")
		}
		n, err := r.uart.Write(data)
		if err != nil {
			return fmt.Errorf("RPC write: %w", err)
		}
		if n != len(data) {
			return fmt.Errorf("RPC write: %w", io.ErrShortWrite)
		}
	}
	return nil
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
	if r.rpcErr == nil {
		r.rpcErr = r.readResponse()
	}
	if r.rpcErr != nil {
		clear(payload[:])
	}
}

func (r *rtl8720dn) readResponse() error {
	r.payloadLength = 0
	for {
		if _, err := io.ReadFull(r.uart, readBuf[:]); err != nil {
			return fmt.Errorf("RPC response header: %w", err)
		}
		length := int(binary.LittleEndian.Uint16(readBuf[:2]))
		if length < 8 || length > len(payload) {
			return fmt.Errorf("invalid RPC response length %d", length)
		}
		data := payload[:length]
		if _, err := io.ReadFull(r.uart, data); err != nil {
			return fmt.Errorf("RPC response body: %w", err)
		}
		if r.debug {
			fmt.Printf("rx : %2d : ", length)
			dumpHex(data)
			fmt.Printf("\r\n")
		}
		if computeCRC16(data) != binary.LittleEndian.Uint16(readBuf[2:]) {
			return fmt.Errorf("invalid RPC response CRC")
		}
		if data[3] != xVersion {
			return fmt.Errorf("invalid RPC response version %d", data[3])
		}
		switch data[0] {
		case 0x01, 0x03:
			continue
		case 0x00, 0x02:
		default:
			return fmt.Errorf("invalid RPC response type %d", data[0])
		}
		if data[1] != r.expected[1] || data[2] != r.expected[2] ||
			binary.LittleEndian.Uint32(data[4:8]) != binary.LittleEndian.Uint32(r.expected[4:8]) {
			return fmt.Errorf("invalid RPC response header")
		}
		r.payloadLength = length
		return nil
	}
}
