package sx128x

type SleepConfig = uint8
type StandbyConfig = uint8
type PeriodBase = uint8
type PacketType = uint8
type RadioRampTime = uint8
type CadSymbolNum = uint8

// GFSK Modulation Params
type GFSKBLEBitrateBandwidth = uint8
type ModulationIndex = uint8
type ModulationShaping = uint8

// GFSK Packet Params
type GFSKPreambleLength = uint8
type GFSKSyncWordLength = uint8
type GFSKSyncWordMatch = uint8
type GFSKHeaderType = uint8
type GFSKCrcType = uint8

// BLE Packet Params
type BLEConnectionState = uint8
type BLECrcType = uint8
type BLETestPayload = uint8

// FLRC Modulation Params
type FLRCBitrateBandwidth = uint8
type FLRCCodingRate = uint8

// FLRC Packet Params
type FLRCPreambleLength = uint8
type FLRCSyncWordLength = uint8
type FLRCSyncWordMatch = uint8
type FLRCHeaderType = uint8
type FLRCCrcType = uint8

// LoRa Modulation Params
type LoRaSpreadingFactor = uint8
type LoRaBandwidth = uint8
type LoRaCodingRate = uint8

// LoRa Packet Params
type LoRaHeaderType = uint8
type LoRaCrcType = uint8
type LoRaIqType = uint8

// Misc
type RegulatorMode = uint8
type IRQMask = uint16
type CircuitMode = uint8
type CommandStatus = uint8

// Packet Status
type GFSKPacketInfo = uint8
type BLEPacketInfo = uint8
type FLRCPacketInfo = uint8

const (
	whiteningDisable = 0x00
	whiteningEnable  = 0x08

	// Circuit Mode
	circuitModeMask         = uint8(0b11100000)
	CIRCUIT_MODE_STDBY_RC   = CircuitMode(0x2)
	CIRCUIT_MODE_STDBY_XOSC = CircuitMode(0x3)
	CIRCUIT_MODE_FS         = CircuitMode(0x4)
	CIRCUIT_MODE_RX         = CircuitMode(0x5)
	CIRCUIT_MODE_TX         = CircuitMode(0x6)

	// Command Status
	commandStatusMask               = uint8(0b00011100)
	COMMAND_STATUS_SUCCESS          = CommandStatus(0x1)
	COMMAND_STATUS_DATA_AVAILABLE   = CommandStatus(0x2)
	COMMAND_STATUS_TIMEOUT          = CommandStatus(0x3)
	COMMAND_STATUS_PROCESSING_ERROR = CommandStatus(0x4)
	COMMAND_STATUS_EXECUTION_ERROR  = CommandStatus(0x5)
	COMMAND_STATUS_TX_DONE          = CommandStatus(0x6)

	// SleepConfig
	SLEEP_DATA_BUFFER_RETAIN = SleepConfig(2)
	SLEEP_DATA_RAM_RETAIN    = SleepConfig(1)

	// StandbyConfig
	STANDBY_RC   = StandbyConfig(0)
	STANDBY_XOSC = StandbyConfig(1)

	// PeriodBase
	PERIOD_BASE_15_625_US = PeriodBase(0)
	PERIOD_BASE_62_5_US   = PeriodBase(1)
	PERIOD_BASE_1_MS      = PeriodBase(2)
	PERIOD_BASE_4_MS      = PeriodBase(3)

	// PacketType
	PACKET_TYPE_GFSK    = PacketType(0x00) // default
	PACKET_TYPE_LORA    = PacketType(0x01)
	PACKET_TYPE_RANGING = PacketType(0x02)
	PACKET_TYPE_FLRC    = PacketType(0x03)
	PACKET_TYPE_BLE     = PacketType(0x04)

	// RampTime
	RADIO_RAMP_02_US = RadioRampTime(0x00)
	RADIO_RAMP_04_US = RadioRampTime(0x20)
	RADIO_RAMP_06_US = RadioRampTime(0x40)
	RADIO_RAMP_08_US = RadioRampTime(0x60)
	RADIO_RAMP_10_US = RadioRampTime(0x80)
	RADIO_RAMP_12_US = RadioRampTime(0xA0)
	RADIO_RAMP_16_US = RadioRampTime(0xC0)
	RADIO_RAMP_20_US = RadioRampTime(0xE0)

	// CadSymbolNum
	LORA_CAD_01_SYMBOL  = CadSymbolNum(0x00)
	LORA_CAD_02_SYMBOLS = CadSymbolNum(0x20)
	LORA_CAD_04_SYMBOLS = CadSymbolNum(0x40)
	LORA_CAD_08_SYMBOLS = CadSymbolNum(0x60)
	LORA_CAD_16_SYMBOLS = CadSymbolNum(0x80)

	// GFSK Modulation Params
	// Bitrate + Bandwidth - same for BLE
	GFSK_BLE_BR_2_000_BW_2_4 = GFSKBLEBitrateBandwidth(0x04)
	GFSK_BLE_BR_1_600_BW_2_4 = GFSKBLEBitrateBandwidth(0x28)
	GFSK_BLE_BR_1_000_BW_2_4 = GFSKBLEBitrateBandwidth(0x4C)
	GFSK_BLE_BR_1_000_BW_1_2 = GFSKBLEBitrateBandwidth(0x45)
	GFSK_BLE_BR_0_800_BW_2_4 = GFSKBLEBitrateBandwidth(0x70)
	GFSK_BLE_BR_0_800_BW_1_2 = GFSKBLEBitrateBandwidth(0x69)
	GFSK_BLE_BR_0_500_BW_1_2 = GFSKBLEBitrateBandwidth(0x8D)
	GFSK_BLE_BR_0_500_BW_0_6 = GFSKBLEBitrateBandwidth(0x86)
	GFSK_BLE_BR_0_400_BW_1_2 = GFSKBLEBitrateBandwidth(0xB1)
	GFSK_BLE_BR_0_400_BW_0_6 = GFSKBLEBitrateBandwidth(0xAA)
	GFSK_BLE_BR_0_250_BW_0_6 = GFSKBLEBitrateBandwidth(0xCE)
	GFSK_BLE_BR_0_250_BW_0_3 = GFSKBLEBitrateBandwidth(0xC7)
	GFSK_BLE_BR_0_125_BW_0_3 = GFSKBLEBitrateBandwidth(0xEF)

	// Modulation Index - same for BLE
	MOD_IND_0_35 = ModulationIndex(0x00)
	MOD_IND_0_5  = ModulationIndex(0x01)
	MOD_IND_0_75 = ModulationIndex(0x02)
	MOD_IND_1_00 = ModulationIndex(0x03)
	MOD_IND_1_25 = ModulationIndex(0x04)
	MOD_IND_1_50 = ModulationIndex(0x05)
	MOD_IND_1_75 = ModulationIndex(0x06)
	MOD_IND_2_00 = ModulationIndex(0x07)
	MOD_IND_2_25 = ModulationIndex(0x08)
	MOD_IND_2_50 = ModulationIndex(0x09)
	MOD_IND_2_75 = ModulationIndex(0x0A)
	MOD_IND_3_00 = ModulationIndex(0x0B)
	MOD_IND_3_25 = ModulationIndex(0x0C)
	MOD_IND_3_50 = ModulationIndex(0x0D)
	MOD_IND_3_75 = ModulationIndex(0x0E)
	MOD_IND_4_00 = ModulationIndex(0x0F)

	// Modulation Shaping - same for BLE and FLRC
	MOD_SHAPING_OFF = ModulationShaping(0x00)
	MOD_SHAPING_1_0 = ModulationShaping(0x10)
	MOD_SHAPING_0_5 = ModulationShaping(0x20)

	// GFSK Packet Params
	// Preamble Length
	GFSK_PREAMBLE_LENGTH_04_BITS = GFSKPreambleLength(0x00)
	GFSK_PREAMBLE_LENGTH_08_BITS = GFSKPreambleLength(0x10)
	GFSK_PREAMBLE_LENGTH_12_BITS = GFSKPreambleLength(0x20)
	GFSK_PREAMBLE_LENGTH_16_BITS = GFSKPreambleLength(0x30)
	GFSK_PREAMBLE_LENGTH_20_BITS = GFSKPreambleLength(0x40)
	GFSK_PREAMBLE_LENGTH_24_BITS = GFSKPreambleLength(0x50)
	GFSK_PREAMBLE_LENGTH_28_BITS = GFSKPreambleLength(0x60)
	GFSK_PREAMBLE_LENGTH_32_BITS = GFSKPreambleLength(0x70)

	// Sync Word Length
	GFSK_SYNC_WORD_LEN_1_B = GFSKSyncWordLength(0x00)
	GFSK_SYNC_WORD_LEN_2_B = GFSKSyncWordLength(0x02)
	GFSK_SYNC_WORD_LEN_3_B = GFSKSyncWordLength(0x04)
	GFSK_SYNC_WORD_LEN_4_B = GFSKSyncWordLength(0x06)
	GFSK_SYNC_WORD_LEN_5_B = GFSKSyncWordLength(0x08)

	// Sync Word Match
	GFSK_SYNCWORD_MATCH_OFF   = GFSKSyncWordMatch(0x00)
	GFSK_SYNCWORD_MATCH_1     = GFSKSyncWordMatch(0x10)
	GFSK_SYNCWORD_MATCH_2     = GFSKSyncWordMatch(0x20)
	GFSK_SYNCWORD_MATCH_1_2   = GFSKSyncWordMatch(0x30)
	GFSK_SYNCWORD_MATCH_3     = GFSKSyncWordMatch(0x40)
	GFSK_SYNCWORD_MATCH_1_3   = GFSKSyncWordMatch(0x50)
	GFSK_SYNCWORD_MATCH_2_3   = GFSKSyncWordMatch(0x60)
	GFSK_SYNCWORD_MATCH_1_2_3 = GFSKSyncWordMatch(0x70)

	// GFSK Header Type
	GFSK_HEADER_FIXED_LENGTH    = GFSKHeaderType(0x00)
	GFSK_HEADER_VARIABLE_LENGTH = GFSKHeaderType(0x20)

	// GFSK CRC Type
	GFSK_CRC_OFF     = GFSKCrcType(0x00)
	GFSK_CRC_1_BYTE  = GFSKCrcType(0x10)
	GFSK_CRC_2_BYTES = GFSKCrcType(0x20)

	// BLE Packet Params
	// Connection State
	BLE_MASTER_SLAVE   = BLEConnectionState(0x00)
	BLE_ADVERTISER     = BLEConnectionState(0x02)
	BLE_TX_TEST_MODE   = BLEConnectionState(0x04)
	BLE_RX_TEST_MODE   = BLEConnectionState(0x06)
	BLE_RXTX_TEST_MODE = BLEConnectionState(0x08)

	// CRC Type
	BLE_CRC_OFF     = BLECrcType(0x00)
	BLE_CRC_3_BYTES = BLECrcType(0x10)

	// BLE Test Payload
	BLE_PAYLOAD_PRBS_9       = BLETestPayload(0x00)
	BLE_PAYLOAD_EYELONG_1_0  = BLETestPayload(0x04)
	BLE_PAYLOAD_EYESHORT_1_0 = BLETestPayload(0x08)
	BLE_PAYLOAD_PRBS_15      = BLETestPayload(0x0C)
	BLE_PAYLOAD_ALL_1        = BLETestPayload(0x10)
	BLE_PAYLOAD_ALL_0        = BLETestPayload(0x14)
	BLE_PAYLOAD_EYELONG_0_1  = BLETestPayload(0x18)
	BLE_PAYLOAD_EYESHORT_0_1 = BLETestPayload(0x1C)

	// FLRC Modulation Params
	// Bitrate + Bandwidth
	FLRC_BR_1_300_BW_1_2 = FLRCBitrateBandwidth(0x45)
	FLRC_BR_1_000_BW_1_2 = FLRCBitrateBandwidth(0x69)
	FLRC_BR_0_650_BW_0_6 = FLRCBitrateBandwidth(0x86)
	FLRC_BR_0_520_BW_0_6 = FLRCBitrateBandwidth(0xAA)
	FLRC_BR_0_325_BW_0_3 = FLRCBitrateBandwidth(0xC7)
	FLRC_BR_0_260_BW_0_3 = FLRCBitrateBandwidth(0xEB)

	// Coding Rate
	FLRC_CR_1_2 = FLRCCodingRate(0x00) // 1/2
	FLRC_CR_3_4 = FLRCCodingRate(0x02) // 3/4
	FLRC_CR_1_0 = FLRCCodingRate(0x04) // 1

	// FLRC Packet Params
	// Preamble Length
	FLRC_PREAMBLE_LENGTH_4_BITS  = FLRCPreambleLength(0x00)
	FLRC_PREAMBLE_LENGTH_8_BITS  = FLRCPreambleLength(0x10)
	FLRC_PREAMBLE_LENGTH_12_BITS = FLRCPreambleLength(0x20)
	FLRC_PREAMBLE_LENGTH_16_BITS = FLRCPreambleLength(0x30)
	FLRC_PREAMBLE_LENGTH_20_BITS = FLRCPreambleLength(0x40)
	FLRC_PREAMBLE_LENGTH_24_BITS = FLRCPreambleLength(0x50)
	FLRC_PREAMBLE_LENGTH_28_BITS = FLRCPreambleLength(0x60)
	FLRC_PREAMBLE_LENGTH_32_BITS = FLRCPreambleLength(0x70)

	// Sync Word Length
	FLRC_SYNC_WORD_LEN_0       = FLRCSyncWordLength(0x00)
	FLRC_SYNC_WORD_LEN_32_BITS = FLRCSyncWordLength(0x04)

	// Sync Word Match
	FLRC_SYNC_WORD_MATCH_DISABLE = FLRCSyncWordMatch(0x00) // Disable Sync Word
	FLRC_SYNC_WORD_MATCH_1       = FLRCSyncWordMatch(0x10) // Sync Word 1
	FLRC_SYNC_WORD_MATCH_2       = FLRCSyncWordMatch(0x20) // Sync Word 2
	FLRC_SYNC_WORD_MATCH_1_2     = FLRCSyncWordMatch(0x30) // Sync Word 1 or Sync Word 2
	FLRC_SYNC_WORD_MATCH_3       = FLRCSyncWordMatch(0x40) // Sync Word 3
	FLRC_SYNC_WORD_MATCH_1_3     = FLRCSyncWordMatch(0x50) // Sync Word 1 or Sync Word 3
	FLRC_SYNC_WORD_MATCH_2_3     = FLRCSyncWordMatch(0x60) // Sync Word 2 or Sync Word 3
	FLRC_SYNC_WORD_MATCH_1_2_3   = FLRCSyncWordMatch(0x70) // Sync Word 1 or Sync Word 2 or Sync Word 3

	// Header Type
	FLRC_HEADER_FIXED_LENGTH    = FLRCHeaderType(0x00)
	FLRC_HEADER_VARIABLE_LENGTH = FLRCHeaderType(0x20)

	// CRC Type
	FLRC_CRC_OFF     = FLRCCrcType(0x00)
	FLRC_CRC_1_BYTE  = FLRCCrcType(0x10)
	FLRC_CRC_2_BYTES = FLRCCrcType(0x20)
	FLRC_CRC_3_BYTES = FLRCCrcType(0x30)

	// LoRa Modulation Params

	// SpreadingFactor
	LORA_SF_5  = LoRaSpreadingFactor(0x50)
	LORA_SF_6  = LoRaSpreadingFactor(0x60)
	LORA_SF_7  = LoRaSpreadingFactor(0x70)
	LORA_SF_8  = LoRaSpreadingFactor(0x80)
	LORA_SF_9  = LoRaSpreadingFactor(0x90)
	LORA_SF_10 = LoRaSpreadingFactor(0xA0)
	LORA_SF_11 = LoRaSpreadingFactor(0xB0)
	LORA_SF_12 = LoRaSpreadingFactor(0xC0)

	// Bandwidth
	LORA_BW_1600 = LoRaBandwidth(0x0A)
	LORA_BW_800  = LoRaBandwidth(0x18)
	LORA_BW_400  = LoRaBandwidth(0x26)
	LORA_BW_200  = LoRaBandwidth(0x34)

	// CodingRate
	LORA_CR_4_5    = LoRaCodingRate(0x01)
	LORA_CR_4_6    = LoRaCodingRate(0x02)
	LORA_CR_4_7    = LoRaCodingRate(0x03)
	LORA_CR_4_8    = LoRaCodingRate(0x04)
	LORA_CR_LI_4_5 = LoRaCodingRate(0x05)
	LORA_CR_LI_4_6 = LoRaCodingRate(0x06)
	LORA_CR_LI_4_8 = LoRaCodingRate(0x07)

	// LoraPacketParams
	// HeaderType
	LORA_HEADER_EXPLICIT = LoRaHeaderType(0x00)
	LORA_HEADER_IMPLICIT = LoRaHeaderType(0x80)

	// CRC Type
	LORA_CRC_ENABLE  = LoRaCrcType(0x20)
	LORA_CRC_DISABLE = LoRaCrcType(0x00)

	// IQ Type
	LORA_IQ_INVERTED = LoRaIqType(0x00)
	LORA_IQ_STD      = LoRaIqType(0x40)

	// RegulatorMode
	REGULATOR_LDO   = RegulatorMode(0)
	REGULATOR_DC_DC = RegulatorMode(1)

	// IRQ masks
	IRQ_ALL_MASK                            = IRQMask(0xFFFF)
	IRQ_NONE_MASK                           = IRQMask(0x0000)
	IRQ_TX_DONE_MASK                        = IRQMask(0b0000000000000001)
	IRQ_RX_DONE_MASK                        = IRQMask(0b0000000000000010)
	IRQ_SYNC_WORD_VALID_MASK                = IRQMask(0b0000000000000100)
	IRQ_SYNC_WORD_ERROR_MASK                = IRQMask(0b0000000000001000)
	IRQ_HEADER_VALID_MASK                   = IRQMask(0b0000000000010000)
	IRQ_HEADER_ERROR_MASK                   = IRQMask(0b0000000000100000)
	IRQ_CRC_ERROR_MASK                      = IRQMask(0b0000000001000000)
	IRQ_RANGING_SLAVE_RESPONSE_DONE_MASK    = IRQMask(0b0000000010000000)
	IRQ_RANGING_SLAVE_RESPONSE_DISCARD_MASK = IRQMask(0b0000000100000000)
	IRQ_RANGING_MASTER_RESULT_VALID_MASK    = IRQMask(0b0000001000000000)
	IRQ_RANGING_MASTER_TIMEOUT_MASK         = IRQMask(0b0000010000000000)
	IRQ_RANGING_SLAVE_REQUEST_VALID_MASK    = IRQMask(0b0000100000000000)
	IRQ_CAD_DONE_MASK                       = IRQMask(0b0001000000000000)
	IRQ_CAD_DETECTED_MASK                   = IRQMask(0b0010000000000000)
	IRQ_RX_TX_TIMEOUT_MASK                  = IRQMask(0b0100000000000000)
	IRQ_PREAMBLE_DETECTED_MASK              = IRQMask(0b1000000000000000)
	IRQ_ADVANCED_RANGING_DONE_MASK          = IRQMask(0b1000000000000000)

	// GFSK Packet Info
	GFSK_SYNC_ERROR       = GFSKPacketInfo(0b1000000)
	GFSK_LENGTH_ERROR     = GFSKPacketInfo(0b0100000)
	GFSK_CRC_ERROR        = GFSKPacketInfo(0b0010000)
	GFSK_ABORT_ERROR      = GFSKPacketInfo(0b0001000)
	GFSK_HEADER_RECEIVED  = GFSKPacketInfo(0b0000100)
	GFSK_PACKET_RECEIVED  = GFSKPacketInfo(0b0000010)
	GFSK_PACKET_CRTL_BUSY = GFSKPacketInfo(0b0000001)

	// BLE Packet Info
	BLE_SYNC_ERROR       = BLEPacketInfo(0b1000000)
	BLE_LENGTH_ERROR     = BLEPacketInfo(0b0100000)
	BLE_CRC_ERROR        = BLEPacketInfo(0b0010000)
	BLE_ABORT_ERROR      = BLEPacketInfo(0b0001000)
	BLE_HEADER_RECEIVED  = BLEPacketInfo(0b0000100)
	BLE_PACKET_RECEIVED  = BLEPacketInfo(0b0000010)
	BLE_PACKET_CRTL_BUSY = BLEPacketInfo(0b0000001)

	// FLRC Packet Info
	FLRC_SYNC_ERROR       = FLRCPacketInfo(0b1000000)
	FLRC_LENGTH_ERROR     = FLRCPacketInfo(0b0100000)
	FLRC_CRC_ERROR        = FLRCPacketInfo(0b0010000)
	FLRC_ABORT_ERROR      = FLRCPacketInfo(0b0001000)
	FLRC_HEADER_RECEIVED  = FLRCPacketInfo(0b0000100)
	FLRC_PACKET_RECEIVED  = FLRCPacketInfo(0b0000010)
	FLRC_PACKET_CRTL_BUSY = FLRCPacketInfo(0b0000001)
)
