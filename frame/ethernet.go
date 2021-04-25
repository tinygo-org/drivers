package frame

// The code below was taken from github.com/mdlayher/ethernet and adapted for embedded use
// All credit to mdlayher and the ethernet Authors

import (
	"encoding/binary"

	"tinygo.org/x/drivers/net"

	"tinygo.org/x/drivers/encoding/hex"
)

const (
	// minPayload is the minimum payload size for an Ethernet frame, assuming
	// that no 802.1Q VLAN tags are present.
	minPayload = 46
)

var (
	// Broadcast is a special hardware address which indicates a Frame should
	// be sent to every device on a given LAN segment.
	Broadcast = net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
)

// An EtherType is a value used to identify an upper layer protocol
// encapsulated in a Frame.
//
// A list of IANA-assigned EtherType values may be found here:
// http://www.iana.org/assignments/ieee-802-numbers/ieee-802-numbers.xhtml.
type EtherType uint16

// Common EtherType values frequently used in a Frame.
const (
	EtherTypeIPv4 EtherType = 0x0800
	EtherTypeARP  EtherType = 0x0806
	EtherTypeIPv6 EtherType = 0x86DD

	// EtherTypeVLAN and EtherTypeServiceVLAN are used as 802.1Q Tag Protocol
	// Identifiers (TPIDs).
	EtherTypeVLAN        EtherType = 0x8100
	EtherTypeServiceVLAN EtherType = 0x88a8
)

// A Frame is an IEEE 802.3 Ethernet II frame.  A Frame contains information
// such as source and destination hardware addresses, zero or more optional
// 802.1Q VLAN tags, an EtherType, and payload data.
type Ethernet struct {
	// EtherType is a value used to identify an upper layer protocol
	// encapsulated in this Frame.
	EtherType EtherType

	// Destination specifies the destination hardware address for this Frame.
	//
	// If this address is set to Broadcast, the Frame will be sent to every
	// device on a given LAN segment.
	Destination net.HardwareAddr

	// Source specifies the source hardware address for this Frame.
	//
	// Typically, this is the hardware address of the network interface used to
	// send this Frame.
	Source net.HardwareAddr

	// Payload is a variable length data payload encapsulated by this Frame.
	Payload []byte
	// Save subframers
	Framer
}

func (f *Ethernet) MarshalFrame(buff []byte) (uint16, error) {
	if uint16(len(buff)) < f.FrameLength() {
		return 0, ErrBufferTooSmall
	}
	copy(buff[0:6], f.Destination)
	copy(buff[6:12], f.Source)
	n := 12
	binary.BigEndian.PutUint16(buff[n:n+2], uint16(f.EtherType))
	n += 2
	if f.Framer != nil {
		m, err := f.Framer.MarshalFrame(buff[n:])
		return uint16(n) + m, err
	}
	n += copy(buff[n:], f.Payload)
	return uint16(n), nil
}

// UnmarshalFrame unmarshals binary data in buffer into Ethernet Frame.
// If Ethernet has a non-nil Framer it will call Framer's UnmarshalFrame
// on Ethernet Payload.
//
// Be sure to call UnmarshalFrame on the packet length recieved and no more.
// Calling UnmarshalFrame on a whole buffer can cause memory segmentation
// failures when attempting to marshal using Framers with MarshalFrame.
func (f *Ethernet) UnmarshalFrame(buff []byte) error {
	_log("ETHunmarshal frame")
	n, err := f.UnmarshalBinary(buff)
	if err != nil {
		return err
	}
	_log("ETH:subframe?")
	if f.Framer != nil {
		_log("ETH:subframe")
		return f.Framer.UnmarshalFrame(buff[n:])
	}
	return nil
}

// FrameLength Returns padded Ethernet frame length after querying attached Framer.
// If no Framer found then returns EthernetFrame length + payload slice length
func (f *Ethernet) FrameLength() uint16 {
	paylen := uint16(len(f.Payload))
	// If payload is less than the required minimum length, we zero-pad up to
	// the required minimum length
	if f.Framer != nil {
		paylen = f.Framer.FrameLength()
	}
	if paylen < minPayload {
		paylen = minPayload
	}
	// 6 bytes: destination hardware address
	// 6 bytes: source hardware address
	// N bytes: VLAN tags (if present)
	// 2 bytes: EtherType
	// N bytes: payload length (may be padded)
	return 6 + 6 + 2 + paylen
}

// UnmarshalBinary unmarshals a byte slice into an Ethernet Frame. Does not unmarshal
// Framer field. Returns length of Ethernet header marshalled.
//
// Payload is marshalled into a slice which points to original buffer.
func (f *Ethernet) UnmarshalBinary(buff []byte) (uint16, error) {
	_log("ETHunmarshal bin")
	bufflen := uint16(len(buff))
	// Verify that both hardware addresses and a single EtherType are present
	if bufflen < 14 {
		_log("ETH:umbin fail")
		return 0, ErrBufferTooSmall
	}

	// Track offset in packet for reading data
	var n uint16 = 14

	// Continue looping and parsing VLAN tags until no more VLAN EtherType
	// values are detected
	// VLAN NOT IMPLEMENTED

	f.EtherType = EtherType(binary.BigEndian.Uint16(buff[n-2 : n]))
	// Future stuff: do VLAN implementation

	// Allocate single byte slice to store destination and source hardware
	// addresses, and payload
	bb := make([]byte, 6+6)
	copy(bb[0:12], buff[0:12])
	f.Destination = bb[0:6]
	f.Source = bb[6:12]
	f.Payload = buff[n:]
	return n, nil
}

// setResponse with own Macaddress. If etherType is equal to 0, etherType is not changed
func (f *Ethernet) SetResponse(MAC net.HardwareAddr) error {
	f.Destination = f.Source
	f.Source = MAC
	if f.Framer != nil {
		return f.Framer.SetResponse(MAC)
	}
	return nil
}

func (f *Ethernet) String() string {
	return "dst: " + f.Destination.String() + "\n" +
		"src: " + f.Source.String() + "\n" +
		"etype: " + string(append(hex.Byte(byte(f.EtherType>>8)), hex.Byte(byte(f.EtherType))...)) + "\n" +
		"Ether payload: " + string(hex.Bytes(f.Payload[:]))
}
