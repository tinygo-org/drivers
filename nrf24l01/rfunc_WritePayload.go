package nrf24l01

import (
	"errors"
)

func (s *nrf24l01) WritePayload(data []uint8) error {
	if len(data) > s.packageSize+1 {
		return errors.New("packagesize too big")
	}
	s.up()
	defer s.down()
	data[0] = W_TX_PAYLOAD
	txErr := s.spi.Tx(data, nil)
	if txErr != nil {
		return txErr
	}
	return nil
}
