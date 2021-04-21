package frame

import (
	"strings"
	"testing"
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
