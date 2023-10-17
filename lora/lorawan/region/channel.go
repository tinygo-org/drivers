package region

type Channel interface {
	Next() bool
	Frequency() uint32
	Bandwidth() uint8
	SpreadingFactor() uint8
	CodingRate() uint8
	PreambleLength() uint16
	TxPowerDBm() int8
	SetFrequency(v uint32)
	SetBandwidth(v uint8)
	SetSpreadingFactor(v uint8)
	SetCodingRate(v uint8)
	SetPreambleLength(v uint16)
	SetTxPowerDBm(v int8)
}

type channel struct {
	frequency       uint32
	bandwidth       uint8
	spreadingFactor uint8
	codingRate      uint8
	preambleLength  uint16
	txPowerDBm      int8
}

// Getter functions
func (c *channel) Frequency() uint32      { return c.frequency }
func (c *channel) Bandwidth() uint8       { return c.bandwidth }
func (c *channel) SpreadingFactor() uint8 { return c.spreadingFactor }
func (c *channel) CodingRate() uint8      { return c.codingRate }
func (c *channel) PreambleLength() uint16 { return c.preambleLength }
func (c *channel) TxPowerDBm() int8       { return c.txPowerDBm }

// Set functions
func (c *channel) SetFrequency(v uint32)      { c.frequency = v }
func (c *channel) SetBandwidth(v uint8)       { c.bandwidth = v }
func (c *channel) SetSpreadingFactor(v uint8) { c.spreadingFactor = v }
func (c *channel) SetCodingRate(v uint8)      { c.codingRate = v }
func (c *channel) SetPreambleLength(v uint16) { c.preambleLength = v }
func (c *channel) SetTxPowerDBm(v int8)       { c.txPowerDBm = v }
