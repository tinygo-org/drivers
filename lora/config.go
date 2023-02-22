package lora

import "errors"

// Config holds the LoRa configuration parameters
type Config struct {
	Freq           uint32 // Frequency
	Cr             uint8  // Coding Rate
	Sf             uint8  // Spread Factor
	Bw             uint8  // Bandwidth
	Ldr            uint8  // Low Data Rate
	Preamble       uint16 // PreambleLength
	SyncWord       uint16 // Sync Word
	HeaderType     uint8  // Header : Implicit/explicit
	Crc            uint8  // CRC : Yes/No
	Iq             uint8  // iq : Standard/inverted
	LoraTxPowerDBm int8   // Tx power in Dbm
}

var (
	ErrUndefinedLoraConf = errors.New("Undefined Lora configuration")
)

const (
	SpreadingFactor5  = 0x05
	SpreadingFactor6  = 0x06
	SpreadingFactor7  = 0x07
	SpreadingFactor8  = 0x08
	SpreadingFactor9  = 0x09
	SpreadingFactor10 = 0x0A
	SpreadingFactor11 = 0x0B
	SpreadingFactor12 = 0x0C
)

const (
	CodingRate4_5 = 0x01 //  7     0     LoRa coding rate: 4/5
	CodingRate4_6 = 0x02 //  7     0                       4/6
	CodingRate4_7 = 0x03 //  7     0                       4/7
	CodingRate4_8 = 0x04 //  7     0                       4/8
)

const (
	HeaderExplicit = 0x00 //  7     0     LoRa header mode: explicit
	HeaderImplicit = 0x01 //  7     0                       implicit
)

const (
	LowDataRateOptimizeOff = 0x00 //  7     0     LoRa low data rate optimization: disabled
	LowDataRateOptimizeOn  = 0x01 //  7     0                                      enabled
)

const (
	CRCOff = 0x00 //  7     0     LoRa CRC mode: disabled
	CRCOn  = 0x01 //  7     0                    enabled
)

const (
	IQStandard = 0x00 //  7     0     LoRa IQ setup: standard
	IQInverted = 0x01 //  7     0                    inverted
)

const (
	Bandwidth_7_8   = iota // 7.8 kHz
	Bandwidth_10_4         // 10.4 kHz
	Bandwidth_15_6         // 15.6 kHz
	Bandwidth_20_8         // 20.8 kHz
	Bandwidth_31_25        // 31.25 kHz
	Bandwidth_41_7         // 41.7 kHz
	Bandwidth_62_5         // 62.5 kHz
	Bandwidth_125_0        // 125.0 kHz
	Bandwidth_250_0        // 250.0 kHz
	Bandwidth_500_0        // 500.0 kHz
)

const (
	SyncPublic = iota
	SyncPrivate
)

const (
	MHz_868_1 = 868100000
	MHz_868_5 = 868500000
	MHz_916_8 = 916800000
	MHz_923_3 = 923300000
)
