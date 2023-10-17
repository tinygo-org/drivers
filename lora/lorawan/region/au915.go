package region

import "tinygo.org/x/drivers/lora"

const (
	AU915_DEFAULT_PREAMBLE_LEN = 8
	AU915_DEFAULT_TX_POWER_DBM = 20
)

type ChannelAU struct {
	channel
}

func (c *ChannelAU) Next() bool {
	return false
}

type SettingsAU915 struct {
	settings
}

func AU915() *SettingsAU915 {
	return &SettingsAU915{settings: settings{
		joinRequestChannel: &ChannelAU{channel: channel{lora.MHz_916_8,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_5,
			AU915_DEFAULT_PREAMBLE_LEN,
			AU915_DEFAULT_TX_POWER_DBM}},
		joinAcceptChannel: &ChannelAU{channel: channel{lora.MHz_923_3,
			lora.Bandwidth_500_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_5,
			AU915_DEFAULT_PREAMBLE_LEN,
			AU915_DEFAULT_TX_POWER_DBM}},
		uplinkChannel: &ChannelAU{channel: channel{lora.MHz_916_8,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_5,
			AU915_DEFAULT_PREAMBLE_LEN,
			AU915_DEFAULT_TX_POWER_DBM}},
	}}
}

func Next(c *ChannelAU) bool {
	return false
}
