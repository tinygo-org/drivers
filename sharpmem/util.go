package sharpmem

// setBit sets the bit at pos in n to 1 and returns the updated number.
func setBit(n uint8, pos uint8) uint8 {
	n |= 1 << pos
	return n
}

// unsetBit sets the bit at pos in n to 0 and returns the updated number.
func unsetBit(n uint8, pos uint8) uint8 {
	n &^= 1 << pos
	return n
}

// hasBit returns whether the bit at pos in n is 1.
func hasBit(n uint8, pos uint8) bool {
	n = n & (1 << pos)
	return n > 0
}

// bitfieldBufLen returns the required buffer size for keeping track of
// changed lines.
func bitfieldBufLen(bits int16) int16 {
	return 1 + (bits-1)/8
}

// ceilDiv divides a with b, but it uses the ceiling if modulo is not 0.
func ceilDiv(a, b int16) int16 {
	return 1 + (a-1)/b
}
