package frame

import (
	"encoding/binary"
	"strings"
	"testing"

	"tinygo.org/x/drivers/encoding/hex"
	"tinygo.org/x/drivers/net2"
)

func TestChecksum(t *testing.T) {

	var tests = []struct {
		hexdata  string
		expected uint16
	}{
		//         17664,115,0,16384, 16401, 49320, 1, 49320, 199 // decimal
		//            17779 -, 34163, 50564, 99884,99885, 149205, 149404 // cumulative sum
		{hexdata: "4500 0073 0000 4000 4011  c0a8 0001 c0a8 00c7", expected: 0xB861},    // from https://en.wikipedia.org/wiki/IPv4_header_checksum
		{hexdata: "4500 0073 0000 4000 4011  c0a8 0001 c0a8 00c7 00", expected: 0xB861}, // with padding
		{hexdata: "4500 0073 0000 4000 4011  c0a8 0001 c0a8 00c7 0000", expected: 0xB861},
	}

	for _, test := range tests {
		buff := hexStringToBytes(test.hexdata)
		got := checksumRFC791(buff)
		if got != test.expected {
			t.Errorf("got %#x. expected %#x for data: %#x", got, test.expected, buff)
		}
	}
}

func TestTCPChecksum(t *testing.T) {
	var tests = []struct {
		tcpf     TCP
		pseudo   IP
		Data     []byte
		expected uint16
	}{
		{
			TCP{Source: 80, Destination: 44084, Seq: 2561, Ack: 1631906285, DataOffset: 5, Flags: 0x019, WindowSize: 1024},
			IP{Source: net2.IP{192, 168, 1, 5}, Destination: net2.IP{192, 168, 1, 112}, Protocol: 6, Version: 69},
			[]byte("HTTP/1.0 200 OK\r\nContent-Type: text/html\r\nPragma: no-cache\r\n\r\n<h2>..::TinyGo Rocks::..</h2>"),
			0x2a3e,
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

func hexStringToBytes(hexes string) []byte {
	const hexString = "0123456789ABCDEF"
	var hx int // hexes processed in current byte (need 2 to form a byte)
	var currentByte byte
	hexes = strings.ToUpper(hexes)
	buff := make([]byte, 0)
	for _, v := range hexes {
		skipFlag := false // skip non hex runes
		var val uint8
		for i, x := range hexString {
			if v == x {
				val = uint8(i)
				break
			} else if i == len(hexString)-1 {
				skipFlag = true
			}
		}
		if skipFlag {
			continue
		}
		hx++
		switch {
		case hx == 1:
			currentByte = val << 4
		case hx == 2:
			currentByte += val
			buff = append(buff, currentByte)
			currentByte = 0
			hx = 0
		}
	}
	return buff
}
