package frame

import (
	"bytes"
	"encoding/binary"

	"tinygo.org/x/drivers/net2"
)

const (
	// ethernet frame type for ARP
	efARPType       = 0x0806
	protoAddrTypeIP = 0x0800
)

/* ARP Frame (Address resolution protocol)
see https://www.youtube.com/watch?v=aamG4-tH_m8

Legend:
	HW:    Hardware
	AT:    Address type
	AL:    Address Length
	AoS:   Address of sender
	AoT:   Address of Target
	Proto: Protocol
0      2          4       5          6         8       14          18       24          28
| HW AT | Proto AT | HW AL | Proto AL | OP Code | HW AoS | Proto AoS | HW AoT | Proto AoT |
|  2B   |  2B      |  1B   |  1B      | 2B      |   6B   |    4B     |  6B    |   4B
| ethern| IP       |macaddr|          |ask|reply|                    |for op=1|
| = 1   |=0x0800   |=6     |=4        | 1 | 2   |       known        |=0      |
*/

type ARP struct {
	HWType, ProtoType uint16
	HWSize, ProtoSize uint8
	OpCode            uint16
	HWSenderAddr      net2.HardwareAddr
	IPSenderAddr      net2.IP
	HWTargetAddr      net2.HardwareAddr
	IPTargetAddr      net2.IP
}

// MarshalFrame marshals an ARP Request into payload byte slice
func (a *ARP) MarshalFrame(payload []byte) error {
	totalSize := 8 + 2*a.HWSize + 2*a.ProtoSize
	if uint16(len(payload)) < uint16(totalSize) {
		return errBufferTooSmall
	}
	binary.BigEndian.PutUint16(payload[0:2], a.HWType)
	binary.BigEndian.PutUint16(payload[2:4], a.ProtoType)
	payload[4] = a.HWSize
	payload[5] = a.ProtoSize
	binary.BigEndian.PutUint16(payload[6:8], a.OpCode)
	var n uint8 = 8
	copy(payload[n:n+a.HWSize], a.HWSenderAddr)
	n += a.HWSize
	copy(payload[n:n+a.ProtoSize], a.IPSenderAddr)
	n += a.ProtoSize
	copy(payload[n:n+a.HWSize], a.HWTargetAddr)
	n += a.HWSize
	copy(payload[n:n+a.ProtoSize], a.IPTargetAddr)
	return nil
}

func (a *ARP) FrameLength() uint16 {
	// TODO maybe set these in some constructor function that returns an *ARPRequest pointer
	if a.HWSize == 0 { // set basic framelengths if not set
		a.HWSize = 6
		a.ProtoSize = 4 //suppose ipv4
	}
	return 8 + uint16(a.HWSize+a.ProtoSize)*2
}

// UnmarshalFrame unmarshals a payload byte slice into a ARP Request. Implements Framer Interface
func (a *ARP) UnmarshalFrame(payload []byte) error {
	// Verify that both proto sizes and HW size are present
	if len(payload) < 6 {
		return errBufferTooSmall
	}
	a.HWType = binary.BigEndian.Uint16(payload[0:2])
	a.ProtoType = binary.BigEndian.Uint16(payload[2:4])
	a.HWSize = payload[4]
	a.ProtoSize = payload[5]
	a.OpCode = binary.BigEndian.Uint16(payload[6:8])

	// 8 header size, contains 2 HWAddr and 2 ProtoAddr (IP's)
	addrSectorLen := 2 * (a.HWSize + a.ProtoSize)
	const addrOffset = 8
	totalSize := addrOffset + addrSectorLen
	if len(payload) < int(totalSize) {
		return errBufferTooSmall
	}

	// Track offset in packet for reading data (can't possibly surpass 256)
	var n uint8 = 0 // bb pointer
	// make one segment allocation and store all addresses there. This eases the copying to one `copy` call
	bb := make([]byte, addrSectorLen)
	a.HWSenderAddr = bb[n : n+a.HWSize]
	n += a.HWSize
	a.IPSenderAddr = bb[n : n+a.ProtoSize]
	n += a.ProtoSize
	a.HWTargetAddr = bb[n : n+a.HWSize]
	n += a.HWSize
	a.IPTargetAddr = bb[n : n+a.ProtoSize]
	copy(bb, payload[addrOffset:addrOffset+addrSectorLen])
	return nil
}

func (a *ARP) SetResponse(macaddr net2.HardwareAddr, ip net2.IP) error {
	// These must be pre-filled by an arp response
	if a.HWTargetAddr == nil || a.HWSenderAddr == nil || !bytes.Equal(a.IPTargetAddr, ip) {
		return errBadARP
	}
	if len(macaddr) > 255 || uint8(len(macaddr)) != a.HWSize {
		return errBadMac
	}
	a.HWTargetAddr, a.HWSenderAddr = a.HWSenderAddr, macaddr
	a.IPTargetAddr = a.IPSenderAddr
	a.IPSenderAddr = ip
	return nil
}

func (a *ARP) String() string {
	// if bytes are only 0, then it is an ARP request
	if bytesAreAll(a.HWTargetAddr, 0) {
		return "ARP " + a.HWSenderAddr.String() + "->" +
			"who has " + a.IPTargetAddr.String() + "?" + " Tell " + a.IPSenderAddr.String()
	}
	return "ARP " + a.HWSenderAddr.String() + "->" +
		"I have " + a.IPSenderAddr.String() + "! Tell " + a.IPTargetAddr.String() + ", aka " + a.HWTargetAddr.String()
}

// bytesAreAll returns true if b is composed of only unit bytes
func bytesAreAll(b []byte, unit byte) bool {
	for i := range b {
		if b[i] != unit {
			return false
		}
	}
	return true
}
