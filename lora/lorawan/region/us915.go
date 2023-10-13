package region

import "tinygo.org/x/drivers/lora"

const (
	US915_DEFAULT_PREAMBLE_LEN     = 8
	US915_DEFAULT_TX_POWER_DBM     = 20
	US915_FREQUENCY_INCREMENT_DR_0 = 200000  // only for 125 kHz Bandwidth
	US915_FREQUENCY_INCREMENT_DR_4 = 1600000 // only for 500 kHz Bandwidth
)

type ChannelUS struct {
	frequency       uint32
	bandwidth       uint8
	spreadingFactor uint8
	codingRate      uint8
	preambleLength  uint16
	txPowerDBm      int8
}

// Getter functions
func (c *ChannelUS) Frequency() uint32      { return c.frequency }
func (c *ChannelUS) Bandwidth() uint8       { return c.bandwidth }
func (c *ChannelUS) SpreadingFactor() uint8 { return c.spreadingFactor }
func (c *ChannelUS) CodingRate() uint8      { return c.codingRate }
func (c *ChannelUS) PreambleLength() uint16 { return c.preambleLength }
func (c *ChannelUS) TxPowerDBm() int8       { return c.txPowerDBm }

// Set functions
// TODO: validate input
func (c *ChannelUS) SetFrequency(v uint32)      { c.frequency = v }
func (c *ChannelUS) SetBandwidth(v uint8)       { c.bandwidth = v }
func (c *ChannelUS) SetSpreadingFactor(v uint8) { c.spreadingFactor = v }
func (c *ChannelUS) SetCodingRate(v uint8)      { c.codingRate = v }
func (c *ChannelUS) SetPreambleLength(v uint16) { c.preambleLength = v }
func (c *ChannelUS) SetTxPowerDBm(v int8)       { c.txPowerDBm = v }

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

type RegionSettingsUS915 struct {
	joinRequestChannel *ChannelUS
	joinAcceptChannel  *ChannelUS
	uplinkChannel      *ChannelUS
}

func US915() *RegionSettingsUS915 {
	return &RegionSettingsUS915{
		joinRequestChannel: &ChannelUS{lora.MHz_902_3,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor10,
			lora.CodingRate4_5,
			US915_DEFAULT_PREAMBLE_LEN,
			US915_DEFAULT_TX_POWER_DBM},
		joinAcceptChannel: &ChannelUS{0,
			lora.Bandwidth_500_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_5,
			US915_DEFAULT_PREAMBLE_LEN,
			US915_DEFAULT_TX_POWER_DBM},
		uplinkChannel: &ChannelUS{lora.Mhz_903_0,
			lora.Bandwidth_500_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_5,
			US915_DEFAULT_PREAMBLE_LEN,
			US915_DEFAULT_TX_POWER_DBM},
	}
}

func (r *RegionSettingsUS915) JoinRequestChannel() Channel {
	return r.joinRequestChannel
}

func (r *RegionSettingsUS915) JoinAcceptChannel() Channel {
	return r.joinAcceptChannel
}

func (r *RegionSettingsUS915) UplinkChannel() Channel {
	return r.uplinkChannel
}
