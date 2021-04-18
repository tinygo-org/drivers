package enc28j60

import "encoding/binary"

const (
	IPHEADER_FLAG_DONTFRAGMENT = 0x4000
	IPHEADER_VERSION_4         = 0x45
	IPHEADER_PROTOCOL_TCP      = 6
)

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

func (ip *IPFrame) String() string {
	return "IPv4 " + ip.Source.String() + "->" + ip.Destination.String()
}
