package rtl8720dn

import (
	"machine"

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

func (d *Driver) SetSeq(s uint64) {
	d.seq = s
}

func (d *Driver) Debug(b bool) {
	d.debug = b
}

func (d *Driver) SetRootCA(s *string) {
	d.root_ca = s
}

func (d *Driver) Version() (string, error) {
	return d.Rpc_system_version()
}

func enable(en machine.Pin) {
	en.Configure(machine.PinConfig{Mode: machine.PinOutput})
	en.Low()
	time.Sleep(100 * time.Millisecond)
	en.High()
	time.Sleep(1000 * time.Millisecond)
}

type UARTx struct {
	*machine.UART
}

func (u *UARTx) Read(p []byte) (n int, err error) {
	if u.Buffered() == 0 {
		time.Sleep(1 * time.Millisecond)
		return 0, nil
	}
	return u.UART.Read(p)
}
