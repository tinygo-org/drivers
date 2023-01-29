package region

import "tinygo.org/x/drivers/lora"

const (
	AU915_DEFAULT_PREAMBLE_LEN = 8
	AU915_DEFAULT_TX_POWER_DBM = 20
)

type RegionSettingsAU915 struct {
	joinRequestChannel *Channel
	joinAcceptChannel  *Channel
	uplinkChannel      *Channel
}

func AU915() *RegionSettingsAU915 {
	return &RegionSettingsAU915{
		joinRequestChannel: &Channel{916800000,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_7,
			AU915_DEFAULT_PREAMBLE_LEN,
			AU915_DEFAULT_TX_POWER_DBM},
		joinAcceptChannel: &Channel{923300000,
			lora.Bandwidth_500_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_7,
			AU915_DEFAULT_PREAMBLE_LEN,
			AU915_DEFAULT_TX_POWER_DBM},
		uplinkChannel: &Channel{868500000,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_7,
			AU915_DEFAULT_PREAMBLE_LEN,
			AU915_DEFAULT_TX_POWER_DBM},
	}
}

func (r *RegionSettingsAU915) JoinRequestChannel() *Channel {
	return r.joinRequestChannel
}

func (r *RegionSettingsAU915) JoinAcceptChannel() *Channel {
	return r.joinAcceptChannel
}

func (r *RegionSettingsAU915) UplinkChannel() *Channel {
	return r.uplinkChannel
}
