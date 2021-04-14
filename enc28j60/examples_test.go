package enc28j60_test

import (
	"machine"

	"github.com/jkaflik/tinygo-w5500-driver/wiznet/net"
	"tinygo.org/x/drivers/enc28j60"
)

func ExampleSocket_aRP() {
	// best declared as a global variable for tinygo application
	var buff [1000]byte
	// Machine-specific configuration
	// use pin D0 as output
	// 8MHz SPI clk
	e, err := enc28j60.New(machine.D10, machine.SPI0, 8e6)
	if err != nil {
		println(err)
	}
	err = e.Init(buff[:], net.HardwareAddr{0xfe, 0xfe, 0xfe, 0x22, 0x22, 0x22})
	if err != nil {
		println(err)
	}
	// Set network specific Address
	e.SetGatewayAddress(net.IPv4(192, 168, 1, 1))
	e.SetIPAddress(net.IPv4(192, 168, 1, 5))
	e.SetSubnetMask(net.IPv4Mask(255, 255, 255, 0))
	s := e.NewSocket()
	// 0 makes a random port
	err = s.Open("arp", 0)
	if err != nil {
		println(err)
	}
	// ARP protocol does not support custom payload
	// we just wait for the destination to resolve our request
	gwHWAddr, err := s.Resolve()
	if err != nil {
		println(err)
	}
	// do something with gateway address
	println(gwHWAddr)
}
