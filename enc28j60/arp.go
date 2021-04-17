package enc28j60

import (
	"bytes"
	"encoding/binary"

	"github.com/jkaflik/tinygo-w5500-driver/wiznet/net"
)

const (
	// ethernet frame type for ARP
	efARPType       = 0x0806
	protoAddrTypeIP = 0x0800
	IPv4len         = 4
	IPv6len         = 6
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
type IP []byte

type ARPRequest struct {
	HWType, ProtoType uint16
	HWSize, ProtoSize uint8
	OpCode            uint16
	HWSenderAddr      net.HardwareAddr
	IPSenderAddr      IP
	HWTargetAddr      net.HardwareAddr
	IPTargetAddr      IP
}

// UnmarshalBinary unmarshals a payload byte slice into a ARP Request.
func (a *ARPRequest) UnmarshalBinary(payload []byte) error {
	// Verify that both proto sizes and HW size are present
	if len(payload) < 6 {
		return errIO
	}
	a.HWType = binary.BigEndian.Uint16(payload[0:2])
	a.ProtoType = binary.BigEndian.Uint16(payload[2:4])
	a.HWSize = payload[4]
	a.ProtoSize = payload[5]
	a.OpCode = binary.BigEndian.Uint16(payload[6:8])

	// 8 header size, contains 2 HWAddr and 2 ProtoAddr (IP's)
	totalSize := 8 + 2*a.HWSize + 2*a.ProtoSize
	if len(payload) < int(totalSize) {
		return errIO
	}
	// Track offset in packet for reading data (can't possibly surpass 256)
	var n uint8 = 8
	a.HWSenderAddr = make(net.HardwareAddr, a.HWSize)
	a.HWTargetAddr = make(net.HardwareAddr, a.HWSize)
	a.IPSenderAddr = make(IP, a.ProtoSize)
	a.IPTargetAddr = make(IP, a.ProtoSize)

	copy(a.HWSenderAddr, payload[n:n+a.HWSize])
	n += a.HWSize
	copy(a.IPSenderAddr, payload[n:n+a.ProtoSize])
	n += a.ProtoSize
	copy(a.HWTargetAddr, payload[n:n+a.HWSize])
	n += a.HWSize
	copy(a.IPTargetAddr, payload[n:n+a.ProtoSize])
	return nil
}

// UnmarshalBinary marshals an ARP Request into payload byte slice
func (a *ARPRequest) MarshalBinary(payload []byte) error {
	totalSize := 8 + 2*a.HWSize + 2*a.ProtoSize
	if uint16(len(payload)) < uint16(totalSize) {
		return errIO
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

func (a *ARPRequest) SetResponse(macaddr net.HardwareAddr, ip IP) error {
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

func (a *ARPRequest) String() string {
	// if bytes are only 0, then it is an ARP request
	if bytesAreAll(a.HWTargetAddr, 0) {
		return a.HWSenderAddr.String() + "->" +
			"who has " + a.IPTargetAddr.String() + "?" + " Tell " + a.IPSenderAddr.String()
	}
	return a.HWSenderAddr.String() + "->" +
		"I have " + a.IPSenderAddr.String() + "! Telling " + a.IPTargetAddr.String() + ", aka " + a.HWTargetAddr.String()
}

// writes ARP bytes to ethframe's payload without modifying the rest of the buffer
// returns length of payload written
func (s *Socket) writeARP() uint16 {
	s.payloadwrite(0, []byte{
		0, 1, // Write HW AT
		byte(protoAddrTypeIP % 256), byte(protoAddrTypeIP >> 8), // write Proto AT
		6, 4, // write HW AL and Proto AL
		0, 1, // write OP Code
	})
	s.payloadwrite(8, s.d.macaddr)
	s.payloadwrite(14, s.d.myip)
	s.payloadwrite(18, []byte{0, 0, 0, 0, 0, 0}) // HW AoT (empty because it is what we want this dude to fill)
	s.payloadwrite(24, s.d.gatewayip)
	return 24 + 4 //28 is length of ARP payload
}

func (s *Socket) Resolve() (net.HardwareAddr, error) {
	if s.mode != socketARPMode {
		return nil, errARPViolation
	}
	var plen uint16
	plen = s.writeARP() + efPayloadOffset
	s.d.PacketSend(s.d.buffer[:plen])

	plen = 0
	for plen == 0 {
		plen = s.d.PacketRecieve(s.d.buffer)
	}
	// discard ethernet frame buffer. look in ARP payload for hardware address of target (us)
	hwAoTIdx := idxRabinKarpBytes(s.d.buffer[efPayloadOffset:], s.dstMacaddr)
	if hwAoTIdx == -1 {
		dbp("got:", s.d.buffer)
		return nil, errUnableToResolveARP
	}
	return net.HardwareAddr(s.d.buffer[hwAoTIdx-10 : hwAoTIdx-4]), nil
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

// To4 converts the IPv4 address ip to a 4-byte representation.
// If ip is not an IPv4 address, To4 returns nil.
func (ip IP) To4() IP {
	if len(ip) == IPv4len {
		return ip
	}
	if len(ip) == IPv6len &&
		bytesAreAll(ip[0:10], 0) &&
		ip[10] == 0xff &&
		ip[11] == 0xff {
		return ip[12:16]
	}
	return nil
}

func (ip IP) String() string {
	if len(ip) == 0 {
		return "<nil>"
	}
	if p4 := ip.To4(); len(p4) == IPv4len {
		const maxIPv4StringLen = len("255.255.255.255")
		b := make([]byte, maxIPv4StringLen)

		n := ubtoa(b, 0, p4[0])
		b[n] = '.'
		n++

		n += ubtoa(b, n, p4[1])
		b[n] = '.'
		n++

		n += ubtoa(b, n, p4[2])
		b[n] = '.'
		n++

		n += ubtoa(b, n, p4[3])
		return string(b[:n])
	}
	return "ipv4+ not implemented"
}

func ubtoa(dst []byte, start int, v byte) int {
	if v < 10 {
		dst[start] = v + '0'
		return 1
	} else if v < 100 {
		dst[start+1] = v%10 + '0'
		dst[start] = v/10 + '0'
		return 2
	}

	dst[start+2] = v%10 + '0'
	dst[start+1] = (v/10)%10 + '0'
	dst[start] = v/100 + '0'
	return 3
}
