package region

import "tinygo.org/x/drivers/lora"

const (
	EU868_DEFAULT_PREAMBLE_LEN = 8
	EU868_DEFAULT_TX_POWER_DBM = 20
)

type RegionSettingsEU868 struct {
	joinRequestChannel *Channel
	joinAcceptChannel  *Channel
	uplinkChannel      *Channel
}

func EU868() *RegionSettingsEU868 {
	return &RegionSettingsEU868{
		joinRequestChannel: &Channel{lora.MHz_868_1,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_7,
			EU868_DEFAULT_PREAMBLE_LEN,
			EU868_DEFAULT_TX_POWER_DBM},
		joinAcceptChannel: &Channel{lora.MHz_868_1,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_7,
			EU868_DEFAULT_PREAMBLE_LEN,
			EU868_DEFAULT_TX_POWER_DBM},
		uplinkChannel: &Channel{lora.MHz_868_1,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_7,
			EU868_DEFAULT_PREAMBLE_LEN,
			EU868_DEFAULT_TX_POWER_DBM},
	}
}

func (r *RegionSettingsEU868) JoinRequestChannel() *Channel {
	return r.joinRequestChannel
}

func (r *RegionSettingsEU868) JoinAcceptChannel() *Channel {
	return r.joinAcceptChannel
}

func (r *RegionSettingsEU868) UplinkChannel() *Channel {
	return r.uplinkChannel
}
