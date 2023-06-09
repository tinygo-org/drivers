package netlink

import (
	"net"
)

// Bridge connects links into an Ethernet (L2) broadcast domain
type Bridge struct {
	ip      net.IP
	links   []Netlinker
	recvEth func([]byte) error
}

func NewBridge(links []Netlinker) *Bridge {
	return &Bridge{links: links}
}

func (b *Bridge) NetConnect(params *ConnectParams) error {
	return nil
}

func (b *Bridge) NetDisconnect() {
}

func (b *Bridge) NetNotify(cb func(Event)) {
}

func (b *Bridge) GetHardwareAddr() (net.HardwareAddr, error) {
	return net.HardwareAddr{}, nil
}

func (b *Bridge) SendEth(pkt []byte) error {
	// L2 Forward ethernet pkt to zero of more []links
	return nil
}

func (b *Bridge) RecvEthFunc(cb func(pkt []byte) error) {
	b.recvEth = cb
}
