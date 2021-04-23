package main

import (
	"bytes"
	_ "embed"
	"machine"
	"unsafe"

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
var buff [500]byte

func main() {
	// linksys mac addr: C0:56:27:07:3D:71
	// laptop 28:D2:44:9A:2F:F3
	enc28j60.SDB = false
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

	for f.EtherType != frame.EtherTypeARP && !bytes.Equal(a.IPTargetAddr, ipAddr) {
		plen := waitForPacket(e, buff[:])
		err = f.UnmarshalFrame(buff[:plen])
		printError(err)
	}
	println(a.String())
	// we must set our mac addresses for the ARP to fulfill. This will be done automatically in future by constructing a Ethernet Frame

	// Set ARP response values using recieved ARP request
	f.SetResponse(macAddr)

	plen, err = f.MarshalFrame(buff[:])
	printError(err)

	// send ARP response
	e.PacketSend(buff[:plen])
	a = nil // clear ARP memory once done
	// Setup TCP frame
	ipf := new(frame.IP)
	tcpf := new(frame.TCP)

	ipf.Framer = tcpf
	tcpf.PseudoHeaderInfo = ipf
	f.Framer = ipf
	// enc28j60.SDB = true
	frame.SDB = true
	// Wait for IPv4 request (browser request) destined for our MAC Addr
	for (f.EtherType != frame.EtherTypeIPv4 /*&& tcpf.HasFlags(frame.TCPHEADER_FLAG_SYN)*/) || !bytes.Equal(f.Destination, macAddr) {
		plen = waitForPacket(e, buff[:])
		f.UnmarshalFrame(buff[:plen])
	}

	// UnmarshalFrame
	// frame.SDB = false
	err = f.UnmarshalFrame(buff[:plen])

	printError(err)
	println(ipf.String())

	println(tcpf.String())

	// prepare answer .SetResponse sets all sub framer responses
	f.SetResponse(macAddr)

	plen, err = f.MarshalFrame(buff[:])
	printError(err)
	// Send ACK through TCP, wait for connection establishment
	e.PacketSend(buff[:plen])
	println("Waiting for TCP response")
	for tcpf.Seq != tcpf.LastSeq+1 {
		plen = waitForPacket(e, buff[:])
		f.UnmarshalFrame(buff[:plen])
		println(tcpf.String())
	}
	// -- connection established --
	// send Ack and prepare to receive HTTP
	// FIRST
	f.SetResponse(macAddr)

	// f.ClearOptions()

	println(f.String())
	println("FullEtherFrame: ", string(hex.Bytes(buff[:plen])))
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func waitForPacket(e *enc28j60.Dev, buff []byte) (plen uint16) {
	for plen == 0 {
		plen = e.PacketRecieve(buff[:])
	}
	return
}

func printError(err error) {
	if err != nil {
		if enc28j60.SDB {
			println(err.Error())
		} else {
			type eface struct {
				typ, val unsafe.Pointer
			}
			passed_value := (*eface)(unsafe.Pointer(&err)).val
			println("error #", *(*uint8)(passed_value))
		}
	}
}
