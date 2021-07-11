// +build !noheap

package enc28j60

import "github.com/soypat/ether-swtch/hex"

// SDB enables serial print debugging of enc28j60 library
var SDB bool

// debug serial print. If SDB is set to false then it is not compiled unless compiler cannot determine
// SDB does not change
func dbp(msg string, datas ...[]byte) {
	if SDB {
		print(msg)
		for d := range datas {
			print(" 0x")
			print(string(hex.Bytes(datas[d])))
		}
		println()
	}
}