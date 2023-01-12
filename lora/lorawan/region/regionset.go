package region

type Channel struct {
	Frequency       uint32
	Bandwidth       uint8
	SpreadingFactor uint8
	CodingRate      uint8
}

type RegionSettings interface {
	GetJoinRequestChannel() Channel
	GetJoinAcceptChannel() Channel
	GetUplinkChannel() Channel
}
