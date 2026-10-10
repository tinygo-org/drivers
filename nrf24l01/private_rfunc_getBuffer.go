package nrf24l01

func (s *nrf24l01) getBuffer(b []uint8) []uint8 {
	for i, _ := range b {
		b[i] = 0x00
	}
	return b
}
