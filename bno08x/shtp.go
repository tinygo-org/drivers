// SHTP specification found at https://www.ceva-ip.com/wp-content/uploads/SH-2-SHTP-Reference-Manual.pdf

package bno08x

import "encoding/binary"

// shtpHandler is a callback for handling SHTP channel data.
type shtpHandler func(payload []byte, timestamp uint32)

// shtp implements the Sensor Hub Transport Protocol layer.
type shtp struct {
	hal      *hal
	handlers map[uint8]shtpHandler
	seq      [8]uint8
	rx       [maxTransferIn]byte  // Reusable receive buffer
	tx       [maxTransferOut]byte // Reusable transmit buffer
}

func newSHTP(hal *hal) *shtp {
	return &shtp{
		hal:      hal,
		handlers: make(map[uint8]shtpHandler),
	}
}

// register registers a handler for a specific SHTP channel.
func (s *shtp) register(channel uint8, handler shtpHandler) {
	if handler == nil {
		delete(s.handlers, channel)
		return
	}
	s.handlers[channel] = handler
}

// send transmits a payload on the specified channel.
func (s *shtp) send(channel uint8, payload []byte) error {
	total := len(payload) + shtpHeaderLength
	if total > maxTransferOut {
		return errFrameTooLarge
	}

	// Use pre-allocated transmit buffer to avoid allocations
	frame := s.tx[:total]
	binary.LittleEndian.PutUint16(frame[0:2], uint16(total))
	frame[2] = channel
	frame[3] = s.seq[channel]
	s.seq[channel]++
	copy(frame[shtpHeaderLength:], payload)

	_, err := s.hal.write(frame)
	return err
}

// poll checks for and processes incoming SHTP packets.
// Returns true if a packet was processed, false if no data available.
func (s *shtp) poll() (bool, error) {
	n, timestamp, err := s.hal.read(s.rx[:])
	if err != nil {
		return false, err
	}
	if n == 0 {
		return false, nil
	}

	packet := s.rx[:n]
	length := int(binary.LittleEndian.Uint16(packet[0:2]) & ^uint16(continueMask))
	if length > n {
		length = n
	}
	if length < shtpHeaderLength {
		return false, nil
	}

	channel := packet[2]
	// seq := packet[3] // sequence number, not currently validated
	payload := packet[shtpHeaderLength:length]

	if handler := s.handlers[channel]; handler != nil {
		handler(payload, timestamp)
	}

	return true, nil
}
