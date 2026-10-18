package ina3221

// Config holds the settings for the configuration register.
type Config struct {
	// Channels selects the enabled channels as a mask of
	// EnableChannel1, EnableChannel2, and EnableChannel3.
	Channels ChannelMask

	// Average sets how many samples are averaged.
	Average Average

	// BusConvTime sets the bus voltage conversion time.
	BusConvTime ConvTime

	// ShuntConvTime sets the shunt voltage conversion time.
	ShuntConvTime ConvTime

	// Mode sets the operating mode.
	Mode Mode
}

// DefaultConfig matches the power-on reset value 0x7127.
var DefaultConfig = Config{
	Channels:      EnableAll,
	Average:       Average1,
	BusConvTime:   ConvTime1100us,
	ShuntConvTime: ConvTime1100us,
	Mode:          ModeContShuntBus,
}

// RegisterValue returns the configuration register value.
func (c Config) RegisterValue() uint16 {
	return uint16(c.Channels&0x7)<<12 |
		uint16(c.Average&0x7)<<9 |
		uint16(c.BusConvTime&0x7)<<6 |
		uint16(c.ShuntConvTime&0x7)<<3 |
		uint16(c.Mode&0x7)
}

// NewConfig creates a configuration from a register value.
func NewConfig(val uint16) Config {
	return Config{
		Channels:      ChannelMask(val >> 12 & 0x7),
		Average:       Average(val >> 9 & 0x7),
		BusConvTime:   ConvTime(val >> 6 & 0x7),
		ShuntConvTime: ConvTime(val >> 3 & 0x7),
		Mode:          Mode(val & 0x7),
	}
}

// ChannelMask selects which channels are enabled.
type ChannelMask uint8

const (
	EnableChannel3 ChannelMask = 1 << 0
	EnableChannel2 ChannelMask = 1 << 1
	EnableChannel1 ChannelMask = 1 << 2
	EnableAll                  = EnableChannel1 | EnableChannel2 | EnableChannel3
)

// Average is the number of samples that are averaged.
type Average uint8

const (
	Average1    Average = 0
	Average4    Average = 1
	Average16   Average = 2
	Average64   Average = 3
	Average128  Average = 4
	Average256  Average = 5
	Average512  Average = 6
	Average1024 Average = 7
)

// ConvTime is the conversion time for a voltage measurement.
type ConvTime uint8

const (
	ConvTime140us  ConvTime = 0
	ConvTime204us  ConvTime = 1
	ConvTime332us  ConvTime = 2
	ConvTime588us  ConvTime = 3
	ConvTime1100us ConvTime = 4
	ConvTime2116us ConvTime = 5
	ConvTime4156us ConvTime = 6
	ConvTime8244us ConvTime = 7
)

// Mode is the operating mode.
type Mode uint8

const (
	ModePowerDown      Mode = 0
	ModeSingleShunt    Mode = 1
	ModeSingleBus      Mode = 2
	ModeSingleShuntBus Mode = 3
	ModeContShunt      Mode = 5
	ModeContBus        Mode = 6
	ModeContShuntBus   Mode = 7
)

// SingleShot reports whether the mode needs Trigger to start a conversion.
func (m Mode) SingleShot() bool {
	return m >= ModeSingleShunt && m <= ModeSingleShuntBus
}
