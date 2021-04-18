package enc28j60_test

import (
	"machine"

	"github.com/jkaflik/tinygo-w5500-driver/wiznet/net"
	"tinygo.org/x/drivers/enc28j60"
)

func ExampleEthernetChatter() {
	// best declared as a global variable for tinygo application
	var buff [1000]byte
	// Machine-specific configuration
	// use pin D0 as output
	// 8MHz SPI clk
	machine.SPI0.Configure(machine.SPIConfig{Frequency: 8e6})
	e := enc28j60.New(machine.D10, machine.SPI0)

	err := e.Init(buff[:], net.HardwareAddr{0xfe, 0xfe, 0xfe, 0x22, 0x22, 0x22})
	if err != nil {
		println(err.Error())
	}
	// Set network specific Address
	e.SetGatewayAddress(net.IPv4(192, 168, 1, 1))
	e.SetIPAddress(net.IPv4(192, 168, 1, 5))
	e.SetSubnetMask(net.IPv4Mask(255, 255, 255, 0))
	// TODO new example with new API
}
