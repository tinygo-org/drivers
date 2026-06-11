package nrf24l01

func (s *nrf24l01) WriteMultiRegister(reg uint8, values []uint8) error {
	s.up()
	defer s.down()
	tx := s.getBuffer(s.tx)
	tx[0] = W_REGISTER | reg
	for i, v := range values {
		tx[i+1] = v
	}
	txErr := s.spi.Tx(tx, nil)
	if txErr != nil {
		return txErr
	}
	return nil
}
