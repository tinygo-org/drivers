package lorawan

// revertByteArray inverts de order of a given byte slice
func revertByteArray(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// byteToHex return string hex representation of byte
func byteToHex(b byte) string {
	bb := (b >> 4) & 0x0F
	ret := ""
	if bb < 10 {
		ret += string(rune('0' + bb))
	} else {
		ret += string(rune('A' + (bb - 10)))
	}

	bb = (b) & 0xF
	if bb < 10 {
		ret += string(rune('0' + bb))
	} else {
		ret += string(rune('A' + (bb - 10)))
	}
	return ret
}

// BytesToHexString converts byte slice to hex string representation
func bytesToHexString(data []byte) string {
	s := ""
	for i := 0; i < len(data); i++ {
		s += byteToHex(data[i])
	}
	return s
}
