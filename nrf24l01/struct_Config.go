package nrf24l01

type Config struct {
	SPI       SPI
	CE        Pin
	CSN       Pin
	IRQ       Pin
	SPIConfig SPIConfig
}
