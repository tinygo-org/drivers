package enc28j60

import (
	"github.com/jkaflik/tinygo-w5500-driver/wiznet/net"
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
