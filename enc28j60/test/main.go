package main

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

// Arduino Mega 2560 CS pin
var spiCS = machine.D53

// Arduino uno CS Pin
// var spiCS = machine.D10 // on Arduino Uno

// declare as global value, can't escape RAM usage
var buff [160]byte

func main() {
	// linksys mac addr: C0:56:27:07:3D:71
	// laptop 28:D2:44:9A:2F:F3
	enc28j60.SDB = false
	// Inline declarations so not used as RAM
	var (
		macAddr = net2.HardwareAddr{0xDE, 0xAD, 0xBE, 0xEF, 0xFE, 0xFF}
		ipAddr  = net2.IP{192, 168, 1, 5}
	)

	// Machine-specific configuration
	// use pin D0 as output
	// 8MHz SPI clk
	machine.SPI0.Configure(machine.SPIConfig{Frequency: 4e6})

	e := enc28j60.New(spiCS, machine.SPI0)

	err := e.Init(buff[:], macAddr)
	if err != nil {
		printError(err)
	}

	// Wait for ARP Package. Make a browser request to http://192.168.1.5
	var plen uint16
	f := new(frame.EtherFrame)
	a := new(frame.ARPRequest)
	f.Framer = a
	for f.EtherType != frame.EtherTypeARP {
		plen := waitForPacket(e, buff[:])
		err = f.UnmarshalFrame(buff[:plen])
		printError(err)
	}
	println(a.String())

	// Set ARP response values using recieved ARP request
	a.SetResponse(macAddr, net2.IP(ipAddr))
	f.SetResponse(macAddr, frame.EtherTypeARP)
	f.Framer = a
	err = f.MarshalFrame(buff[:])
	printError(err)

	// send ARP response
	e.PacketSend(buff[:f.FrameLength()])

	// Wait for IPv4 request (browser request) destined for our MAC Addr
	for f.EtherType != frame.EtherTypeIPv4 || !bytes.Equal(f.Destination, macAddr) {
		plen = waitForPacket(e, buff[:])
		f.UnmarshalBinary(buff[:plen])
	}
	ipf := new(frame.IPFrame)
	tcpf := new(frame.TCP)
	// ipf.Framer = tcpf
	f.Framer = ipf

	err = ipf.UnmarshalBinary(f.Payload)
	printError(err)
	println(ipf.String())

	tcpf.UnmarshalBinary(ipf.Data)
	println(tcpf.String())

	// prepare answer
	tcpf.Ack++ // ack receive
	tcpf.SetFlags(frame.TCPHEADER_FLAG_ACK)

	// prepare ip answer
	// ipf.SetResponse()
	// ipf.Destination, ipf.Source = ipf.Source, ipf.Destination
	// ipf.Framer = &tcpf
	// f.Framer = &ipf

	// f.Framer = &ipf

	println("send tcp: ", string(byteSliceToHex(buff[:tcpf.FrameLength()])))

	plen = waitForPacket(e, buff[:])
	f.UnmarshalBinary(buff[:plen])
	println(f.String())
	println("FullEtherFrame: ", string(byteSliceToHex(buff[:plen])))

}

func waitForPacket(e *enc28j60.Dev, buff []byte) (plen uint16) {
	for plen == 0 {
		plen = e.PacketRecieve(buff[:])
		delay(500)
	}
	return
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
	if err != nil {
		if enc28j60.SDB {
			println(err.Error())
		} else {
			println("error #", err.(enc28j60.ErrorCode))
		}
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
