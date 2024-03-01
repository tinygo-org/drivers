package stusb4500

// lendU32 returns a uint32 from the given slice of bytes.
// the first element (0) in the given slice becomes the least-significant byte;
// this can be interpreted as a "little-endian" conversion, hence the name.
// if less than 4 bytes are given (len(word) < 4), the missing most-significant
// bytes are set to 0.
// any additional bytes trailing the first 4 bytes are ignored.
func lendU32(word ...uint8) uint32 {
	numWords := 4 // read at most 4 bytes to form a 32-bit object
	if len(word) < numWords {
		numWords = len(word)
	}
	var data uint32
	for i, j := 0, 0; i < numWords; i, j = i+1, j+8 {
		data |= uint32(word[i]) << j
	}
	return data
}

// lendU16 returns a uint16 from the given slice of bytes.
// the first element (0) in the given slice becomes the least-significant byte;
// this can be interpreted as a "little-endian" conversion, hence the name.
// if less than 2 bytes are given (len(word) < 2), the missing most-significant
// bytes are set to 0.
// any additional bytes trailing the first 2 bytes are ignored.
func lendU16(word ...uint8) uint16 {
	numWords := 2 // read at most 2 bytes to form a 16-bit object
	if len(word) < numWords {
		numWords = len(word)
	}
	var data uint16
	for i, j := 0, 0; i < numWords; i, j = i+1, j+8 {
		data |= uint16(word[i]) << j
	}
	return data
}

// lendU8 returns a uint8 from the given slice of bytes.
// the first element (0) in the given slice becomes the resulting byte returned;
// this can be interpreted as a "little-endian" conversion, hence the name.
// if 0 bytes are given (len(word) == 0), the zero byte is returned.
// any additional bytes trailing the first byte are ignored.
func lendU8(word ...uint8) uint8 {
	if len(word) > 0 {
		return word[0]
	}
	return 0
}

func bytes32(data uint32) []byte {
	b := make([]byte, 4)
	for i := range b {
		b[i] = byte((data >> (8 * i)) & 0xFF)
	}
	return b
}

// boolToByte returns 1 if and only if set is true, otherwise 0.
func boolToByte(set bool) byte {
	if set {
		return 1
	}
	return 0
}
