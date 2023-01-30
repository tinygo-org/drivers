package lora

type Radio interface {
	Reset()
	Tx(pkt []uint8, timeoutMs uint32) error
	Rx(timeoutMs uint32) ([]uint8, error)
	SetFrequency(freq uint32)
	SetIqMode(mode uint8)
	SetCodingRate(cr uint8)
	SetBandwidth(bw uint8)
	SetCrc(enable bool)
	SetSpreadingFactor(sf uint8)
	SetPreambleLength(plen uint16)
	SetTxPower(txpow int8)
	SetSyncWord(syncWord uint16)
	SetPublicNetwork(enable bool)
	SetHeaderType(headerType uint8)
	LoraConfig(cnf Config)
}
