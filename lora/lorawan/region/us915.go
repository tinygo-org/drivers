package region

import "tinygo.org/x/drivers/lora"

const (
	US915_DEFAULT_PREAMBLE_LEN = 8
	US915_DEFAULT_TX_POWER_DBM = 20
)

type RegionSettingsUS915 struct {
	joinRequestChannel *Channel
	joinAcceptChannel  *Channel
	uplinkChannel      *Channel
}

// see https://www.thethingsnetwork.org/docs/lorawan/regional-parameters/#us902-928-ism-band

func US915() *RegionSettingsUS915 {
	return &RegionSettingsUS915{
		joinRequestChannel: &Channel{lora.MHz_902_3,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor10,
			lora.CodingRate4_5,
			US915_DEFAULT_PREAMBLE_LEN,
			US915_DEFAULT_TX_POWER_DBM},
		joinAcceptChannel: &Channel{lora.MHz_923_3,
			lora.Bandwidth_500_0,
			lora.SpreadingFactor7,
			lora.CodingRate4_5,
			US915_DEFAULT_PREAMBLE_LEN,
			US915_DEFAULT_TX_POWER_DBM},
		uplinkChannel: &Channel{lora.MHz_902_3,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor7,
			lora.CodingRate4_5,
			US915_DEFAULT_PREAMBLE_LEN,
			US915_DEFAULT_TX_POWER_DBM},
	}
}

func (r *RegionSettingsUS915) JoinRequestChannel() *Channel {
	return r.joinRequestChannel
}

func (r *RegionSettingsUS915) JoinAcceptChannel() *Channel {
	return r.joinAcceptChannel
}

func (r *RegionSettingsUS915) UplinkChannel() *Channel {
	return r.uplinkChannel
}
