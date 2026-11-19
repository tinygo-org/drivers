package nrf24l01

type Pin interface {
	High()
	Low()
	Get() bool
}
