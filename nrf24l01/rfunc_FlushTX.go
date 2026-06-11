package nrf24l01

func (s *nrf24l01) FlushTX() error {
	s.up()
	defer s.down()
	tx := s.getBuffer(s.txCMD)
	tx[0] = FLUSH_TX
	txErr := s.spi.Tx(tx, nil)
	return txErr
}
