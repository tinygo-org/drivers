package nrf24l01

type bandWidth uint8

const (
	Bandwidth_256kbps bandWidth = iota
	Bandwidth_1Mbps   bandWidth = iota
	Bandwidth_2Mbps   bandWidth = iota
)
