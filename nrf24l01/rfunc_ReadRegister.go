package nrf24l01

func (s *nrf24l01) ReadRegister(reg uint8) uint8 {
	s.up()
	defer s.down()
	tx := s.getBuffer(s.txCMD)
	tx[0] = R_REGISTER | reg
	tx[1] = NOP
	rx := s.getBuffer(s.rxCMD)
	txErr := s.spi.Tx(tx, rx)
	if txErr != nil {
		println("ERROR: SPI: ", txErr.Error())
	}
	return rx[1]
}
