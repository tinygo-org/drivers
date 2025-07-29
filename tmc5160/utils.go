package tmc5160

func ToHex(value uint32) string {
	hexChars := "0123456789ABCDEF"
	result := ""

	for i := 0; i < 8; i++ { // 8 nibbles for a 32-bit number
		nibble := (value >> (28 - i*4)) & 0xF
		result += string(hexChars[nibble])
	}

	return "0x" + result
}
