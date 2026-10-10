package nrf24l01

func (s *nrf24l01) up() {
	s.csn.Low()
}
