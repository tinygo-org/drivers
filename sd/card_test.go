package sd

import (
	"encoding/hex"
	"testing"
)

func TestCRC16(t *testing.T) {
	tests := []struct {
		block   string
		wantcrc uint16
	}{
		{
			block:   "fa33c08ed0bc007c8bf45007501ffbfcbf0006b90001f2a5ea1d060000bebe07b304803c80740e803c00751c83c610fecb75efcd188b148b4c028bee83c610fecb741a803c0074f4be8b06ac3c00740b56bb0700b40ecd105eebf0ebfebf0500bb007cb8010257cd135f730c33c0cd134f75edbea306ebd3bec206bffe7d813d55aa75c78bf5ea007c0000496e76616c696420706172746974696f6e207461626c65004572726f72206c6f6164696e67206f7065726174696e672073797374656d004d697373696e67206f7065726174696e672073797374656d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000094c6dffd0000000401040cfec2ff000800000000f00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000055aa",
			wantcrc: 0x52ce,
		},
	}

	for _, tt := range tests {
		b, err := hex.DecodeString(tt.block)
		if err != nil {
			t.Fatal(err)
		}
		gotcrc := CRC16(b)
		if gotcrc != tt.wantcrc {
			t.Errorf("calculateCRC(%s) = %#x, want %#x", tt.block, gotcrc, tt.wantcrc)
		}
	}
}

func TestCRC7(t *testing.T) {
	const cmdSendMask = 0x40
	tests := []struct {
		data    []byte
		wantCRC uint8
	}{
		{ // See CRC7 Examples from section 4.5 of the SD Card Physical Layer Simplified Specification.
			data:    []byte{cmdSendMask, 4: 0}, // CMD0, arg=0
			wantCRC: 0b1001010,
		},
		{
			data:    []byte{cmdSendMask | 17, 4: 0}, // CMD17, arg=0
			wantCRC: 0b0101010,
		},
		{
			data:    []byte{17, 3: 0b1001, 4: 0}, // Response of CMD17
			wantCRC: 0b0110011,
		},
		{ // CSD for a 8GB card.
			data:    []byte{64, 14, 0, 50, 83, 89, 0, 0, 60, 1, 127, 128, 10, 64, 0},
			wantCRC: 0b1110010,
		},
	}

	for _, tt := range tests {
		gotcrc := CRC7(tt.data[:])
		if gotcrc != tt.wantCRC {
			t.Errorf("got crc=%#b, want=%#b for %#b", gotcrc, tt.wantCRC, tt.data)
		}
	}

	cmdTests := []struct {
		cmd     command
		arg     uint32
		wantCRC uint8
	}{
		{
			cmd:     cmdGoIdleState,
			arg:     0,
			wantCRC: 0x95,
		},
		{
			cmd:     cmdSendIfCond,
			arg:     0x1AA,
			wantCRC: 0x87,
		},
	}
	var dst [6]byte
	for _, test := range cmdTests {
		putCmd(dst[:], test.cmd, test.arg)
		gotcrc := dst[5]
		if gotcrc != test.wantCRC {
			t.Errorf("got crc=%#x, want=%#x", gotcrc, test.wantCRC)
		}
	}
}

func putCmd(dst []byte, cmd command, arg uint32) {
	dst[0] = byte(cmd) | (1 << 6)
	dst[1] = byte(arg >> 24)
	dst[2] = byte(arg >> 16)
	dst[3] = byte(arg >> 8)
	dst[4] = byte(arg)
	dst[5] = crc7noshift(dst[:5]) | 1 // Stop bit added.
}
