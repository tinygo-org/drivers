//go:build rtl8720dn_hosttest

package rtl8720dn

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"strings"
	"testing"

	"tinygo.org/x/drivers/netdev"
)

func tlsResponse(method byte, seq uint32, result *int32) []byte {
	data := []byte{2, method, 17, 1, 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(data[4:], seq)
	if result != nil {
		data = binary.LittleEndian.AppendUint32(data, uint32(*result))
	}
	return frame(data)
}

func TestTLSConnect(t *testing.T) {
	for _, tc := range []struct {
		name           string
		result, socket int32
		want           string
	}{
		{"positive", 3, 3, ""},
		{"zero socket", 0, 0, ""},
		{"minus one", -1, 0, "code -1"},
		{"firmware error", -29312, 0, "code -29312"},
		{"closed socket", 0, -1, "socket -1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := tlsResponse(18, 0, &tc.result)
			calls := 1
			if tc.result >= 0 {
				data = append(data, tlsResponse(6, 1, &tc.socket)...)
				calls++
			}
			data = append(data, tlsResponse(19, uint32(calls), nil)...)
			data = append(data, tlsResponse(2, uint32(calls+1), nil)...)
			uart := &transport{Reader: bytes.NewReader(data)}
			skt := newSocket(netdev.IPPROTO_TLS)
			r := &rtl8720dn{rpcState: rpcState{uart: uart}, sockets: map[sock]*socket{42: skt}}
			err := r.Connect(42, "example.org", netip.MustParseAddrPort("192.0.2.1:443"))
			if tc.want == "" {
				if err != nil || skt.connectFailed {
					t.Fatalf("got %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tc.want) || !skt.connectFailed {
					t.Fatalf("got %v, want %s", err, tc.want)
				}
			}
			if r.seq != uint64(calls) {
				t.Fatalf("wrong RPC count %d", r.seq)
			}
			if err := r.Close(42); err != nil {
				t.Fatal(err)
			}
			if skt.inuse || len(r.sockets) != 0 || uart.Len() != 0 {
				t.Fatal("TLS resources remain")
			}
			writes := uart.writes.Len()
			if err := r.Close(42); err != nil || uart.writes.Len() != writes {
				t.Fatal("duplicate close")
			}
			wire := uart.writes.Bytes()
			var methods []byte
			for len(wire) > 0 {
				n := int(binary.LittleEndian.Uint16(wire[:2]))
				methods = append(methods, wire[5])
				wire = wire[4+n:]
			}
			wantMethods := []byte{18, 19, 2}
			if calls == 2 {
				wantMethods = []byte{18, 6, 19, 2}
			}
			if !bytes.Equal(methods, wantMethods) {
				t.Fatalf("methods %v", methods)
			}
		})
	}
}

func TestTLSConnectRPCError(t *testing.T) {
	zero := int32(0)
	for _, stage := range []int{0, 1} {
		data := []byte(nil)
		if stage == 1 {
			data = tlsResponse(18, 0, &zero)
		}
		uart := &transport{Reader: bytes.NewReader(data)}
		r := &rtl8720dn{rpcState: rpcState{uart: uart}, sockets: map[sock]*socket{42: newSocket(netdev.IPPROTO_TLS)}}
		err := r.Connect(42, "example.org", netip.MustParseAddrPort("192.0.2.1:443"))
		if !errors.Is(err, io.EOF) || !r.sockets[42].connectFailed {
			t.Fatalf("got %v", err)
		}
		uart.Reader = bytes.NewReader(append(tlsResponse(19, uint32(r.seq), nil), tlsResponse(2, uint32(r.seq+1), nil)...))
		if err := r.Close(42); err != nil || len(r.sockets) != 0 {
			t.Fatalf("close failed %v", err)
		}
	}
}

func TestTLSConnectBadResponse(t *testing.T) {
	zero := int32(0)
	for _, stage := range []int{0, 1} {
		for _, failure := range []string{"CRC", "header", "write", "short write"} {
			t.Run(fmt.Sprintf("%d %s", stage, failure), func(t *testing.T) {
				method := byte(18)
				data := []byte(nil)
				if stage == 1 {
					data = tlsResponse(18, 0, &zero)
					method = 6
				}
				bad := tlsResponse(method, uint32(stage), &zero)
				switch failure {
				case "CRC":
					bad[2] ^= 1
				case "header":
					bad[5]++
					binary.LittleEndian.PutUint16(bad[2:4], referenceCRC16(bad[4:]))
				}
				uart := &transport{Reader: bytes.NewReader(append(data, bad...))}
				want := errors.New("UART write failed")
				if failure == "write" {
					uart.failWrite = 2*stage + 1
					uart.writeError = want
				} else if failure == "short write" {
					uart.failWrite = 2*stage + 2
					uart.shortWrite = true
				}
				r := &rtl8720dn{rpcState: rpcState{uart: uart}, sockets: map[sock]*socket{42: newSocket(netdev.IPPROTO_TLS)}}
				err := r.Connect(42, "example.org", netip.MustParseAddrPort("192.0.2.1:443"))
				if err == nil || !r.sockets[42].connectFailed {
					t.Fatalf("got %v", err)
				}
				if failure == "write" && !errors.Is(err, want) {
					t.Fatalf("lost UART error %v", err)
				}
				if failure == "short write" && !errors.Is(err, io.ErrShortWrite) {
					t.Fatalf("lost short write %v", err)
				}
				uart.failWrite = 0
				uart.Reader = bytes.NewReader(append(tlsResponse(19, uint32(r.seq), nil), tlsResponse(2, uint32(r.seq+1), nil)...))
				if err := r.Close(42); err != nil || len(r.sockets) != 0 {
					t.Fatalf("close failed %v", err)
				}
			})
		}
	}
}

func TestPlainConnect(t *testing.T) {
	for _, protocol := range []int{netdev.IPPROTO_TCP, netdev.IPPROTO_UDP} {
		for _, result := range []int32{0, -1} {
			data := []byte{2, 9, 16, 1, 0, 0, 0, 0}
			data = binary.LittleEndian.AppendUint32(data, uint32(result))
			uart := &transport{Reader: bytes.NewReader(frame(data))}
			r := &rtl8720dn{rpcState: rpcState{uart: uart}, sockets: map[sock]*socket{42: newSocket(protocol)}}
			err := r.Connect(42, "", netip.MustParseAddrPort("192.0.2.1:443"))
			if (err != nil) != (result == -1) {
				t.Fatalf("got %v for %d", err, result)
			}
			if r.seq != 1 {
				t.Fatalf("wrong RPC count %d", r.seq)
			}
		}
	}
}
