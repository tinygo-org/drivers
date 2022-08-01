package rtl8720dn

import (
	"fmt"
	"io"
	"time"
)

const maxUartRecvSize = 128

type RTL8720DN struct {
	port  io.ReadWriter
	seq   uint64
	sema  chan bool
	debug bool

	connectionType ConnectionType
	socket         int32
	client         uint32
	length         int
	root_ca        *string
	udpInfo        [6]byte // Port: [2]byte + IP: [4]byte
}

type ConnectionType int

const (
	ConnectionTypeNone ConnectionType = iota
	ConnectionTypeTCP
	ConnectionTypeUDP
	ConnectionTypeTLS
)

func New(r io.ReadWriter) *RTL8720DN {
	ret := &RTL8720DN{
		port:  r,
		seq:   1,
		sema:  make(chan bool, 1),
		debug: false,
	}

	return ret
}

func (r *RTL8720DN) Rpc_tcpip_adapter_init_with_timeout(d time.Duration) (int32, error) {
	timeout := make(chan bool)
	go func() {
		time.Sleep(d)
		timeout <- true
	}()

	var ret int32
	var err error
	done := make(chan bool)
	go func() {
		ret, err = r.Rpc_tcpip_adapter_init()
		done <- true
	}()

	select {
	case <-timeout:
		return ret, fmt.Errorf("Rpc_tcpip_adapter_init: timeout")
	case <-done:
		if err != nil {
			return ret, err
		}
	}

	return ret, nil
}

func (r *RTL8720DN) SetSeq(s uint64) {
	r.seq = s
}

func (r *RTL8720DN) Debug(b bool) {
	r.debug = b
}

func (r *RTL8720DN) SetRootCA(s *string) {
	r.root_ca = s
}

func (r *RTL8720DN) Version() (string, error) {
	return r.Rpc_system_version()
}
