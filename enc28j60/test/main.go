package main

import (
	"machine"
	"time"

	"github.com/jkaflik/tinygo-w5500-driver/wiznet/net"

	"tinygo.org/x/drivers/enc28j60"
)

/* Arduino Uno SPI pins:
sck:  PB5, // is D13
sdo:  PB3, // MOSI is D11
sdi:  PB4, // MISO is D12
cs:   PB2} // CS  is D10
*/

/* Arduino MEGA 2560 SPI pins as taken from tinygo library (online documentation seems to be wrong at times)
SCK: PB1 == D52
MOSI(sdo): PB2 == D51
MISO(sdi): PB3 == D50
CS: PB0 == D53
*/

// Arduino Mega 2560 CS pin
var spiCS = machine.D53

// Arduino uno CS Pin
// var spiCS = machine.D10 // on Arduino Uno

// declare as global value, can't escape RAM usage
var buff [80]byte

var (
	// gateway or router address
	gwAddr = net.IP{192, 168, 1, 1}
	// // IP address of ENC28J60
	ipAddr = net.IP{192, 168, 1, 5}
	// // Hardware address of ENC28J60
	macAddr = net.HardwareAddr{0xDE, 0xAD, 0xBE, 0xEF, 0xFE, 0xFF}
	// // network mask
	netmask = net.IPMask{255, 255, 255, 0}
)

func main() {
	// linksys mac addr: C0:56:27:07:3D:71
	// laptop 28:D2:44:9A:2F:F3
	enc28j60.SDB = true
	// Inline declarations so not used as RAM

	// Machine-specific configuration
	// use pin D0 as output
	// 8MHz SPI clk
	machine.SPI0.Configure(machine.SPIConfig{Frequency: 4e6})

	e := enc28j60.New(spiCS, machine.SPI0)

	err := e.Init(buff[:], macAddr)
	if err != nil {
		printError(err)
	}
	// Set network specific Address
	e.SetGatewayAddress(gwAddr)
	e.SetIPAddress(ipAddr)
	e.SetSubnetMask(netmask)
	plen := e.PacketRecieve(buff[:])
	for plen == 0 {
		delay(500)
		plen = e.PacketRecieve(buff[:])
	}
	var f enc28j60.EtherFrame
	f.UnmarshalBinary(buff[:plen])
	if f.EtherType == enc28j60.EtherTypeARP {
		f.Destination, f.Source = f.Source, macAddr
		var a enc28j60.ARPRequest
		a.UnmarshalBinary(f.Payload)
		println("arp:", a.String())
		a.SetResponse(macAddr, enc28j60.IP(ipAddr))
		println("arp:", a.String())
		a.MarshalBinary(f.Payload)

	}
	println(string(byteSliceToHex(buff[:])))
	println(f.String())
}

func testConn() {
	machine.SPI0.Configure(machine.SPIConfig{Frequency: 4e6})
	e := enc28j60.TestConn(spiCS, machine.SPI0)
	if e != nil {
		printError(e)
	}
}

func test() {
	machine.SPI0.Configure(machine.SPIConfig{Frequency: 4e6})
	e := enc28j60.TestSPI(spiCS, machine.SPI0)
	if e != nil {
		printError(e)
	}
}

func printError(err error) {
	if enc28j60.SDB {
		println(err.Error())
	} else {
		println("error #", err.(enc28j60.ErrorCode))
	}
}

func byteSliceToHex(b []byte) []byte {
	o := make([]byte, len(b)*2)
	for i := 0; i < len(b); i++ {
		aux := byteToHex(b[i])
		o[i*2] = aux[0]
		o[i*2+1] = aux[1]
	}
	return o
}

func byteToHex(b byte) []byte {
	var res [2]byte
	res[0], res[1] = (b>>4)+'0', (b&0b0000_1111)+'0'
	if (b >> 4) > 9 {
		res[0] = (b >> 4) + 'A' - 10
	}
	if (b & 0b0000_1111) > 9 {
		res[1] = (b & 0b0000_1111) + 'A' - 10
	}
	return res[:]
}

func delay(millis uint32) {
	time.Sleep(time.Millisecond * time.Duration(millis))
}
