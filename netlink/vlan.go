package netlink

import (
	"net"
)

// Vlan is a virtual LAN
type Vlan struct {
	ip net.IP
	// Vlan ID
	id      uint16
	link    Netlinker
	recvEth func([]byte) error
}

func NewVlan(id uint16, link Netlinker) *Vlan {
	return &Vlan{id: id, link: link}
}

func (v *Vlan) NetConnect(params *ConnectParams) error {
	return nil
}

func (v *Vlan) NetDisconnect() {
}

func (v *Vlan) NetNotify(cb func(Event)) {
}

func (v *Vlan) GetHardwareAddr() (net.HardwareAddr, error) {
	return net.HardwareAddr{}, nil
}

func (v *Vlan) SendEth(pkt []byte) error {
	// Prepend VLAN hdr to pkt and send on link
	return nil
}

func (v *Vlan) RecvEthHandle(handler func(pkt []byte) error) {
	v.recvEth = handler
}
