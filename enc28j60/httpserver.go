package enc28j60

import (
	"bytes"
	"unsafe"

	"tinygo.org/x/drivers/bytealg"
	"tinygo.org/x/drivers/frame"
	"tinygo.org/x/drivers/net2"
)

func (e *Dev) HTTPListenAndServe(IPAddr net2.IP, handler func(URL []byte) (response []byte)) (err error) {
	var plen uint16
	const httpHeader = "HTTP/1.0 200 OK\r\nContent-Type: text/html\r\nPragma: no-cache\r\n\r\n"
	macAddr := e.macaddr
	buff := e.buffer

	f := new(frame.Ethernet)
	a := new(frame.ARP)
	ipf := new(frame.IP)
	tcpf := new(frame.TCP)
	ipf.Framer = tcpf
	tcpf.PseudoHeaderInfo = ipf
	var count uint
	var url []byte
A:
	for {
		// erase previous state
		a.IPTargetAddr = nil
		a.IPTargetAddr = nil
		f.Destination = nil
		tcpf.Flags = 0
		plen = waitForPacket(e, buff[:])
		f.UnmarshalBinary(buff[:plen])

		// ARP Packet control
		if f.EtherType == frame.EtherTypeARP {
			f.Framer = a
			err = f.UnmarshalFrame(buff[:plen])
			if err != nil || !bytes.Equal(a.IPTargetAddr, IPAddr) {
				printError(err)
				continue
			}
			f.SetResponse(macAddr)
			plen, err = f.MarshalFrame(buff[:])
			printError(err)
			e.PacketSend(buff[:plen])
			count++
			// TCP Packet control
		} else if f.EtherType == frame.EtherTypeIPv4 {
			f.Framer = ipf
			err = f.UnmarshalFrame(buff[:plen])
			if err != nil || !bytes.Equal(ipf.Destination, IPAddr) || !bytes.Equal(f.Destination, macAddr) || !tcpf.HasFlags(frame.TCPHEADER_FLAG_SYN) {
				continue
			}
			f.SetResponse(macAddr)
			plen, err = f.MarshalFrame(buff[:])
			printError(err)
			e.PacketSend(buff[:plen])
			loopsDone := 0
			for tcpf.Seq != tcpf.LastSeq+1 || len(tcpf.Data) == 0 || tcpf.HasFlags(frame.TCPHEADER_FLAG_SYN) {
				// Get incoming ACK and skip it (len=0) and get HTTP request
				plen = waitForPacket(e, buff[:])
				err = f.UnmarshalFrame(buff[:plen])
				printError(err)
				loopsDone++
				if loopsDone > 4 {
					continue A
				}
			}
			endlineIdx := bytealg.IdxRabinKarpBytes(tcpf.Data, []byte("\r\n"))
			if endlineIdx < 0 {
				continue
			}
			if endlineIdx-9 < 5 { // error prevention
				continue
			}
			url = tcpf.Data[4 : endlineIdx-9]
			response := handler(url)
			// send ACK
			f.SetResponse(macAddr)
			plen, err = f.MarshalFrame(buff[:])
			printError(err)

			e.PacketSend(buff[:plen])

			// Send HTTP and FIN|PSH bit
			tcpf.Data = append([]byte(httpHeader), response...)

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
			count++
		}

	}
}
func waitForPacket(e *Dev, buff []byte) (plen uint16) {
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
			typ *uint
			val *uint8
		}
		ptr := unsafe.Pointer(&err)
		val := (*uint8)(unsafe.Pointer((*eface)(ptr).val))
		return *val
	}
	return 0
}
