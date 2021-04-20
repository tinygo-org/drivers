package main

import (
	"bytes"
	"machine"
	"time"

	"tinygo.org/x/drivers/enc28j60"
	"tinygo.org/x/drivers/encoding/hex"
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
var buff [300]byte

func main() {
	// linksys mac addr: C0:56:27:07:3D:71
	// laptop 28:D2:44:9A:2F:F3
	enc28j60.SDB = true
	// Inline declarations so not used as RAM
	var (
		macAddr = net2.HardwareAddr{0xDE, 0xAD, 0xBE, 0xEF, 0xFF, 0xFF}
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
	a.SetResponse(macAddr, ipAddr)
	f.SetResponse(macAddr, frame.EtherTypeARP)
	f.Framer = a
	plen, err = f.MarshalFrame(buff[:])
	printError(err)

	// send ARP response
	e.PacketSend(buff[:plen])

	// Wait for IPv4 request (browser request) destined for our MAC Addr
	for f.EtherType != frame.EtherTypeIPv4 || !bytes.Equal(f.Destination, macAddr) {
		plen = waitForPacket(e, buff[:])
		f.UnmarshalBinary(buff[:plen])
	}
	ipf := new(frame.IP)
	tcpf := new(frame.TCP)
	ipf.Framer = tcpf
	f.Framer = ipf

	// UnmarshalFrame

	err = f.UnmarshalFrame(buff[:plen])

	printError(err)
	println(ipf.String())

	println(tcpf.String())

	// prepare answer
	ipf.SetResponse()
	tcpf.SetResponse(80, ipf)

	f.SetResponse(macAddr, 0)

	plen, err = f.MarshalFrame(buff[:])
	printError(err)
	f.ClearOptions()
	// Send ACK through TCP
	e.PacketSend(buff[:plen])
	println("Waiting for TCP response")
	for tcpf.Seq == tcpf.LastSeq {
		plen = waitForPacket(e, buff[:])
		f.UnmarshalFrame(buff[:plen])
	}

	println(f.String())
	println("FullEtherFrame: ", string(hex.Bytes(buff[:plen])))
	println(tcpf.String())
}

func waitForPacket(e *enc28j60.Dev, buff []byte) (plen uint16) {
	for plen == 0 {
		plen = e.PacketRecieve(buff[:])
		time.Sleep(time.Millisecond * time.Duration(1))
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
