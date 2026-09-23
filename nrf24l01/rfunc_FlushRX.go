package nrf24l01

func (s *nrf24l01) FlushRX() error {
	s.up()
	defer s.down()
	tx := s.getBuffer(s.txCMD)
	tx[0] = FLUSH_RX
	txErr := s.spi.Tx(tx, nil)
	return txErr
}
