// +build !noheap

package enc28j60

// SDB enables serial print debugging of enc28j60 library
var SDB bool

// debug serial print. If SDB is set to false then it is not compiled unless compiler cannot determine
// SDB does not change
func dbp(msg string, datas ...[]byte) {
	if SDB {
		print(msg)
		for d := range datas {
			print(" 0x")
			print(string(Bytes(datas[d])))
		}
		println()
	}
}

// Byte converts a single byte to an ASCII
// byte slice representation.
//
// Example:
//  string(hex.Byte(0xff))
//  Output: "ff"
func Byte(b byte) []byte {
	var res [2]byte
	if (b >> 4) > 9 {
		res[0] = (b >> 4) + 'A' - 10
	} else {
		res[0] = (b >> 4) + '0'
	}
	if (b & 0b0000_1111) > 9 {
		res[1] = (b & 0b0000_1111) + 'A' - 10
	} else {
		res[1] = (b & 0b0000_1111) + '0'
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
