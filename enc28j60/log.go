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
			print(" 0x" + string(hex.Bytes(datas[d])))
			// for i := 0; i < len(datas[d]); i++ {
			// 	print(string(hex.Byte(datas[d][i])))
			// }
		}
		println()
	}
}
