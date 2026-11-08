package nrf24l01

type powerOutput uint8

const (
	PowerOutput_n18dBm powerOutput = iota
	PowerOutput_n12dBm powerOutput = iota
	PowerOutput_n6dBm  powerOutput = iota
	PowerOutput_0dBm   powerOutput = iota
)
