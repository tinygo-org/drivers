package nrf24l01

type NRF24L01 interface {
	FlushRX() error
	FlushTX() error
	HasIRQ() bool
	HasData() bool
	ReadMultiRegister(uint8, int) []uint8
	ReadPayload() ([]uint8, error)
	ReadRegister(uint8) uint8
	Send([]uint8) error
	WriteMultiRegister(uint8, []uint8) error
	WritePayload([]uint8) error
	WriteRegister(uint8, uint8) error
}
