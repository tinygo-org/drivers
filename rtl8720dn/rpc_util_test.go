//go:build rtl8720dn_hosttest

package rtl8720dn

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

func referenceCRC16(data []byte) uint16 {
	crc := uint16(0xef4a)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = crc<<1 ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

func frame(data []byte) []byte {
	header := make([]byte, 4)
	binary.LittleEndian.PutUint16(header[:2], uint16(len(data)))
	binary.LittleEndian.PutUint16(header[2:], referenceCRC16(data))
	return append(header, data...)
}

func response(msg []byte, result int32) []byte {
	data := append([]byte(nil), msg[:8]...)
	data[0] = 2
	data = binary.LittleEndian.AppendUint32(data, uint32(result))
	return data
}

type transport struct {
	*bytes.Reader
	writes     bytes.Buffer
	writeError error
	shortWrite bool
	failWrite  int
	calls      int
}

func (u *transport) Write(data []byte) (int, error) {
	u.calls++
	if u.failWrite == u.calls {
		if u.writeError != nil {
			return 0, u.writeError
		}
		if u.shortWrite {
			return len(data) - 1, nil
		}
	}
	return u.writes.Write(data)
}

func TestInt32(t *testing.T) {
	msg := []byte{0, 6, 17, 1, 7, 0, 0, 0, 42, 0, 0, 0}
	for _, result := range []int32{0, 3, -1, -29312} {
		t.Run(stringResult(result), func(t *testing.T) {
			uart := &transport{Reader: bytes.NewReader(frame(response(msg, result)))}
			r := &rtl8720dn{rpcState: rpcState{uart: uart}}
			err := r.request(msg)
			got := int32(binary.LittleEndian.Uint32(payload[8:12]))
			if err != nil || got != result {
				t.Fatalf("got (%d, %v), want (%d, nil)", got, err, result)
			}
			if !bytes.Equal(uart.writes.Bytes(), frame(msg)) {
				t.Fatalf("incorrect request %x", uart.writes.Bytes())
			}
		})
	}
}

func stringResult(result int32) string {
	switch result {
	case 0:
		return "zero"
	case 3:
		return "positive"
	case -1:
		return "closed"
	default:
		return "mbedTLS error"
	}
}

func TestInvalidResponse(t *testing.T) {
	msg := []byte{0, 6, 17, 1, 7, 0, 0, 0}
	valid := frame(response(msg, 0))
	badCRC := append([]byte(nil), valid...)
	badCRC[2] ^= 1
	cases := map[string][]byte{
		"no header":         nil,
		"partial header":    valid[:3],
		"partial body":      valid[:len(valid)-1],
		"bad CRC":           badCRC,
		"short message":     frame([]byte{2}),
		"oversized message": {1, 8, 0, 0},
		"zero length":       {0, 0, 0, 0},
	}
	for _, index := range []int{1, 2, 3, 4, 5, 6, 7} {
		data := response(msg, 0)
		data[index]++
		cases["wrong field "+string(rune('0'+index))] = frame(data)
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			uart := &transport{Reader: bytes.NewReader(data)}
			r := &rtl8720dn{rpcState: rpcState{uart: uart}}
			if err := r.request(msg); err == nil {
				t.Fatal("invalid response accepted")
			}
		})
	}
}

func TestWriteFailure(t *testing.T) {
	msg := []byte{0, 6, 17, 1, 7, 0, 0, 0}
	writeError := errors.New("UART write failed")
	for _, call := range []int{1, 2} {
		for _, short := range []bool{false, true} {
			uart := &transport{Reader: bytes.NewReader(frame(response(msg, 0))), failWrite: call, shortWrite: short}
			want := error(io.ErrShortWrite)
			if !short {
				uart.writeError = writeError
				want = writeError
			}
			r := &rtl8720dn{rpcState: rpcState{uart: uart}}
			if err := r.request(msg); !errors.Is(err, want) {
				t.Fatalf("got %v, want %v", err, want)
			}
		}
	}
}

func TestNotificationBeforeResponse(t *testing.T) {
	msg := []byte{0, 6, 17, 1, 7, 0, 0, 0}
	notification := append(append([]byte(nil), msg...), bytes.Repeat([]byte{0xa5}, 600)...)
	notification[0] = 3
	data := append(frame(notification), frame(response(msg, 0))...)
	uart := &transport{Reader: bytes.NewReader(data)}
	r := &rtl8720dn{rpcState: rpcState{uart: uart}}
	err := r.request(msg)
	got := int32(binary.LittleEndian.Uint32(payload[8:12]))
	if err != nil || got != 0 {
		t.Fatalf("got (%d, %v), want (0, nil)", got, err)
	}
}

func TestInt32ResponseLength(t *testing.T) {
	for _, size := range []int{8, 9, 11, 13} {
		msg := []byte{0, 6, 17, 1, 0, 0, 0, 0}
		data := append(response(msg, 0), 0)[:size]
		r := &rtl8720dn{rpcState: rpcState{uart: &transport{Reader: bytes.NewReader(frame(data))}}}
		if _, err := r.rpc_wifi_ssl_get_socket(42); err == nil {
			t.Fatalf("accepted result length %d", size)
		}
	}
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error)    { return 0, r.err }
func (r errorReader) Write(b []byte) (int, error) { return len(b), nil }

func TestReadError(t *testing.T) {
	want := errors.New("UART read failed")
	r := &rtl8720dn{rpcState: rpcState{uart: errorReader{want}}}
	if _, err := r.rpc_wifi_ssl_get_socket(42); !errors.Is(err, want) {
		t.Fatalf("got %v, want %v", err, want)
	}
}

func TestLegacyRequest(t *testing.T) {
	msg := []byte{0, 2, 1, 1, 9, 0, 0, 0, 7}
	notify := append([]byte(nil), msg[:8]...)
	notify[0] = 3
	uart := &transport{Reader: bytes.NewReader(append(frame(notify), frame(response(msg, 7))...))}
	r := &rtl8720dn{rpcState: rpcState{uart: uart, seq: 9}}
	if got := r.rpc_system_ack(7); got != 7 {
		t.Fatalf("got %d, want 7", got)
	}
	if !bytes.Equal(uart.writes.Bytes(), frame(msg)) || r.seq != 10 {
		t.Fatal("incorrect legacy request or sequence")
	}
}

func TestLegacyWriteError(t *testing.T) {
	uart := &transport{Reader: bytes.NewReader(nil), failWrite: 1, writeError: errors.New("write failed")}
	r := &rtl8720dn{rpcState: rpcState{uart: uart}}
	payload[8] = 7
	if got := r.rpc_system_ack(7); got != 0 || r.rpcErr == nil {
		t.Fatalf("got %d, error %v", got, r.rpcErr)
	}
}
