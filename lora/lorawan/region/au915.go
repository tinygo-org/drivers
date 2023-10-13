package region

import "tinygo.org/x/drivers/lora"

const (
	AU915_DEFAULT_PREAMBLE_LEN = 8
	AU915_DEFAULT_TX_POWER_DBM = 20
)

type ChannelAU struct {
	frequency       uint32
	bandwidth       uint8
	spreadingFactor uint8
	codingRate      uint8
	preambleLength  uint16
	txPowerDBm      int8
}

// Getter functions
func (c *ChannelAU) Frequency() uint32      { return c.frequency }
func (c *ChannelAU) Bandwidth() uint8       { return c.bandwidth }
func (c *ChannelAU) SpreadingFactor() uint8 { return c.spreadingFactor }
func (c *ChannelAU) CodingRate() uint8      { return c.codingRate }
func (c *ChannelAU) PreambleLength() uint16 { return c.preambleLength }
func (c *ChannelAU) TxPowerDBm() int8       { return c.txPowerDBm }

// Set functions
func (c *ChannelAU) SetFrequency(v uint32)      { c.frequency = v }
func (c *ChannelAU) SetBandwidth(v uint8)       { c.bandwidth = v }
func (c *ChannelAU) SetSpreadingFactor(v uint8) { c.spreadingFactor = v }
func (c *ChannelAU) SetCodingRate(v uint8)      { c.codingRate = v }
func (c *ChannelAU) SetPreambleLength(v uint16) { c.preambleLength = v }
func (c *ChannelAU) SetTxPowerDBm(v int8)       { c.txPowerDBm = v }

func (c *ChannelAU) Next() bool {
	return false
}

type RegionSettingsAU915 struct {
	joinRequestChannel *ChannelAU
	joinAcceptChannel  *ChannelAU
	uplinkChannel      *ChannelAU
}

func AU915() *RegionSettingsAU915 {
	return &RegionSettingsAU915{
		joinRequestChannel: &ChannelAU{lora.MHz_916_8,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_5,
			AU915_DEFAULT_PREAMBLE_LEN,
			AU915_DEFAULT_TX_POWER_DBM},
		joinAcceptChannel: &ChannelAU{lora.MHz_923_3,
			lora.Bandwidth_500_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_5,
			AU915_DEFAULT_PREAMBLE_LEN,
			AU915_DEFAULT_TX_POWER_DBM},
		uplinkChannel: &ChannelAU{lora.MHz_916_8,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_5,
			AU915_DEFAULT_PREAMBLE_LEN,
			AU915_DEFAULT_TX_POWER_DBM},
	}
}

func Next(c *ChannelAU) bool {
	return false
}

func (r *RegionSettingsAU915) JoinRequestChannel() Channel {
	return r.joinRequestChannel
}

func (r *RegionSettingsAU915) JoinAcceptChannel() Channel {
	return r.joinAcceptChannel
}

func (r *RegionSettingsAU915) UplinkChannel() Channel {
	return r.uplinkChannel
}
