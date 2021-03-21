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
// Arduino Mega 2560 CS pin
// var spiCS = machine.D44
// Arduino uno CS Pin
var spiCS = machine.D10

// declare as global value, can't escape RAM usage
var buff [1000]byte

func main() {
	// Inline declarations so not used as RAM
	var (
		// gateway or router address
		gwAddr = net.IP{192, 168, 1, 1}
		// IP address of ENC28J60
		ipAddr = net.IP{192, 168, 1, 5}
		// Hardware address of ENC28J60
		macAddr = net.HardwareAddr{0xfe, 0xfe, 0xfe, 0x22, 0x22, 0x22}
		// network mask
		netmask = net.IPMask{255, 255, 255, 0}
	)
	// enc28j60.SDB = false
	// Machine-specific configuration
	spiCS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.SPI0.Configure(machine.SPIConfig{Mode: machine.Mode0, LSBFirst: false})
	// use pin D0 as output
	machine.D0.Configure(machine.PinConfig{Mode: machine.PinOutput})

	e := enc28j60.New(spiCS, machine.SPI0)
	// Set network specific Address
	e.SetGatewayAddress(gwAddr)
	e.SetIPAddress(ipAddr)
	e.SetSubnetMask(netmask)
	err := e.Init(buff[:], macAddr)
	if err != nil {
		// println(err.Error())
		println(err.(enc28j60.ErrorCode))
	}
	// rv := e.GetRev()
	// print("enc28j60 rev:")
	// println(rv)

	s := e.NewSocket()
	_ = s
	// 0 makes a random port
	err = s.Open("arp", 0)
	if err != nil {
		println(err.(enc28j60.ErrorCode))
	}
	// ARP protocol does not support custom payload
	// we just wait for the destination to resolve our request
	println("resolving")
	gwHWAddr, err := s.Resolve()

	if err != nil {
		println(err.(enc28j60.ErrorCode))
	}
	// _ = gwHWAddr
	// // do something with gateway address
	println(string(gwHWAddr))
	println("finish!")
}
