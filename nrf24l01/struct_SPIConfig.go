package nrf24l01

type SPIConfig struct {
	Channel         uint8
	SenderAddress   []uint8
	ListenerAddress []uint8
	Sender          bool
	Bandwidth       bandWidth
	PowerOutput     powerOutput
	PackageSize     uint8
	ACK             bool
}

func SPIConfigDefault() SPIConfig {
	return SPIConfig{
		Channel:         76,
		SenderAddress:   []uint8{0, 0, 0, 0, 0},
		ListenerAddress: []uint8{0, 0, 0, 0, 0},
		Sender:          false,
		Bandwidth:       Bandwidth_256kbps,
		PowerOutput:     PowerOutput_0dBm,
		PackageSize:     32,
		ACK:             false,
	}
}
