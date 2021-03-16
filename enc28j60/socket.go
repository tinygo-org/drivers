package enc28j60

import (
	"errors"
	"math/rand"

	"github.com/jkaflik/tinygo-w5500-driver/wiznet/net"
)

const (
	preambleSize    = 7
	startFrameDelim = 1
	// Ethernet frame offset
	ICOffset = preambleSize + startFrameDelim
	// payload offset (where ARP frame begins)
	efPayloadOffset = 22
)

type socketMode uint8

const (
	arpMode = iota
)

// Socket w5500
type Socket struct {
	d      *Dev
	efType uint16
	Port   uint16
	Num    uint8
	mode   socketMode

	dstMacaddr    net.HardwareAddr
	receivedSize  uint16
	receiveOffset uint16
}

var (
	errOutOfBound = errors.New("out of buff bounds")
	errMacAddr    = errors.New("bad mac addr")
	errBadIP      = errors.New("bad ip")
)

func (s Socket) payloadwrite(offset uint16, buff []byte) error {
	return s.bufwrite(efPayloadOffset, buff)
}
func (s Socket) bufwrite(offset uint16, buff []byte) error {
	offset -= ICOffset
	if offset+uint16(len(buff)) > uint16(len(s.d.buffer)) {
		return errOutOfBound
	}
	copy(s.d.buffer[offset:offset+uint16(len(buff))], buff)
	return nil
}

/* Ethernet Frame

 0        6        7         13         19        21
| 7 bytes  | 1 byte |  6 bytes |  6 bytes | 2 bytes | 46-1500 bytes /.../ | 4 bytes   |
| Preamble | SFD    | Dst Addr | Src Addr | Type    | PAYLOAD       /.../ | FCS (CRC) |

the enc28j60 takes care of the preamble and the SFD. It is by default configured to take care
of the FCS too.
*/

// ethernet frame write mac addresses to buffer
func (s Socket) efWriteHWAdresses() error {
	if s.dstMacaddr == nil || len(s.dstMacaddr) != 6 {
		return errMacAddr
	}
	s.bufwrite(7, append(s.dstMacaddr, s.d.macaddr...))
	return nil
}

func (s Socket) efWriteType() error {
	return s.bufwrite(19, []byte{uint8(s.efType), uint8(s.efType >> 8)})
}

// Open supports ARP for network discovery
func (s *Socket) Open(protocol string, port uint16) error {
	if !validIP(s.d.broadcastip) || !validIP(s.d.myip) {
		return errBadIP
	}
	switch protocol {
	case "arp": //address resolution protocol
		s.dstMacaddr = net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
		s.efType = efARPType
		s.mode = arpMode
		s.writeARP()
	}

	if port == 0 { // pick random local port instead
		s.Port = 49152 + uint16(rand.Intn(16383))
	}
	return nil
}

func validIP(ip net.IP) bool {
	if ip == nil || (len(ip) != 4 && len(ip) != 16) {
		return false
	}
	return true
}
