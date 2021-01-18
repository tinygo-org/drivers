package sx127x

const (
	// registers
	REG_FIFO                 = 0x00
	REG_OP_MODE              = 0x01
	REG_FRF_MSB              = 0x06
	REG_FRF_MID              = 0x07
	REG_FRF_LSB              = 0x08
	REG_PA_CONFIG            = 0x09
	REG_PA_RAMP              = 0x0a
	REG_OCP                  = 0x0b
	REG_LNA                  = 0x0c
	REG_FIFO_ADDR_PTR        = 0x0d
	REG_FIFO_TX_BASE_ADDR    = 0x0e
	REG_FIFO_RX_BASE_ADDR    = 0x0f
	REG_FIFO_RX_CURRENT_ADDR = 0x10
	REG_IRQ_FLAGS_MASK       = 0x11
	REG_IRQ_FLAGS            = 0x12
	REG_RX_NB_BYTES          = 0x13
	REG_PKT_SNR_VALUE        = 0x19
	REG_PKT_RSSI_VALUE       = 0x1a
	REG_RSSI_VALUE           = 0x1b
	REG_MODEM_CONFIG_1       = 0x1d
	REG_MODEM_CONFIG_2       = 0x1e
	REG_SYMB_TIMEOUT_LSB     = 0x1f
	REG_PREAMBLE_MSB         = 0x20
	REG_PREAMBLE_LSB         = 0x21
	REG_PAYLOAD_LENGTH       = 0x22
	REG_MAX_PAYLOAD_LENGTH   = 0x23
	REG_HOP_PERIOD           = 0x24
	REG_MODEM_CONFIG_3       = 0x26
	REG_FREQ_ERROR_MSB       = 0x28
	REG_FREQ_ERROR_MID       = 0x29
	REG_FREQ_ERROR_LSB       = 0x2a
	REG_RSSI_WIDEBAND        = 0x2c
	REG_DETECTION_OPTIMIZE   = 0x31
	REG_INVERTIQ             = 0x33
	REG_DETECTION_THRESHOLD  = 0x37
	REG_SYNC_WORD            = 0x39
	REG_INVERTIQ2            = 0x3b
	REG_DIO_MAPPING_1        = 0x40
	REG_DIO_MAPPING_2        = 0x41
	REG_VERSION              = 0x42
	REG_PA_DAC               = 0x4d

	// PA config
	PA_BOOST = 0x80

	// Bits masking the corresponding IRQs from the radio
	IRQ_LORA_RXTOUT_MASK = uint8(0x80)
	IRQ_LORA_RXDONE_MASK = uint8(0x40)
	IRQ_LORA_CRCERR_MASK = uint8(0x20)
	IRQ_LORA_HEADER_MASK = uint8(0x10)
	IRQ_LORA_TXDONE_MASK = uint8(0x08)
	IRQ_LORA_CDDONE_MASK = uint8(0x04)
	IRQ_LORA_FHSSCH_MASK = uint8(0x02)
	IRQ_LORA_CDDETD_MASK = uint8(0x01)

	// DIO function mappings                D0D1D2D3
	MAP_DIO0_LORA_RXDONE = uint8(0x00) // 00------
	MAP_DIO0_LORA_TXDONE = uint8(0x40) // 01------
	MAP_DIO1_LORA_RXTOUT = uint8(0x00) // --00----
	MAP_DIO1_LORA_NOP    = uint8(0x30) // --11----
	MAP_DIO2_LORA_NOP    = uint8(0xC0) // ----11--

	PAYLOAD_LENGTH = uint8(0x40)

	// Low Noise Amp
	LNA_MAX_GAIN = uint8(0x23)
	LNA_OFF_GAIN = uint8(0x00)
	LNA_LOW_GAIN = uint8(0x20)

	// Header type
	SX127X_LORA_HEADER_EXPLICIT = uint8(0x00)
	SX127X_LORA_HEADER_IMPLICIT = uint8(0x01)
	// CRC
	SX127X_LORA_CRC_OFF = uint8(0x00)
	SX127X_LORA_CRC_ON  = uint8(0x01)
	// IQ Polarity
	SX127X_LORA_IQ_STANDARD = uint8(0x00)
	SX127X_LORA_IQ_INVERTED = uint8(0x01)
	// Coding Rate
	SX127X_LORA_CR_4_5 = uint8(0x01)
	SX127X_LORA_CR_4_6 = uint8(0x02)
	SX127X_LORA_CR_4_7 = uint8(0x03)
	SX127X_LORA_CR_4_8 = uint8(0x04)
	// Bandwidth
	SX127X_LORA_BW_7_8   = uint8(0x00)
	SX127X_LORA_BW_10_4  = uint8(0x01)
	SX127X_LORA_BW_15_6  = uint8(0x02)
	SX127X_LORA_BW_20_8  = uint8(0x03)
	SX127X_LORA_BW_31_25 = uint8(0x04)
	SX127X_LORA_BW_41_7  = uint8(0x05)
	SX127X_LORA_BW_62_5  = uint8(0x06)
	SX127X_LORA_BW_125_0 = uint8(0x07)
	SX127X_LORA_BW_250_0 = uint8(0x08)
	SX127X_LORA_BW_500_0 = uint8(0x09)
	// Spreading factor
	SX127X_LORA_SF6  = uint8(0x06)
	SX127X_LORA_SF7  = uint8(0x07)
	SX127X_LORA_SF8  = uint8(0x08)
	SX127X_LORA_SF9  = uint8(0x09)
	SX127X_LORA_SF10 = uint8(0x0A)
	SX127X_LORA_SF11 = uint8(0x0B)
	SX127X_LORA_SF12 = uint8(0x0C)
	// Optim Low DR
	SX127X_LOW_DATARATE_OPTIM_OFF = uint8(0x00)
	SX127X_LOW_DATARATE_OPTIM_ON  = uint8(0x01)
	// Automatic gain control
	SX127X_AGC_AUTO_OFF = uint8(0x00)
	SX127X_AGC_AUTO_ON  = uint8(0x01)
	// Operation modes
	SX127X_OPMODE_LORA      = uint8(0x80)
	SX127X_OPMODE_MASK      = uint8(0x07)
	SX127X_OPMODE_SLEEP     = uint8(0x00)
	SX127X_OPMODE_STANDBY   = uint8(0x01)
	SX127X_OPMODE_FSTX      = uint8(0x02)
	SX127X_OPMODE_TX        = uint8(0x03)
	SX127X_OPMODE_FSRX      = uint8(0x04)
	SX127X_OPMODE_RX        = uint8(0x05)
	SX127X_OPMODE_RX_SINGLE = uint8(0x06)
	SX127X_OPMODE_CAD       = uint8(0x07)
)
