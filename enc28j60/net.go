package enc28j60

import "github.com/jkaflik/tinygo-w5500-driver/wiznet/net"

// SetIPAddress Sets device static IP Address
func (d *Dev) SetIPAddress(ip net.IP) { copy(d.myip, ip) }

// SetGatewayAdress Sets router/gateway address. where requests outside the network will come from
func (d *Dev) SetGatewayAdress(ip net.IP) { copy(d.broadcastip, ip) }

//SetSubnetMask sets the subnet mask for the device
func (d *Dev) SetSubnetMask(mask net.IPMask) { copy(d.mask, mask) }

// NewSocket instances a socket to write to buffer
func (d *Dev) NewSocket() Socket {
	return Socket{
		Num: 0,
		d:   d,
	}
}
