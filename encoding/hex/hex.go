package hex

// Byte converts a single byte to an ASCII
// byte slice representation.
//
// Example:
//  string(hex.Byte(0xff))
//  Output: "ff"
func Byte(b byte) []byte {
	var res [2]byte
	res[0], res[1] = (b>>4)+'0', (b&0b0000_1111)+'0'
	if (b >> 4) > 9 {
		res[0] = (b >> 4) + 'A' - 10
	}
	if (b & 0b0000_1111) > 9 {
		res[1] = (b & 0b0000_1111) + 'A' - 10
	}
	return res[:]
}

// Bytes converts a binary slice of bytes to an ASCII
// hex representation.
//
// Example:
//  string(hex.Bytes([]byte{0xff,0xaa}))
//  Output: "ffaa"
func Bytes(b []byte) []byte {
	o := make([]byte, len(b)*2)
	for i := 0; i < len(b); i++ {
		aux := Byte(b[i])
		o[i*2] = aux[0]
		o[i*2+1] = aux[1]
	}
	return o
}
