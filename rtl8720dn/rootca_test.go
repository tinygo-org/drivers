//go:build rtl8720dn_hosttest

package rtl8720dn

import (
	"bytes"
	"encoding/binary"
	"machine"
	"net/netip"
	"strings"
	"testing"

	"tinygo.org/x/drivers/netdev"
)

type caReply struct {
	method byte
	result uint32
}

func caDevice(replies ...caReply) (*rtl8720dn, *bytes.Buffer) {
	var input bytes.Buffer
	for seq, reply := range replies {
		msg := make([]byte, 8)
		copy(msg, []byte{0, reply.method, 0x11, 1})
		binary.LittleEndian.PutUint32(msg[4:], uint32(seq))
		input.Write(tlsReply(msg, int32(reply.result)))
	}
	output := new(bytes.Buffer)
	r := New(&Config{})
	r.uart = &machine.UART{Reader: &input, Writer: output}
	return r, output
}

func caRequests(t *testing.T, output *bytes.Buffer) [][]byte {
	t.Helper()
	var requests [][]byte
	data := output.Bytes()
	for len(data) != 0 {
		if len(data) < 4 {
			t.Fatal("Short request header")
		}
		n := int(binary.LittleEndian.Uint16(data))
		if n < 8 || len(data) < n+4 {
			t.Fatal("Invalid request length")
		}
		msg := data[4 : n+4]
		if computeCRC16(msg) != binary.LittleEndian.Uint16(data[2:]) {
			t.Fatal("Invalid request CRC")
		}
		if msg[0] != 0 || msg[2] != 0x11 || msg[3] != 1 || binary.LittleEndian.Uint32(msg[4:]) != uint32(len(requests)) {
			t.Fatal("Invalid request header")
		}
		requests = append(requests, msg)
		data = data[n+4:]
	}
	return requests
}

func caSocket(t *testing.T, r *rtl8720dn, want int) {
	t.Helper()
	fd, err := r.Socket(netdev.AF_INET, netdev.SOCK_STREAM, netdev.IPPROTO_TLS)
	if err != nil || fd != want {
		t.Fatalf("Socket = %d, %v; want %d", fd, err, want)
	}
}

func caMethods(t *testing.T, requests [][]byte, want []byte) {
	t.Helper()
	var got []byte
	for _, request := range requests {
		got = append(got, request[1])
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("Methods = %v; want %v", got, want)
	}
}

func caData(t *testing.T, request []byte, client uint32, pem string) {
	t.Helper()
	if len(request) != 16+len(pem) || binary.LittleEndian.Uint32(request[8:]) != client ||
		binary.LittleEndian.Uint32(request[12:]) != uint32(len(pem)) || string(request[16:]) != pem {
		t.Fatal("CA request does not match the client and PEM data")
	}
}

func TestRootCAOrderingAndEncoding(t *testing.T) {
	for _, result := range []uint32{0, 0xffffffff} {
		r, output := caDevice(caReply{1, 100}, caReply{3, 0}, caReply{5, 0}, caReply{8, result}, caReply{0x12, 7}, caReply{6, 7})
		pem := "-----BEGIN CERTIFICATE-----\nfirst\n-----END CERTIFICATE-----\n" +
			"-----BEGIN CERTIFICATE-----\nsecond\n-----END CERTIFICATE-----\n"
		if err := r.SetRootCA(pem); err != nil {
			t.Fatal(err)
		}
		if output.Len() != 0 {
			t.Fatal("Configuration sent an RPC request")
		}
		caSocket(t, r, 100)
		if err := r.Connect(100, "test.example", netip.MustParseAddrPort("192.0.2.1:443")); err != nil {
			t.Fatal(err)
		}
		requests := caRequests(t, output)
		caMethods(t, requests, []byte{1, 3, 5, 8, 0x12, 6})
		caData(t, requests[3], 100, pem)
		for _, request := range requests[1:] {
			if binary.LittleEndian.Uint32(request[8:]) != 100 {
				t.Fatal("Request uses the wrong client")
			}
		}
	}
}

func TestRootCAEmpty(t *testing.T) {
	for _, clear := range []bool{false, true} {
		r, output := caDevice(caReply{1, 100}, caReply{3, 0}, caReply{5, 0}, caReply{0x12, 7}, caReply{6, 7})
		if clear {
			if err := r.SetRootCA("old"); err != nil {
				t.Fatal(err)
			}
			if err := r.SetRootCA(""); err != nil {
				t.Fatal(err)
			}
		}
		caSocket(t, r, 100)
		if err := r.Connect(100, "test.example", netip.MustParseAddrPort("192.0.2.1:443")); err != nil {
			t.Fatal(err)
		}
		caMethods(t, caRequests(t, output), []byte{1, 3, 5, 0x12, 6})
	}
}

func TestRootCASizeAndNUL(t *testing.T) {
	r, output := caDevice(caReply{1, 100}, caReply{3, 0}, caReply{5, 0}, caReply{8, 0})
	pem := strings.Repeat("x", 4080)
	if err := r.SetRootCA(pem); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []string{pem + "x", strings.Repeat("x", 65536), "before\x00after", strings.Repeat("é", 2041)} {
		if err := r.SetRootCA(invalid); err == nil {
			t.Fatal("Invalid configuration was accepted")
		}
		if r.rootCA != pem || output.Len() != 0 {
			t.Fatal("Invalid configuration changed or transmitted the CA data")
		}
	}
	caSocket(t, r, 100)
	requests := caRequests(t, output)
	caMethods(t, requests, []byte{1, 3, 5, 8})
	caData(t, requests[3], 100, pem)
	if len(requests[3]) != 4096 {
		t.Fatal("Maximum CA request does not have 4096 bytes")
	}
}

func TestRootCAExistingAndNewClients(t *testing.T) {
	r, output := caDevice(
		caReply{1, 100}, caReply{3, 0}, caReply{5, 0}, caReply{8, 0},
		caReply{1, 200}, caReply{3, 0}, caReply{5, 0}, caReply{8, 0},
		caReply{0x12, 7}, caReply{6, 7}, caReply{0x12, 8}, caReply{6, 8},
		caReply{1, 300}, caReply{3, 0}, caReply{5, 0},
	)
	if err := r.SetRootCA("first"); err != nil {
		t.Fatal(err)
	}
	caSocket(t, r, 100)
	if err := r.SetRootCA("second"); err != nil {
		t.Fatal(err)
	}
	caSocket(t, r, 200)
	if err := r.SetRootCA(""); err != nil {
		t.Fatal(err)
	}
	for _, fd := range []int{100, 200} {
		if err := r.Connect(fd, "test.example", netip.MustParseAddrPort("192.0.2.1:443")); err != nil {
			t.Fatal(err)
		}
	}
	caSocket(t, r, 300)
	requests := caRequests(t, output)
	caMethods(t, requests, []byte{1, 3, 5, 8, 1, 3, 5, 8, 0x12, 6, 0x12, 6, 1, 3, 5})
	caData(t, requests[3], 100, "first")
	caData(t, requests[7], 200, "second")
}
