package frame

import (
	"encoding/binary"
	"testing"

	"tinygo.org/x/drivers/net"

	"tinygo.org/x/drivers/encoding/hex"
)

func TestTCPChecksum(t *testing.T) {
	var tests = []struct {
		tcpf     TCP
		pseudo   IP
		Data     []byte
		expected uint16
	}{
		{
			TCP{Source: 80, Destination: 44084, Seq: 2561, Ack: 1631906285, DataOffset: 5, Flags: 0x019, WindowSize: 1024},
			IP{Source: net.IP{192, 168, 1, 5}, Destination: net.IP{192, 168, 1, 112}, Protocol: 6, Version: 69},
			[]byte("HTTP/1.0 200 OK\r\nContent-Type: text/html\r\nPragma: no-cache\r\n\r\n<h2>..::TinyGo Rocks::..</h2>"),
			0x2a3e,
		},
		{
			TCP{Source: 80, Destination: 44984, Seq: 2561, Ack: 3511653306, DataOffset: 5, Flags: 0x019, WindowSize: 1024},
			IP{Source: net.IP{192, 168, 1, 5}, Destination: net.IP{192, 168, 1, 112}, Protocol: 6, Version: 69},
			[]byte("HTTP/1.0 200 OK\r\nContent-Type: text/html\r\nPragma: no-cache\r\n\r\n<h2>..::TinyGo Rocks::..</h2>"),
			0x0ce2, // getting 0x0cb2
		},
	}

	for _, test := range tests {
		tcpBuff := make([]byte, 400)
		tcpf := &test.tcpf
		tcpf.Data = test.Data
		tcpf.PseudoHeaderInfo = &test.pseudo
		plen, err := tcpf.MarshalFrame(tcpBuff)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		got := binary.BigEndian.Uint16(tcpBuff[16:18])
		tcpBuff[16] = 0
		tcpBuff[17] = 0
		hgot := checksumRFC791(tcpBuff[:20])
		pgot := checksumRFC791(tcpBuff[:plen])
		tgot := checksumRFC791(tcpBuff[:])
		if got != test.expected {
			t.Errorf("expected %#x checksum. got %#x. buff: %s", test.expected, got, hex.Bytes(tcpBuff[:plen]))
			t.Errorf("hbuff:%#x pack:%#x total:%#x", hgot, pgot, tgot)
			if pgot != tgot {
				t.Errorf("checksum mismatch: tbuff:%x", tcpBuff)
			}
		}
	}
}

var allFlags = []uint16{
	TCPHEADER_FLAG_FIN, TCPHEADER_FLAG_SYN, TCPHEADER_FLAG_RST, TCPHEADER_FLAG_PSH,
	TCPHEADER_FLAG_ACK, TCPHEADER_FLAG_URG, TCPHEADER_FLAG_ECE, TCPHEADER_FLAG_CWR, TCPHEADER_FLAG_NS}

func TestTCPHasFlags(t *testing.T) {
	var tests = []struct {
		flags []uint16
	}{
		{[]uint16{TCPHEADER_FLAG_ACK, TCPHEADER_FLAG_FIN}},
		{[]uint16{TCPHEADER_FLAG_FIN, TCPHEADER_FLAG_PSH}},
		{[]uint16{TCPHEADER_FLAG_ACK, TCPHEADER_FLAG_PSH, TCPHEADER_FLAG_FIN}},
	}
	for _, test := range tests {
		tcp := TCP{}
		orflags := or16(test.flags...)
		tcp.SetFlags(orflags)
		for i := range allFlags {
			expect := contains(test.flags, allFlags[i])
			got := tcp.HasFlags(allFlags[i])
			if expect != got {
				t.Errorf("expected {%#x in %#x}=%t. %v", allFlags[i], test.flags, expect, tcp.StringFlags())
			}
		}
		if !tcp.HasFlags(orflags) {
			t.Error("tcp header contains flags it was originally created with")
		}
	}
}
func contains(s []uint16, v uint16) bool {
	for i := range s {
		if s[i] == v {
			return true
		}
	}
	return false
}
func or16(bit ...uint16) (res uint16) {
	for i := range bit {
		res |= bit[i]
	}
	return
}
