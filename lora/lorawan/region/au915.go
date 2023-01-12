package region

import "tinygo.org/x/drivers/lora"

type RegionSettingsAU915 struct {
}

func AU915() *RegionSettingsAU915 {
	return &RegionSettingsAU915{}
}

func (r *RegionSettingsAU915) GetJoinRequestChannel() Channel {
	return Channel{916800000, lora.Bandwidth_125_0, lora.SpreadingFactor9, lora.CodingRate4_5}
}

func (r *RegionSettingsAU915) GetJoinAcceptChannel() Channel {
	return Channel{923300000, lora.Bandwidth_500_0, lora.SpreadingFactor9, lora.CodingRate4_5}
}

func (r *RegionSettingsAU915) GetUplinkChannel() Channel {
	return Channel{868500000, lora.Bandwidth_125_0, lora.SpreadingFactor9, lora.CodingRate4_5}
}
