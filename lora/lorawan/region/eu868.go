package region

import "tinygo.org/x/drivers/lora"

type RegionSettingsEU868 struct {
}

func EU868() *RegionSettingsEU868 {
	return &RegionSettingsEU868{}
}

func (r *RegionSettingsEU868) GetJoinRequestChannel() Channel {
	return Channel{868100000, lora.Bandwidth_125_0, lora.SpreadingFactor9, lora.CodingRate4_7}
}

func (r *RegionSettingsEU868) GetJoinAcceptChannel() Channel {
	return Channel{868100000, lora.Bandwidth_125_0, lora.SpreadingFactor9, lora.CodingRate4_7}
}

func (r *RegionSettingsEU868) GetUplinkChannel() Channel {
	return Channel{868100000, lora.Bandwidth_125_0, lora.SpreadingFactor9, lora.CodingRate4_7}
}
