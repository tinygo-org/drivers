package region

import "tinygo.org/x/drivers/lora"

const (
	EU868_DEFAULT_PREAMBLE_LEN = 8
	EU868_DEFAULT_TX_POWER_DBM = 20
)

type ChannelEU struct {
	frequency       uint32
	bandwidth       uint8
	spreadingFactor uint8
	codingRate      uint8
	preambleLength  uint16
	txPowerDBm      int8
}

// Getter functions
func (c *ChannelEU) Frequency() uint32      { return c.frequency }
func (c *ChannelEU) Bandwidth() uint8       { return c.bandwidth }
func (c *ChannelEU) SpreadingFactor() uint8 { return c.spreadingFactor }
func (c *ChannelEU) CodingRate() uint8      { return c.codingRate }
func (c *ChannelEU) PreambleLength() uint16 { return c.preambleLength }
func (c *ChannelEU) TxPowerDBm() int8       { return c.txPowerDBm }

func (c *ChannelEU) Next() bool {
	return false
}

type RegionSettingsEU868 struct {
	joinRequestChannel *ChannelEU
	joinAcceptChannel  *ChannelEU
	uplinkChannel      *ChannelEU
}

func EU868() *RegionSettingsEU868 {
	return &RegionSettingsEU868{
		joinRequestChannel: &ChannelEU{lora.MHz_868_1,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_7,
			EU868_DEFAULT_PREAMBLE_LEN,
			EU868_DEFAULT_TX_POWER_DBM},
		joinAcceptChannel: &ChannelEU{lora.MHz_868_1,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_7,
			EU868_DEFAULT_PREAMBLE_LEN,
			EU868_DEFAULT_TX_POWER_DBM},
		uplinkChannel: &ChannelEU{lora.MHz_868_1,
			lora.Bandwidth_125_0,
			lora.SpreadingFactor9,
			lora.CodingRate4_7,
			EU868_DEFAULT_PREAMBLE_LEN,
			EU868_DEFAULT_TX_POWER_DBM},
	}
}

func (r *RegionSettingsEU868) JoinRequestChannel() Channel {
	return r.joinRequestChannel
}

func (r *RegionSettingsEU868) JoinAcceptChannel() Channel {
	return r.joinAcceptChannel
}

func (r *RegionSettingsEU868) UplinkChannel() Channel {
	return r.uplinkChannel
}
