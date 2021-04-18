package frame

import (
	"encoding/binary"

	"tinygo.org/x/drivers/net2"
)

const (
	IPHEADER_FLAG_DONTFRAGMENT = 0x4000
	IPHEADER_VERSION_4         = 0x45
	IPHEADER_PROTOCOL_TCP      = 6
)

// See https://hpd.gasmi.net/ to decode Hex Frames

// TODO Handle IGMP
// Frame example: 01 00 5E 00 00 FB 28 D2 44 9A 2F F3 08 00 46 C0 00 20 00 00 40 00 01 02 41 04 C0 A8 01 70 E0 00 00 FB 94 04 00 00 16 00 09 04 E0 00 00 FB 00 00 00 00 00 00 00 00 00 00 00 00 00 00

// TODO Handle LLC Logical Link Control
// Frame example: 05 62 70 73 D7 10 80 04 6C 00 02 00 00 04 00 00 10 20 41 70 00 00 00 0E 00 00 00 19 40 40 00 01 16 4E E9 B0 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00

type IPFrame struct {
	Version, IHL uint8
	TotalLength  uint16
	//TOS          uint16
	ID, Flags     uint16
	TTL, Protocol uint8

	HeaderChecksum      uint16
	Source, Destination net2.IP
	Data                []byte
	// OOP comes to save the day
	Framer
}

func (ip *IPFrame) MarshalFrame(payload []byte) error {
	const addrlen = 4 // for now only IPv4
	if uint16(len(payload)) < ip.FrameLength() {
		return errBufferTooSmall
	}
	if len(ip.Source) != addrlen || len(ip.Destination) != addrlen {
		return errBadIP
	}

	payload[0] = ip.Version
	payload[1] = ip.IHL
	ip.setLength()
	binary.BigEndian.PutUint16(payload[2:4], ip.TotalLength)
	binary.BigEndian.PutUint16(payload[4:6], ip.ID)
	binary.BigEndian.PutUint16(payload[6:8], ip.Flags)
	payload[8] = ip.TTL
	payload[9] = ip.Protocol
	// skip checksum data [10:12] until end
	n := 12
	n += copy(payload[n:n+addrlen], ip.Source)

	n += copy(payload[n:n+addrlen], ip.Destination)

	binary.BigEndian.PutUint16(payload[10:12], checksum(payload[:n]))
	if ip.Framer != nil {
		return ip.Framer.MarshalFrame(payload[n:])
	}
	copy(payload[n:], ip.Data)
	return nil
}

func (ip *IPFrame) FrameLength() uint16 {
	const addrlen uint16 = 4 // for now only IPv4
	headlen := 12 + 2*addrlen
	if ip.Framer != nil {
		return headlen + ip.Framer.FrameLength()
	}
	return headlen + uint16(len(ip.Data))
}

func (ip *IPFrame) UnmarshalBinary(payload []byte) error {
	ip.Version = payload[0]
	addrlen := 4
	if ip.Version != IPHEADER_VERSION_4 {
		return errIPNotImplemented
	}
	if len(payload) < 12+2*addrlen {
		return errBufferTooSmall
	}
	ip.IHL = payload[1]
	ip.TotalLength = binary.BigEndian.Uint16(payload[2:4])
	ip.ID = binary.BigEndian.Uint16(payload[4:6])
	ip.Flags = binary.BigEndian.Uint16(payload[6:8])
	if ip.Flags&IPHEADER_FLAG_DONTFRAGMENT == 0 {
		return errIPNotImplemented
	}
	ip.TTL = payload[8]
	ip.Protocol = payload[9]
	if ip.Protocol != IPHEADER_PROTOCOL_TCP {
		return errIPNotImplemented
	}
	ip.HeaderChecksum = binary.BigEndian.Uint16(payload[10:12])
	n := 12
	ip.Source = make(net2.IP, addrlen)
	ip.Destination = make(net2.IP, addrlen)
	copy(ip.Source, payload[n:n+addrlen])
	n += addrlen
	copy(ip.Destination, payload[n:n+addrlen])
	n += addrlen
	ip.Data = payload[n:]
	return nil
}

// SetResponse removes Data Pointer and reverses Source and Destination IP Addresses
func (ip *IPFrame) SetResponse() {
	ip.Destination, ip.Source = ip.Source, ip.Destination
	ip.Data = nil
}

// func (ip *IPFrame) MarshalBinary(payload []byte) error {
// 	addrlen := 4 // for now only IPv4
// 	if len(payload) < 12+2*addrlen {
// 		return errBufferSize
// 	}
// 	payload[0] = ip.Version
// 	payload[1] = ip.IHL
// 	ip.setLength()
// 	binary.BigEndian.PutUint16(payload[2:4], ip.TotalLength)
// 	binary.BigEndian.PutUint16(payload[4:6], ip.ID)
// 	binary.BigEndian.PutUint16(payload[6:8], ip.Flags)
// 	payload[8] = ip.TTL
// 	payload[9] = ip.Protocol
// 	// skip checksum data [10:12] until end
// 	n := 12
// 	copy(payload[n:n+addrlen], ip.Source)
// 	n += addrlen
// 	copy(payload[n:n+addrlen], ip.Destination)
// 	n += addrlen

// 	binary.BigEndian.PutUint16(payload[10:12], checksum(payload))
// 	return nil
// }

func (ip *IPFrame) setLength() { ip.TotalLength = ip.FrameLength() }

func (ip *IPFrame) String() string {
	return "IPv4 " + ip.Source.String() + "->" + ip.Destination.String()
}

func checksum(data []byte) uint16 {
	var sum uint32
	n := len(data) / 2
	// automatic padding of data
	if len(data)%2 != 0 {
		sum += uint32(data[len(data)-1]) << 8
	}
	for i := 0; i < n; i++ {
		sum += uint32(binary.BigEndian.Uint16(data[i : i+2]))
	}
	for sum > 0xffff {
		sum = sum&0xffff + sum>>8
	}
	return ^uint16(sum)
}
