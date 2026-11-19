package nrf24l01

func (s *nrf24l01) WriteRegister(reg, value uint8) error {
	s.up()
	defer s.down()
	tx := s.getBuffer(s.txCMD)
	tx[0] = W_REGISTER | reg
	tx[1] = value
	txErr := s.spi.Tx(tx, nil)
	if txErr != nil {
		return txErr
	}
	return nil
}
