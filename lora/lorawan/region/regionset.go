package region

type Channel struct {
	Frequency       uint32
	Bandwidth       uint8
	SpreadingFactor uint8
	CodingRate      uint8
	PreambleLength  uint16
	TxPowerDBm      int8
}

type RegionSettings interface {
	JoinRequestChannel() *Channel
	JoinAcceptChannel() *Channel
	UplinkChannel() *Channel
}
