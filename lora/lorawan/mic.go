package lorawan

import (
	"encoding/binary"
)

// genPayloadMIC computes MIC given the payload and the key
func genPayloadMIC(payload []uint8, key [16]uint8) [4]uint8 {
	var mic [4]uint8
	hash, _ := NewCmac(key[:])
	hash.Write(payload)
	hb := hash.Sum([]byte{})
	copy(mic[:], hb[0:4])
	return mic
}

func calcMessageMIC(payload []uint8, key [16]uint8, dir uint8, addr []byte, fCnt uint32, lenMessage uint8) [4]uint8 {
	var b0 []byte
	b0 = append(b0, 0x49, 0x00, 0x00, 0x00, 0x00)
	b0 = append(b0, dir)
	b0 = append(b0, addr[:]...)
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], fCnt)
	b0 = append(b0, b[:]...)
	b0 = append(b0, 0x00)
	b0 = append(b0, lenMessage)

	var full []byte
	full = append(full, b0...)
	full = append(full, payload...)

	var mic [4]uint8
	hash, _ := NewCmac(key[:])
	hash.Write(full)
	hb := hash.Sum([]byte{})
	copy(mic[:], hb[0:4])
	return mic
}
