package frame

import (
	"encoding/binary"
	"strconv"
)

// There are 9 flags, bits 100 thru 103 are reserved
const (
	// TCP words are 4 octals, or uint32s
	TCP_WORDLEN                 = 4
	TCPHEADER_FLAGS_MASK uint16 = 0x01ff
)
const (
	TCPHEADER_FLAG_FIN = 1 << iota
	TCPHEADER_FLAG_SYN
	TCPHEADER_FLAG_RST
	TCPHEADER_FLAG_PSH
	TCPHEADER_FLAG_ACK
	TCPHEADER_FLAG_URG
	TCPHEADER_FLAG_ECE
	TCPHEADER_FLAG_CWR
	TCPHEADER_FLAG_NS
)

type TCP struct {
	// Source and Destination ports
	Source, Destination uint16
	Seq, Ack            uint32
	// LastSeq is persistent in marshalling
	// is modified when calling SetResponse
	LastSeq             uint32
	DataOffset          uint8
	Flags, WindowSize   uint16
	Checksum, UrgentPtr uint16

	// not implemented
	Options []byte
	Data    []byte
}

// UnmarshalFrame unmarshals a TCP frame from a byte slice, usually
// the byte slice contains IP data segment.
func (tcp *TCP) UnmarshalFrame(data []byte) error {
	if len(data) < 20 {
		return errBufferTooSmall
	}
	tcp.Source = binary.BigEndian.Uint16(data[0:2])
	tcp.Destination = binary.BigEndian.Uint16(data[2:4])
	tcp.Seq = binary.BigEndian.Uint32(data[4:8])
	tcp.Ack = binary.BigEndian.Uint32(data[8:12])
	tcp.DataOffset = data[12] >> 4
	tcp.Flags = TCPHEADER_FLAGS_MASK & binary.BigEndian.Uint16(data[12:14])
	tcp.WindowSize = binary.BigEndian.Uint16(data[14:16])
	tcp.Checksum = binary.BigEndian.Uint16(data[16:18])
	tcp.UrgentPtr = binary.BigEndian.Uint16(data[18:20])
	if tcp.DataOffset > 5 {
		if uint16(tcp.DataOffset)*TCP_WORDLEN > uint16(len(data)) {
			return errBufferTooSmall
		}
		tcp.Options = data[20 : tcp.DataOffset*TCP_WORDLEN]
	}
	tcp.Data = data[tcp.DataOffset*TCP_WORDLEN:]
	return nil
}

func (tcp *TCP) MarshalFrame(data []byte) error {
	if len(data) < int(tcp.FrameLength()) {
		return errBufferTooSmall
	}
	binary.BigEndian.PutUint16(data[0:2], tcp.Source)
	binary.BigEndian.PutUint16(data[2:4], tcp.Destination)

	binary.BigEndian.PutUint32(data[4:8], tcp.Seq)
	binary.BigEndian.PutUint32(data[8:12], tcp.Ack)

	binary.BigEndian.PutUint16(data[12:14], tcp.Flags)
	data[12] |= tcp.DataOffset << 4

	binary.BigEndian.PutUint16(data[14:16], tcp.WindowSize)
	// skip checksum data[16:18]
	binary.BigEndian.PutUint16(data[18:20], tcp.UrgentPtr)
	n := 20
	if tcp.DataOffset > 5 {
		copy(data[n:n+len(tcp.Options)], tcp.Options)
		n += len(tcp.Options)
	}
	copy(data[n:n+len(tcp.Data)], tcp.Data)

	binary.BigEndian.PutUint16(data[16:18], checksum(data))
	return nil
}

func (tcp *TCP) FrameLength() uint16 {
	return uint16(tcp.DataOffset)*TCP_WORDLEN + uint16(len(tcp.Options)+len(tcp.Data))
}

func (tcp *TCP) ClearOptions() {
	tcp.Options = nil
	tcp.Data = nil
}

func (tcp *TCP) SetResponse(port uint16) {
	tcp.Destination = tcp.Source
	tcp.Source = port
	if tcp.HasFlags(TCPHEADER_FLAG_SYN) {
		tcp.SetFlags(TCPHEADER_FLAG_ACK)
		tcp.Ack++
		tcp.LastSeq = tcp.Seq
		return
	}

	tcp.LastSeq = tcp.Seq
}

func (tcp *TCP) SetFlags(ORflags uint16) {
	if ORflags & ^TCPHEADER_FLAGS_MASK != 0 {
		panic("bad flag")
	}
	tcp.Flags |= uint16(ORflags)
}

// Has Flags returns true if ORflags are all set
func (tcp *TCP) HasFlags(ORflags uint16) bool { return (tcp.Flags & ORflags) == ORflags }

func (tcp *TCP) String() string {
	data := ""
	if len(tcp.Data) > 0 {
		data = string(tcp.Data)
	}
	return "TCP port " + u32toa(uint32(tcp.Source)) + "->" + u32toa(uint32(tcp.Destination)) +
		" datapacket:" + data
}

func u32toa(u uint32) string {
	return strconv.Itoa(int(u))
}
