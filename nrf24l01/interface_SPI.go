package nrf24l01

type SPI interface {
	Tx([]uint8, []uint8) error
	Transfer(uint8) (uint8, error)
}
