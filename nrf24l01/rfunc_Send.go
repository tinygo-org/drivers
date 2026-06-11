package nrf24l01

import (
	"errors"
	"time"
)

// Opinionated send method with the first bit indicating the data length.
// Fist bit is the length of data, therefore data must be smaller than set package size
func (s *nrf24l01) Send(data []uint8) error {
	if len(data) > s.packageSize-1 {
		return errors.New("package size too big")
	}
	fifoStatus := s.ReadRegister(FIFO_STATUS)
	if fifoStatus&0x20 != 0 { //TX_FULL
		return errors.New("TX_FULL")
	}
	s.ce.Low()
	s.WriteRegister(STATUS, 0x70)
	tx := s.getBuffer(s.tx)
	tx[0] = W_TX_PAYLOAD
	tx[1] = uint8(len(data))
	for i, d := range data {
		tx[i+2] = d
	}
	writeErr := s.WritePayload(tx)
	if writeErr != nil {
		return writeErr
	}
	s.ce.High()
	time.Sleep(time.Microsecond * 15)
	s.ce.Low()
	// wait for TX_DS or MAX_RT
	for {
		status := s.ReadRegister(STATUS)
		if status&0x20 != 0 { // TX_DS
			s.WriteRegister(STATUS, 0x20)
			return nil
		}
		if status&0x10 != 0 { // MAX_RT
			s.WriteRegister(STATUS, 0x10)
			s.FlushTX()
			return errors.New("MAX_RT")
		}
	}
}
