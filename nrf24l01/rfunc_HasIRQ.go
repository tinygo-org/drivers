package nrf24l01

func (s *nrf24l01) HasIRQ() bool {
	return !s.irq.Get()
}
