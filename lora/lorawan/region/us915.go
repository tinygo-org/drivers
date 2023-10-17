package region

import "tinygo.org/x/drivers/lora"

const (
	US915_DEFAULT_PREAMBLE_LEN     = 8
	US915_DEFAULT_TX_POWER_DBM     = 20
	US915_FREQUENCY_INCREMENT_DR_0 = 200000  // only for 125 kHz Bandwidth
	US915_FREQUENCY_INCREMENT_DR_4 = 1600000 // only for 500 kHz Bandwidth
)

type ChannelUS struct {
	channel
}

func (c *ChannelUS) Next() bool {
	switch c.Bandwidth() {
	case lora.Bandwidth_125_0:
		freq, ok := stepFrequency125(c.frequency)
		if ok {
			c.frequency = freq
		} else {
			c.frequency = lora.Mhz_903_0
			c.bandwidth = lora.Bandwidth_500_0
		}
	case lora.Bandwidth_500_0:
		freq, ok := stepFrequency500(c.frequency)
		if ok {
			c.frequency = freq
		} else {
			// there are no more frequencies to check after sweeping all 8 500 kHz channels
			return false
		}
	}

	return true
}

func stepFrequency125(freq uint32) (uint32, bool) {
	f := freq + US915_FREQUENCY_INCREMENT_DR_0
	if f >= lora.MHZ_915_0 {
		return 0, false
	}

	return f, true
}

func stepFrequency500(freq uint32) (uint32, bool) {
	f := freq + US915_FREQUENCY_INCREMENT_DR_4
	if f >= lora.MHZ_915_0 {
		return 0, false
	}

	return f, true
}

type SettingsUS915 struct {
	settings
}

func US915() *SettingsUS915 {
	return &SettingsUS915{settings: settings{
		joinRequestChannel: &ChannelUS{channel: channel{lora.MHz_902_3,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor10,
			lora.CodingRate4_5,
			US915_DEFAULT_PREAMBLE_LEN,
			US915_DEFAULT_TX_POWER_DBM}},
		joinAcceptChannel: &ChannelUS{channel: channel{0,
			lora.Bandwidth_500_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_5,
			US915_DEFAULT_PREAMBLE_LEN,
			US915_DEFAULT_TX_POWER_DBM}},
		uplinkChannel: &ChannelUS{channel: channel{lora.Mhz_903_0,
			lora.Bandwidth_500_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_5,
			US915_DEFAULT_PREAMBLE_LEN,
			US915_DEFAULT_TX_POWER_DBM}},
	}}
}
