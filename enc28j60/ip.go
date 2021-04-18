package enc28j60

import "encoding/binary"

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
	Source, Destination IP
	Data                []byte
}

func (ip *IPFrame) UnmarshalBinary(payload []byte) error {
	ip.Version = payload[0]
	addrlen := 4
	if ip.Version != IPHEADER_VERSION_4 {
		return errIPNotImplemented
	}
	if len(payload) < 12+2*addrlen {
		return errBufferSize
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
	ip.Source = make(IP, addrlen)
	ip.Destination = make(IP, addrlen)
	copy(ip.Source, payload[n:n+addrlen])
	n += addrlen
	copy(ip.Destination, payload[n:n+addrlen])
	n += addrlen
	ip.Data = payload[n:]
	return nil
}

func (ip *IPFrame) MarshalBinary(payload []byte) error {
	addrlen := 4 // for now only IPv4
	if len(payload) < 12+2*addrlen {
		return errBufferSize
	}
	payload[0] = ip.Version
	payload[1] = ip.IHL
	ip.SetLength()
	binary.BigEndian.PutUint16(payload[2:4], ip.TotalLength)
	binary.BigEndian.PutUint16(payload[4:6], ip.ID)
	binary.BigEndian.PutUint16(payload[6:8], ip.Flags)
	payload[8] = ip.TTL
	payload[9] = ip.Protocol
	// skip checksum until end
	n := 12
	copy(payload[n:n+addrlen], ip.Source)
	n += addrlen
	copy(payload[n:n+addrlen], ip.Destination)
	n += addrlen

	// TODO Separate checksum in its own function
	var sum uint32
	for i := 0; i < n/2; i++ {
		sum += uint32(binary.BigEndian.Uint16(payload[i : i+2]))
	}
	for sum > 0xffff {
		sum = sum&0xffff + sum>>8 // sum&0xffff0000 == sum>>8 TODO, see which one is better
	}
	binary.BigEndian.PutUint16(payload[10:12], ^uint16(sum))
	return nil
}

func (ip *IPFrame) SetLength() {
	addrlen := 4
	header := 12 + uint16(addrlen*2)
	ip.TotalLength = header + uint16(len(ip.Data))
}

func (ip *IPFrame) String() string {
	return "IPv4 " + ip.Source.String() + "->" + ip.Destination.String()
}
