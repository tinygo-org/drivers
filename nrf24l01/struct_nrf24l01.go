package nrf24l01

type nrf24l01 struct {
	packageSize     int
	channel         uint8
	senderAddress   []uint8
	listenerAddress []uint8
	addressWidth    uint8
	txCMD           []uint8
	rxCMD           []uint8
	tx              []uint8
	rx              []uint8

	spi SPI
	ce  Pin
	csn Pin
	irq Pin
}
