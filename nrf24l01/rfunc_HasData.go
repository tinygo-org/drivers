package nrf24l01

func (s *nrf24l01) HasData() bool {
	return s.ReadRegister(FIFO_STATUS)&0x01 == 0
}
