//go:build rtl8720dn_hosttest

package rtl8720dn

import "sync"

type rtl8720dn struct {
	rpcState
	mu      sync.Mutex
	sockets map[sock]*socket
}

var _debug debug = debugOff
