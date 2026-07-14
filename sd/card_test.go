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

// TestCSDv2Capacity checks CSDv2 C_SIZE decoding and the capacity formula.
//
// Field layout and formula are from the SD Physical Layer Simplified
// Specification Version 9.10, section 5.3.3 "CSD Register (CSD Version 2.0)":
// C_SIZE occupies CSD bits [69:48] and user memory capacity is
//
//	memory capacity = (C_SIZE+1) * 512KByte  (512KByte = 524288 bytes)
//
// Specification download: https://www.sdcard.org/downloads/pls/
//
// Regression test for two past bugs in CSDv2:
//   - csize() read byte 7 as data[7]>>2 instead of data[7]&0x3F, dropping
//     C_SIZE bits [17:16] (misdecoded cards > 32GiB).
//   - DeviceCapacity() computed csize*512000 instead of (csize+1)*524288.
func TestCSDv2Capacity(t *testing.T) {
	tests := []struct {
		name      string
		csd       []byte
		wantCSize uint32
		wantCap   int64
	}{
		{
			// Same 8GB-card register as in TestCRC7 above, CRC byte appended
			// (CRC7=0b1110010 per that test, stored as crc<<1|always1).
			// C_SIZE = 0x003C01 = 15361 -> 15362 * 524288 = 8054112256 bytes.
			name:      "8GB card (in-repo vector)",
			csd:       []byte{64, 14, 0, 50, 83, 89, 0, 0, 60, 1, 127, 128, 10, 64, 0, 0b1110010<<1 | 1},
			wantCSize: 0x003C01,
			wantCap:   8054112256,
		},
		// {
		// 	name:      "8GB card",
		// 	csd:       csdv2Bytes(0x003C01),
		// 	wantCSize: 0x003FFF,
		// 	wantCap:   2 << 40,
		// },
		{
			// Synthetic: C_SIZE = 0x01FFFF -> 131072 * 524288 = 64GiB.
			// Exercises C_SIZE bits [17:16], stored in CSD byte 7 (CSD bits [65:64]).
			name:      "64GiB synthetic",
			csd:       csdv2Bytes(0x01FFFF),
			wantCSize: 0x01FFFF,
			wantCap:   64 << 30,
		},
		{
			// Synthetic: maximum v2 C_SIZE = 0x3FFFFF -> 4194304 * 524288 = 2TiB,
			// the SDXC upper bound per section 5.3.3.
			name:      "2TiB synthetic (max C_SIZE)",
			csd:       csdv2Bytes(0x3FFFFF),
			wantCSize: 0x3FFFFF,
			wantCap:   2 << 40,
		},
	}
	for _, tt := range tests {
		csd, err := DecodeCSD(tt.csd)
		if err != nil {
			t.Fatalf("%s: DecodeCSD: %v", tt.name, err)
		}
		v2 := csd.MustV2()
		if got := v2.csize(); got != tt.wantCSize {
			t.Errorf("%s: csize() = %#x, want %#x", tt.name, got, tt.wantCSize)
		}
		if got := csd.DeviceCapacity(); got != tt.wantCap {
			t.Errorf("%s: DeviceCapacity() = %d, want %d", tt.name, got, tt.wantCap)
		}
		wantBlocks := tt.wantCap / int64(csd.ReadBlockLen())
		if got := csd.NumberOfBlocks(); got != wantBlocks {
			t.Errorf("%s: NumberOfBlocks() = %d, want %d", tt.name, got, wantBlocks)
		}
	}
}

// csdv2Bytes returns a 16-byte CSD v2 register with csize spliced into
// CSD bits [69:48] and a freshly computed CRC7+always1 last byte.
// Non-capacity fields are copied from the 8GB-card vector above.
func csdv2Bytes(csize uint32) []byte {
	b := []byte{64, 14, 0, 50, 83, 89, 0, 0, 60, 1, 127, 128, 10, 64, 0}
	b[7] = byte(csize >> 16 & 0x3F)
	b[8] = byte(csize >> 8)
	b[9] = byte(csize)
	return append(b, crc7noshift(b)|1)
}

func putCmd(dst []byte, cmd command, arg uint32) {
	dst[0] = byte(cmd) | (1 << 6)
	dst[1] = byte(arg >> 24)
	dst[2] = byte(arg >> 16)
	dst[3] = byte(arg >> 8)
	dst[4] = byte(arg)
	dst[5] = crc7noshift(dst[:5]) | 1 // Stop bit added.
}
