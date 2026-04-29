package sx128x

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
)
