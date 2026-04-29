package sx128x

type SleepConfig = uint8
type StandbyConfig = uint8
type PeriodBase = uint8
type PacketType = uint8
type RadioRampTime = uint8
type CadSymbolNum = uint8
type SpreadingFactor = uint8
type Bandwidth = uint8
type CodingRate = uint8
type RegulatorMode = uint8

type CircuitMode = uint8
type CommandStatus = uint8

const (
	// SX128X SPI commands
	CMD_NOP = uint8(0x00)

	CMD_SET_UART_SPEED = uint8(0x9D)
	CMD_GET_STATUS     = uint8(0xC0)

	// Register Access Operations
	CMD_WRITE_REGISTER = uint8(0x18)
	CMD_READ_REGISTER  = uint8(0x19)

	// Data Buffer Operations
	CMD_WRITE_BUFFER = uint8(0x1A)
	CMD_READ_BUFFER  = uint8(0x1B)

	// Radio Operation Modes
	CMD_SET_SLEEP                  = uint8(0x84)
	CMD_SET_STANDBY                = uint8(0x80)
	CMD_SET_FS                     = uint8(0xC1)
	CMD_SET_TX                     = uint8(0x83)
	CMD_SET_RX                     = uint8(0x82)
	CMD_SET_RX_DUTY_CYCLE          = uint8(0x94)
	CMD_SET_LONG_PREAMBLE          = uint8(0x9B)
	CMD_SET_CAD                    = uint8(0xC5)
	CMD_SET_TX_CONTINUOUS_WAVE     = uint8(0xD1)
	CMD_SET_TX_CONTINUOUS_PREAMBLE = uint8(0xD2)
	CMD_SET_AUTO_TX                = uint8(0x98)
	CMD_SET_AUTO_FS                = uint8(0x9E)

	// Radio Configuration
	CMD_SET_PACKET_TYPE         = uint8(0x8A)
	CMD_GET_PACKET_TYPE         = uint8(0x03)
	CMD_SET_RF_FREQUENCY        = uint8(0x86)
	CMD_SET_TX_PARAMS           = uint8(0x8E)
	CMD_SET_CAD_PARAMS          = uint8(0x88)
	CMD_SET_BUFFER_BASE_ADDRESS = uint8(0x8F)
	CMD_SET_MODULATION_PARAMS   = uint8(0x8B)
	CMD_SET_PACKET_PARAMS       = uint8(0x8C)

	// Communication Status Information
	CMD_GET_RX_BUFFER_STATUS = uint8(0x17)
	CMD_GET_PACKET_STATUS    = uint8(0x1D)
	CMD_GET_RSSI_INST        = uint8(0x1F)

	// IRQ Handling
	CMD_SET_DIO_IRQ_PARAMS = uint8(0x8D)
	CMD_GET_IRQ_STATUS     = uint8(0x15)
	CMD_CLEAR_IRQ_STATUS   = uint8(0x97)

	// Miscellaneous
	CMD_SET_REGULATOR_MODE = uint8(0x96)
	CMD_SET_SAVE_CONTEXT   = uint8(0xD5)

	// GetStatus
	CIRCUIT_MODE_MASK       = uint8(0b11100000)
	CIRCUIT_MODE_STDBY_RC   = CircuitMode(0x2)
	CIRCUIT_MODE_STDBY_XOSC = CircuitMode(0x3)
	CIRCUIT_MODE_FS         = CircuitMode(0x4)
	CIRCUIT_MODE_RX         = CircuitMode(0x5)
	CIRCUIT_MODE_TX         = CircuitMode(0x6)

	COMMAND_STATUS_MASK             = uint8(0b00011100)
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

	RX_CONTINUOUS_MODE = uint16(0xFFFF)

	// LongPreamble
	LONG_PREAMBLE_ENABLE  = uint8(1)
	LONG_PREAMBLE_DISABLE = uint8(0)

	// AutoFs
	AUTO_FS_ENABLE  = uint8(1)
	AUTO_FS_DISABLE = uint8(0)

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

	// SpreadingFactor
	LORA_SF_5  = SpreadingFactor(0x50)
	LORA_SF_6  = SpreadingFactor(0x60)
	LORA_SF_7  = SpreadingFactor(0x70)
	LORA_SF_8  = SpreadingFactor(0x80)
	LORA_SF_9  = SpreadingFactor(0x90)
	LORA_SF_10 = SpreadingFactor(0xA0)
	LORA_SF_11 = SpreadingFactor(0xB0)
	LORA_SF_12 = SpreadingFactor(0xC0)

	// Bandwidth
	LORA_BW_1600 = Bandwidth(0x0A)
	LORA_BW_800  = Bandwidth(0x18)
	LORA_BW_400  = Bandwidth(0x26)
	LORA_BW_200  = Bandwidth(0x34)

	// CodingRate
	LORA_CR_4_5    = CodingRate(0x01)
	LORA_CR_4_6    = CodingRate(0x02)
	LORA_CR_4_7    = CodingRate(0x03)
	LORA_CR_4_8    = CodingRate(0x04)
	LORA_CR_LI_4_5 = CodingRate(0x05)
	LORA_CR_LI_4_6 = CodingRate(0x06)
	LORA_CR_LI_4_8 = CodingRate(0x07)

	// LoraPacketParams
	LORA_EXPLICIT_HEADER = uint8(0x00)
	LORA_IMPLICIT_HEADER = uint8(0x80)

	LORA_CRC_ENABLE  = uint8(0x20)
	LORA_CRC_DISABLE = uint8(0x00)

	LORA_IQ_INVERTED = uint8(0x00)
	LORA_IQ_STD      = uint8(0x40)

	// RegulatorMode
	REGULATOR_LDO   = RegulatorMode(0)
	REGULATOR_DC_DC = RegulatorMode(1)

	// IRQ masks
	IRQ_ALL_MASK                            = uint16(0xFFFF)
	IRQ_NONE_MASK                           = uint16(0x0000)
	IRQ_TX_DONE_MASK                        = uint16(0b0000000000000001)
	IRQ_RX_DONE_MASK                        = uint16(0b0000000000000010)
	IRQ_SYNC_WORD_VALID_MASK                = uint16(0b0000000000000100)
	IRQ_SYNC_WORD_ERROR_MASK                = uint16(0b0000000000001000)
	IRQ_HEADER_VALID_MASK                   = uint16(0b0000000000010000)
	IRQ_HEADER_ERROR_MASK                   = uint16(0b0000000000100000)
	IRQ_CRC_ERROR_MASK                      = uint16(0b0000000001000000)
	IRQ_RANGING_SLAVE_RESPONSE_DONE_MASK    = uint16(0b0000000010000000)
	IRQ_RANGING_SLAVE_RESPONSE_DISCARD_MASK = uint16(0b0000000100000000)
	IRQ_RANGING_MASTER_RESULT_VALID_MASK    = uint16(0b0000001000000000)
	IRQ_RANGING_MASTER_TIMEOUT_MASK         = uint16(0b0000010000000000)
	IRQ_RANGING_SLAVE_REQUEST_VALID_MASK    = uint16(0b0000100000000000)
	IRQ_CAD_DONE_MASK                       = uint16(0b0001000000000000)
	IRQ_CAD_DETECTED_MASK                   = uint16(0b0010000000000000)
	IRQ_RX_TX_TIMEOUT_MASK                  = uint16(0b0100000000000000)
	IRQ_PREAMBLE_DETECTED_MASK              = uint16(0b1000000000000000)
	IRQ_ADVANCED_RANGING_DONE_MASK          = uint16(0b1000000000000000)

	// SX128X register map
	REG_FIRMWARE_VERSIONS              = uint16(0x153)
	REG_RX_GAIN                        = uint16(0x891)
	REG_MANUAL_GAIN_SETTING            = uint16(0x895)
	REG_LNA_GAIN_VALUE                 = uint16(0x89E)
	REG_LNA_GAIN_CONTROL               = uint16(0x89F)
	REG_SYNCH_PEAK_ATTENUATION         = uint16(0x8C2)
	REG_PAYLOAD_LENGTH                 = uint16(0x901)
	REG_LORA_HEADER_MODE               = uint16(0x903)
	REG_RANGING_REQUEST_ADDRESS_BYTE_3 = uint16(0x912)
	REG_RANGING_REQUEST_ADDRESS_BYTE_2 = uint16(0x913)
	REG_RANGING_REQUEST_ADDRESS_BYTE_1 = uint16(0x914)
	REG_RANGING_REQUEST_ADDRESS_BYTE_0 = uint16(0x915)
	REG_RANGING_DEVICE_ADDRESS_BYTE_3  = uint16(0x916)
	REG_RANGING_DEVICE_ADDRESS_BYTE_2  = uint16(0x917)
	REG_RANGING_DEVICE_ADDRESS_BYTE_1  = uint16(0x918)
	REG_RANGING_DEVICE_ADDRESS_BYTE_0  = uint16(0x919)
	REG_RANGING_FILTER_WINDOW_SIZE     = uint16(0x91E)
	REG_RESET_RANGING_FILTER           = uint16(0x923)
	REG_RANGING_RESULT_MUX             = uint16(0x924)
	REG_SF_ADDITIONAL_CONFIGURATION    = uint16(0x925)
	REG_RANGING_CALIBRATION_BYTE_2     = uint16(0x92B)
	REG_RANGING_CALIBRATION_BYTE_1     = uint16(0x92C)
	REG_RANGING_CALIBRATION_BYTE_0     = uint16(0x92D)
	REG_RANGING_ID_CHECK_LENGTH        = uint16(0x931)
	REG_FREQUENCY_ERROR_CORRECTION     = uint16(0x93C)
	REG_CAD_DETECT_PEAK                = uint16(0x942)
	REG_LORA_SYNC_WORD_MSB             = uint16(0x944)
	REG_LORA_SYNC_WORD_LSB             = uint16(0x945)
	REG_HEADER_CRC                     = uint16(0x954)
	REG_CODING_RATE                    = uint16(0x950)
	REG_FEI_BYTE_2                     = uint16(0x954)
	REG_FEI_BYTE_1                     = uint16(0x955)
	REG_FEI_BYTE_0                     = uint16(0x956)
	REG_RANGING_RESULT_BYTE_2          = uint16(0x961)
	REG_RANGING_RESULT_BYTE_1          = uint16(0x962)
	REG_RANGING_RESULT_BYTE_0          = uint16(0x963)
	REG_RANGING_RSSI                   = uint16(0x964)
	REG_FREEZE_RANGING_RESULT          = uint16(0x97F)
	REG_PACKET_PREAMBLE_SETTINGS       = uint16(0x9C1)
	REG_WHITENING_INITIAL_VALUE        = uint16(0x9C5)
	REG_CRC_POLYNOMIAL_DEFINITION_MSB  = uint16(0x9C6)
	REG_CRC_POLYNOMIAL_DEFINITION_LSB  = uint16(0x9C7)
	REG_CRC_POLYNOMIAL_SEED_BYTE_2     = uint16(0x9C7)
	REG_CRC_POLYNOMIAL_SEED_BYTE_1     = uint16(0x9C8)
	REG_CRC_POLYNOMIAL_SEED_BYTE_0     = uint16(0x9C9)
	REG_CRC_MSB_INITIAL_VALUE          = uint16(0x9C8)
	REG_CRC_LSB_INITIAL_VALUE          = uint16(0x9C9)
	REG_SYNC_ADDRESS_CONTROL           = uint16(0x9CD)
	REG_SYNC_ADDRESS_1_BYTE_4          = uint16(0x9CE)
	REG_SYNC_ADDRESS_1_BYTE_3          = uint16(0x9CF)
	REG_SYNC_ADDRESS_1_BYTE_2          = uint16(0x9D0)
	REG_SYNC_ADDRESS_1_BYTE_1          = uint16(0x9D1)
	REG_SYNC_ADDRESS_1_BYTE_0          = uint16(0x9D2)
	REG_SYNC_ADDRESS_2_BYTE_4          = uint16(0x9D3)
	REG_SYNC_ADDRESS_2_BYTE_3          = uint16(0x9D4)
	REG_SYNC_ADDRESS_2_BYTE_2          = uint16(0x9D5)
	REG_SYNC_ADDRESS_2_BYTE_1          = uint16(0x9D6)
	REG_SYNC_ADDRESS_2_BYTE_0          = uint16(0x9D7)
	REG_SYNC_ADDRESS_3_BYTE_4          = uint16(0x9D8)
	REG_SYNC_ADDRESS_3_BYTE_3          = uint16(0x9D9)
	REG_SYNC_ADDRESS_3_BYTE_2          = uint16(0x9DA)
	REG_SYNC_ADDRESS_3_BYTE_1          = uint16(0x9DB)
	REG_SYNC_ADDRESS_3_BYTE_0          = uint16(0x9DC)
)
