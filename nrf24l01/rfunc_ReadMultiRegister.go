package nrf24l01

func (s *nrf24l01) ReadMultiRegister(reg uint8, length int) []uint8 {
	s.up()
	defer s.down()

	tx := s.getBuffer(s.tx)
	rx := s.getBuffer(s.rx)
	tx[0] = reg
	for i := 1; i < length+1; i++ {
		tx[i] = NOP
	}
	txErr := s.spi.Tx(tx, rx)
	if txErr != nil {
		println("SPI error: ", txErr.Error())
	}
	return rx
}
