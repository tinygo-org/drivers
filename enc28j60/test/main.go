package main

import (
	"machine"

	"github.com/jkaflik/tinygo-w5500-driver/wiznet/net"

	"tinygo.org/x/drivers/enc28j60"
)

/* Arduino MEGA 2560 SPI pins as taken from tinygo library (online documentation seems to be wrong at times)
SCK: PB1 == D52
MISO(sdo): PB2 == D51
MOSI(sdi): PB3 == D50
CS: PB0 == D53
*/
var spiCS = machine.D53

func main() {
	// best declared as a global variable for tinygo application
	var buff [1000]byte
	// Machine-specific configuration

	spiCS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.SPI0.Configure(machine.SPIConfig{Mode: machine.Mode0})
	// use pin D0 as output
	machine.D0.Configure(machine.PinConfig{Mode: machine.PinOutput})

	e := enc28j60.New(spiCS, machine.SPI0)
	err := e.Init(buff[:], net.HardwareAddr{0xfe, 0xfe, 0xfe, 0x22, 0x22, 0x22})
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
	println(string(gwHWAddr))
}
