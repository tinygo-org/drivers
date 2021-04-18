package enc28j60_test

import (
	"bytes"
	"machine"
	"time"

	"tinygo.org/x/drivers/enc28j60"
	"tinygo.org/x/drivers/frame"
	"tinygo.org/x/drivers/net2"
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

func ExampleEthernetChatter() {
	// best declared as a global variable for tinygo application
	var buff [1000]byte
	// Machine-specific configuration
	// use pin D0 as output
	// 8MHz SPI clk
	machine.SPI0.Configure(machine.SPIConfig{Frequency: 8e6})
	e := enc28j60.New(machine.D10, machine.SPI0)
	MAC := net2.HardwareAddr{0xfe, 0xfe, 0xfe, 0x22, 0x22, 0x22}
	err := e.Init(buff[:], MAC)
	if err != nil {
		println(err.Error())
	}

	// TODO new example with new API
}

func Example_tCPResponse() {
	var buff [160]byte
	// SPI Chip select pin. Can be any Digital pin
	var spiCS = machine.D53
	// Inline declarations so not used as RAM
	var (
		MAC  = net2.HardwareAddr{0xDE, 0xAD, 0xBE, 0xEF, 0xFE, 0xFF}
		MyIP = net2.IP{192, 168, 1, 5}
	)

	// 8MHz SPI clk or higher is preffered for the ENC28J60 for older revisions. See Errata Rev 4B
	machine.SPI0.Configure(machine.SPIConfig{Frequency: 4e6})

	e := enc28j60.New(spiCS, machine.SPI0)

	err := e.Init(buff[:], MAC)
	if err != nil {
		printError(err)
	}

	// Wait for ARP Package. Make a browser request to http://192.168.1.5
	var plen uint16
	f := new(frame.Ethernet)
	a := new(frame.ARP)
	f.Framer = a
	for f.EtherType != frame.EtherTypeARP {
		plen := waitForPacket(e, buff[:])
		err = f.UnmarshalFrame(buff[:plen])
		printError(err)
	}
	println(a.String())

	// Set ARP response values using recieved ARP request
	a.SetResponse(MAC, MyIP)
	f.SetResponse(MAC, frame.EtherTypeARP)
	f.Framer = a
	err = f.MarshalFrame(buff[:])
	printError(err)

	// send ARP response
	e.PacketSend(buff[:f.FrameLength()])

	// Wait for IPv4 TCP request (browser request) destined for our MAC Addr
	for f.EtherType != frame.EtherTypeIPv4 || !bytes.Equal(f.Destination, MAC) {
		plen = waitForPacket(e, buff[:])
		f.UnmarshalBinary(buff[:plen])
	}
	ipf := new(frame.IP)
	tcpf := new(frame.TCP)
	ipf.Framer = tcpf
	f.Framer = ipf

	// UnmarshalFrame
	err = f.UnmarshalFrame(buff[:])
	printError(err)
	// Print responses recieved
	println(ipf.String())
	println(tcpf.String())
}

func waitForPacket(e *enc28j60.Dev, buff []byte) (plen uint16) {
	for plen == 0 {
		plen = e.PacketRecieve(buff[:])
		time.Sleep(time.Millisecond * time.Duration(500))
	}
	return
}

func printError(err error) {
	if err != nil {
		if enc28j60.SDB {
			println(err.Error())
		} else {
			println("error #", err.(enc28j60.ErrorCode))
		}
	}
}
