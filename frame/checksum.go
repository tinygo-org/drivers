package frame

import "encoding/binary"

// Checksum function as defined by RFC 791. The checksumRFC791 field
// is the 16-bit ones' complement of the ones' complement sum of
// all 16-bit words in the header. For purposes of computing the checksumRFC791,
// the value of the checksumRFC791 field is zero.
func checksumRFC791(data []byte) uint16 {
	var sum uint32
	n := len(data) / 2
	// automatic padding of data
	if len(data)%2 != 0 {
		sum += uint32(data[len(data)-1]) << 8
	}
	for i := 0; i < n; i++ {
		sum += uint32(binary.BigEndian.Uint16(data[i*2 : i*2+2]))
	}
	for sum > 0xffff {
		sum = sum&0xffff + sum>>16
	}
	return ^uint16(sum)
}
