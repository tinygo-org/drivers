// +build noheap

package enc28j60

// SDB enables serial print debugging of enc28j60 library
var SDB bool

// debug serial print. If SDB is set to false then it is not compiled unless compiler cannot determine
// SDB does not change
func dbp(msg string, datas ...[]byte) {}
