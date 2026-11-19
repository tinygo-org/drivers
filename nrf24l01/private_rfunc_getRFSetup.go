package nrf24l01

func (s *nrf24l01) getRFSetup(config SPIConfig) uint8 {
	var rfSetup uint8 = 0x00

	// LNA_HCURR
	rfSetup |= 1 << 1

	// Bandwidth
	switch config.Bandwidth {
	case Bandwidth_256kbps:
		rfSetup |= 1 << 6
		break
	case Bandwidth_1Mbps:
		break
	case Bandwidth_2Mbps:
		rfSetup |= 1 << 4
		break
	}

	// PowerOutput
	switch config.PowerOutput {
	case PowerOutput_n18dBm:
		break
	case PowerOutput_n12dBm:
		rfSetup |= 1 << 2
		break
	case PowerOutput_n6dBm:
		rfSetup |= 1 << 3
		break
	case PowerOutput_0dBm:
		rfSetup |= 1 << 2
		rfSetup |= 1 << 3
		break
	}
	return rfSetup
}
