package nrf24l01

import (
	"errors"
	"strconv"
)

// Opinionated read method.
// Fist bit is the length of data, therefore data is smaller than set package size
// Use ReadMultiRegister() for custom framing
func (s *nrf24l01) ReadPayload() ([]uint8, error) {
	s.ce.Low()
	defer s.ce.High()
	payload := s.ReadMultiRegister(R_RX_PAYLOAD, s.packageSize)
	payloadLength := payload[1]
	if payloadLength == 0 || payloadLength > uint8(s.packageSize) {
		s := string(payload) + " out of bounds " + strconv.Itoa(int(payloadLength))
		return payload, errors.New(s)
	}
	if len(payload) < int(payloadLength)+2 {
		s := "payload to small:" + strconv.Itoa(len(payload))
		return payload, errors.New(s)
	}
	message := payload[1 : payloadLength+2]
	return message, nil
}
