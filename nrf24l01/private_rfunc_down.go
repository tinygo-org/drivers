package nrf24l01

func (s *nrf24l01) down() {
	s.csn.High()
}
