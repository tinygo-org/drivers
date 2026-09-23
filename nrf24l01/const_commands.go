package nrf24l01

const (
	R_REGISTER uint8 = 0x00
	W_REGISTER uint8 = 0x20

	R_RX_PL_WID  uint8 = 0x60
	R_RX_PAYLOAD uint8 = 0x61
	W_TX_PAYLOAD uint8 = 0xA0

	W_TX_PAYLOAD_NOACK uint8 = 0xB0

	FLUSH_TX uint8 = 0xE1
	FLUSH_RX uint8 = 0xE2

	NOP uint8 = 0xFF
)
