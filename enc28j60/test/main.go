package main

import (
	"bytes"
	_ "embed"
	"machine"
	"unsafe"

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
var buff [500]byte

// Inline declarations so not used as RAM
var (
	macAddr = net2.HardwareAddr{0xDE, 0xAD, 0xBE, 0xEF, 0xFF, 0xFF}
	ipAddr  = net2.IP{192, 168, 1, 5}
)

func main() {
	frame.SDB = false
	// Machine-specific configuration
	// use pin D0 as output
	// 8MHz SPI clk
	machine.SPI0.Configure(machine.SPIConfig{Frequency: 4e6})

	e := enc28j60.New(spiCS, machine.SPI0)

	err := e.Init(buff[:], macAddr)
	if err != nil {
		printError(err)
	}
	// create variables in use
	var plen uint16
	f := new(frame.Ethernet)
	a := new(frame.ARP)
	ipf := new(frame.IP)
	tcpf := new(frame.TCP)
	ipf.Framer = tcpf
	tcpf.PseudoHeaderInfo = ipf
	var count uint
A:
	for {
		// erase previous state
		a.IPTargetAddr = nil
		a.IPTargetAddr = nil
		f.Destination = nil
		tcpf.Flags = 0
		plen = waitForPacket(e, buff[:])
		f.UnmarshalBinary(buff[:plen])
		println("count", count)
		// ARP Packet control
		if f.EtherType == frame.EtherTypeARP {
			f.Framer = a
			err = f.UnmarshalFrame(buff[:plen])
			if err != nil || !bytes.Equal(a.IPTargetAddr, ipAddr) {
				printError(err)
				continue
			}
			// println(a.String())
			f.SetResponse(macAddr)
			plen, err = f.MarshalFrame(buff[:])
			printError(err)
			e.PacketSend(buff[:plen])
			println("finish ARP shake")
			// TCP Packet control
		} else if f.EtherType == frame.EtherTypeIPv4 {
			f.Framer = ipf
			err = f.UnmarshalFrame(buff[:plen])
			if err != nil || !bytes.Equal(ipf.Destination, ipAddr) || !bytes.Equal(f.Destination, macAddr) || !tcpf.HasFlags(frame.TCPHEADER_FLAG_SYN) {
				continue
			}
			// println(ipf.String())
			f.SetResponse(macAddr)
			plen, err = f.MarshalFrame(buff[:])
			printError(err)
			e.PacketSend(buff[:plen])
			loopsDone := 0
			for (tcpf.Seq != tcpf.LastSeq+1 && len(tcpf.Data) == 0) || tcpf.HasFlags(frame.TCPHEADER_FLAG_SYN) {
				// We'll skip the incoming ACK. contains no critical information. HTTP request is what we want
				plen = waitForPacket(e, buff[:])
				err = f.UnmarshalFrame(buff[:plen])
				printError(err)
				loopsDone++
				if loopsDone > 4 {
					continue A
				}
			}
			// send ACK
			f.SetResponse(macAddr)
			plen, err = f.MarshalFrame(buff[:])
			printError(err)
			e.PacketSend(buff[:plen])

			// Send HTTP and FIN|PSH bit
			tcpf.Data = []byte(httpResponse)
			tcpf.SetFlags(frame.TCPHEADER_FLAG_FIN | frame.TCPHEADER_FLAG_PSH)
			plen, err = f.MarshalFrame(buff[:])
			printError(err)
			e.PacketSend(buff[:plen])
			nextseq := tcpf.Seq + uint32(len(tcpf.Data)) + 1
			tcpf.ClearFlags(frame.TCPHEADER_FLAG_FIN)

			for (tcpf.Seq != nextseq && !tcpf.HasFlags(frame.TCPHEADER_FLAG_FIN)) || tcpf.HasFlags(frame.TCPHEADER_FLAG_SYN) {
				plen = waitForPacket(e, buff[:])
				err = f.UnmarshalFrame(buff[:plen])
				printError(err)
				loopsDone++
				if loopsDone > 4 {
					continue A
				}
			}
			err = f.SetResponse(macAddr)
			printError(err)
			plen, err = f.MarshalFrame(buff[:])
			printError(err)
			e.PacketSend(buff[:plen])
			println("finish TCP shake")
		}
		count++
	}
}

const httpResponse = "HTTP/1.0 200 OK\r\nContent-Type: text/html\r\nPragma: no-cache\r\n\r\n<h2>..::TinyGo Rocks::..</h2>"

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
		if frame.SDB {
			print(err.Error())
		} else {
			print("error #", codeFromErrorUnsafe(err))
		}
		println()
	}
}

func codeFromErrorUnsafe(err error) uint8 {
	if err != nil {
		type eface struct { // This is how interface{} is implemented under the hood in Go
			typ uintptr
			val *uint8
		}
		ptr := unsafe.Pointer(&err)
		val := (*uint8)(unsafe.Pointer((*eface)(ptr).val))
		return *val
	}
	return 0
}
