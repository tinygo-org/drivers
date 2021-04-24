package frame

import (
	"encoding/binary"
	"strconv"

	"tinygo.org/x/drivers/net2"
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
	// TCP requires a 12 byte pseudo-header to calculate the checksum
	PseudoHeaderInfo *IP
	// not implemented
	Options []byte
	Data    []byte
}

// UnmarshalFrame unmarshals a TCP frame from a byte slice, usually
// the byte slice contains IP data segment.
func (tcp *TCP) UnmarshalFrame(data []byte) error {
	_log("TCP:unmarshal")
	if len(data) < 20 {
		return errBufferTooSmall
	}
	tcp.Source = binary.BigEndian.Uint16(data[0:2])
	tcp.Destination = binary.BigEndian.Uint16(data[2:4])
	tcp.Ack = binary.BigEndian.Uint32(data[4:8])  // Seq/Ack switcheroo. Client/remote ack is our seq
	tcp.Seq = binary.BigEndian.Uint32(data[8:12]) // Ack will be our counter of amount of data recieved
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

func (tcp *TCP) MarshalFrame(data []byte) (uint16, error) {
	_log("TCP:marshal")
	if len(data) < int(tcp.FrameLength()) {
		return 0, errBufferTooSmall
	}
	binary.BigEndian.PutUint16(data[0:2], tcp.Source)
	binary.BigEndian.PutUint16(data[2:4], tcp.Destination)

	binary.BigEndian.PutUint32(data[4:8], tcp.Seq)
	binary.BigEndian.PutUint32(data[8:12], tcp.Ack)

	binary.BigEndian.PutUint16(data[12:14], tcp.Flags)

	binary.BigEndian.PutUint16(data[14:16], tcp.WindowSize)
	// skip checksum data[16:18]
	// zero out checksum field so as to ignore fields in checksum calculation
	data[16] = 0
	data[17] = 0
	binary.BigEndian.PutUint16(data[18:20], tcp.UrgentPtr)
	n := 20
	if tcp.DataOffset > 5 && tcp.Options != nil {
		n += copy(data[n:], tcp.Options)
	}
	if n%TCP_WORDLEN > 0 {
		n += TCP_WORDLEN - n%TCP_WORDLEN // [padding to fulfill TCP]
	}
	tcp.DataOffset = uint8(n / TCP_WORDLEN)
	data[12] &^= 0b1111_0000 // zero out past value
	data[12] |= tcp.DataOffset << 4

	n += copy(data[n:], tcp.Data)
	// Checksum preparations after data is copied
	if tcp.PseudoHeaderInfo.Version != IPHEADER_VERSION_4 {
		return uint16(n), errIPNotImplemented
	}

	ph := tcp.PseudoHeaderInfo
	// checksum IPv4 TCP packet and PseudoHeader
	var checksumPadding uint8
	if n%2 > 0 {
		checksumPadding++
	}
	chk := checksumRFC791(append(data[:n+int(checksumPadding)],
		ph.Source[0], ph.Source[1], ph.Source[2], ph.Source[3],
		ph.Destination[0], ph.Destination[1], ph.Destination[2], ph.Destination[3],
		0, ph.Protocol, uint8(n>>8), uint8(n),
	))
	binary.BigEndian.PutUint16(data[16:18], chk)
	return uint16(n), nil
}

// FrameLength for TCP frame. Should be called right after unmarshalling/marshalling TCP frame.
func (tcp *TCP) FrameLength() uint16 {
	return uint16(tcp.DataOffset)*TCP_WORDLEN + uint16(len(tcp.Options)+len(tcp.Data))
}

// Set TCP response header.
func (tcp *TCP) SetResponse(MAC net2.HardwareAddr) error {
	tcp.Destination, tcp.Source = tcp.Source, tcp.Destination
	tcp.SetFlags(TCPHEADER_FLAG_ACK)
	if tcp.PseudoHeaderInfo == nil {
		return errNoTCPPseudoHeader
	}

	if tcp.HasFlags(TCPHEADER_FLAG_SYN) {
		tcp.Seq = 2560 // uint32(checksumRFC791([]byte{byte(tcp.Ack)}))
		// set Maximum segment size (option 0x02) length 4 (0x04) to 1280 (0x0500)
		tcp.Options = []byte{0x02, 0x04, 0x05, 0x00}
		tcp.LastSeq = tcp.Seq
		tcp.Ack++
		tcp.WindowSize = 1400 // this is what EtherCard does?
		return nil
	}
	tcp.LastSeq = tcp.Seq
	tcp.Options = nil

	tcp.WindowSize = 1024 // TODO assign meaningful value to window size (or not?)
	// End TCP connection branch
	if tcp.HasFlags(TCPHEADER_FLAG_FIN) {
		tcp.ClearFlags(TCPHEADER_FLAG_FIN | TCPHEADER_FLAG_PSH)
		tcp.Data = nil
		tcp.Ack++
		return nil
	}
	tcp.ClearFlags(TCPHEADER_FLAG_FIN | TCPHEADER_FLAG_PSH)
	tcp.Ack += uint32(len(tcp.Data))
	tcp.Data = nil
	return nil
}

func (tcp *TCP) SetFlags(ORflags uint16) {
	if ORflags & ^TCPHEADER_FLAGS_MASK != 0 {
		panic("bad flag")
	}
	tcp.Flags |= uint16(ORflags)
}

// Has Flags returns true if ORflags are all set
func (tcp *TCP) HasFlags(ORflags uint16) bool { return (tcp.Flags & ORflags) == ORflags }
func (tcp *TCP) ClearFlags(ORflags uint16)    { tcp.Flags &^= ORflags }

// String Flag const
const flaglen = 3

var flagbuff = [2 + (flaglen+1)*9]byte{}

// StringFlags returns human readable flag string. i.e:
// "[SYN,ACK]".
//
// Beware use on AVR boards and other tiny places as it causes
// a lot of heap allocation and can quickly drain space.
func (tcp *TCP) StringFlags() string {
	const strflags = "FINSYNRSTPSHACKURGECECWRNS "
	n := 0
	for i := 0; i*3 < len(strflags)-flaglen; i++ {
		if tcp.HasFlags(1 << i) {
			if n == 0 {
				flagbuff[0] = '['
				n++
			} else {
				flagbuff[n] = ','
				n++
			}
			copy(flagbuff[n:n+3], []byte(strflags[i*flaglen:i*flaglen+flaglen]))
			n += 3
		}
	}
	if n > 0 {
		flagbuff[n] = ']'
		n++
	}
	return string(flagbuff[:n])
}

func (tcp *TCP) String() string {
	dlen := min(20, len(tcp.Data))
	return "TCP port " + u32toa(uint32(tcp.Source)) + "->" + u32toa(uint32(tcp.Destination)) +
		tcp.StringFlags() + "seq(" + strconv.Itoa(int(tcp.Seq-tcp.LastSeq)) + ")" + " data:" + string(tcp.Data[0:dlen])

}

func u32toa(u uint32) string {
	return strconv.Itoa(int(u))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
