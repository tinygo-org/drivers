package region

import "tinygo.org/x/drivers/lora"

const (
	EU868_DEFAULT_PREAMBLE_LEN = 8
	EU868_DEFAULT_TX_POWER_DBM = 20
)

type ChannelEU struct {
	channel
}

func (c *ChannelEU) Next() bool {
	return false
}

type SettingsEU868 struct {
	settings
}

func EU868() *SettingsEU868 {
	return &SettingsEU868{settings: settings{
		joinRequestChannel: &ChannelEU{channel: channel{lora.MHz_868_1,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_7,
			EU868_DEFAULT_PREAMBLE_LEN,
			EU868_DEFAULT_TX_POWER_DBM}},
		joinAcceptChannel: &ChannelEU{channel: channel{lora.MHz_868_1,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_7,
			EU868_DEFAULT_PREAMBLE_LEN,
			EU868_DEFAULT_TX_POWER_DBM}},
		uplinkChannel: &ChannelEU{channel: channel{lora.MHz_868_1,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_7,
			EU868_DEFAULT_PREAMBLE_LEN,
			EU868_DEFAULT_TX_POWER_DBM}},
	}}
}
