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
		joinRequestChannel: &Channel{lora.MHz_916_8,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_5,
			AU915_DEFAULT_PREAMBLE_LEN,
			AU915_DEFAULT_TX_POWER_DBM},
		joinAcceptChannel: &Channel{lora.MHz_923_3,
			lora.Bandwidth_500_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_5,
			AU915_DEFAULT_PREAMBLE_LEN,
			AU915_DEFAULT_TX_POWER_DBM},
		uplinkChannel: &Channel{lora.MHz_916_8,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_5,
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
