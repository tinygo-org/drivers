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

type IP struct {
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

func (ip *IP) MarshalFrame(payload []byte) (uint16, error) {
	_log("IP:marshal")
	const addrlen = 4 // for now only IPv4
	if uint16(len(payload)) < 12 {
		return 0, ErrBufferTooSmall
	}
	if len(ip.Source) != addrlen || len(ip.Destination) != addrlen {
		return 0, ErrBadIP
	}

	payload[0] = ip.Version
	payload[1] = ip.IHL
	ip.setLength()

	binary.BigEndian.PutUint16(payload[4:6], ip.ID)
	binary.BigEndian.PutUint16(payload[6:8], ip.Flags)
	payload[8] = ip.TTL
	payload[9] = ip.Protocol
	// skip checksum data [10:12] until end
	n := 12
	n += copy(payload[n:n+addrlen], ip.Source)

	n += copy(payload[n:n+addrlen], ip.Destination)
	payload[10] = 0 // we set checksum field to zero to exclude it from calculation
	payload[11] = 0

	if ip.Framer != nil {
		m, err := ip.Framer.MarshalFrame(payload[n:])
		ip.TotalLength = m + uint16(n)
		binary.BigEndian.PutUint16(payload[2:4], ip.TotalLength)
		binary.BigEndian.PutUint16(payload[10:12], checksumRFC791(payload[:n]))
		return ip.TotalLength, err
	}
	binary.BigEndian.PutUint16(payload[10:12], checksumRFC791(payload[:n]))
	n += copy(payload[n:], ip.Data)
	return uint16(n), nil
}

func (ip *IP) UnmarshalFrame(payload []byte) error {
	_log("IP:unmarshal")
	ip.Version = payload[0]
	addrlen := 4
	if ip.Version != IPHEADER_VERSION_4 {
		return ErrIPNotImplemented
	}
	if len(payload) < 12+2*addrlen {
		return ErrBufferTooSmall
	}
	ip.IHL = payload[1]
	ip.TotalLength = binary.BigEndian.Uint16(payload[2:4])
	ip.ID = binary.BigEndian.Uint16(payload[4:6])
	ip.Flags = binary.BigEndian.Uint16(payload[6:8])
	if ip.Flags&IPHEADER_FLAG_DONTFRAGMENT == 0 {
		return ErrIPNotImplemented
	}
	ip.TTL = payload[8]
	ip.Protocol = payload[9]
	if ip.Protocol != IPHEADER_PROTOCOL_TCP {
		return ErrIPNotImplemented
	}
	ip.HeaderChecksum = binary.BigEndian.Uint16(payload[10:12])
	n := 12
	// allocate single segment to store both source and destination. uses only one `copy`
	bb := make([]byte, addrlen*2)

	ip.Source = bb[0:addrlen]                //make(net2.IP, addrlen)
	ip.Destination = bb[addrlen : addrlen*2] //make(net2.IP, addrlen)
	n += copy(bb, payload[n:n+addrlen*2])
	if ip.Framer != nil {
		return ip.Framer.UnmarshalFrame(payload[n:])
	}
	ip.Data = payload[n:]
	return nil
}
func (ip *IP) FrameLength() uint16 {
	const addrlen uint16 = 4 // for now only IPv4
	headlen := 12 + 2*addrlen
	if ip.Framer != nil {
		return headlen + ip.Framer.FrameLength()
	}
	return headlen + uint16(len(ip.Data))
}

// SetResponse removes Data Pointer and reverses Source and Destination IP Addresses
func (ip *IP) SetResponse(MAC net2.HardwareAddr) error {
	ip.Destination, ip.Source = ip.Source, ip.Destination
	ip.Data = nil
	if ip.Framer != nil {
		return ip.Framer.SetResponse(MAC)
	}
	return nil
}

func (ip *IP) setLength() { ip.TotalLength = ip.FrameLength() }

func (ip *IP) String() string {
	return "IPv4 " + ip.Source.String() + "->" + ip.Destination.String()
}
