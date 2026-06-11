package nrf24l01

const (
	CONFIG     uint8 = 0x00
	EN_AA      uint8 = 0x01
	EN_RXADDR  uint8 = 0x02
	SETUP_AW   uint8 = 0x03
	SETUP_RETR uint8 = 0x04
	RF_CH      uint8 = 0x05
	RF_SETUP   uint8 = 0x06
	STATUS     uint8 = 0x07
	OBSERVE_TX uint8 = 0x08

	RX_ADDR_P0 uint8 = 0x0A
	RX_ADDR_P1 uint8 = 0x0B
	RX_ADDR_P2 uint8 = 0x0C
	RX_ADDR_P3 uint8 = 0x0D
	RX_ADDR_P4 uint8 = 0x0E
	RX_ADDR_P5 uint8 = 0x0F

	TX_ADDR uint8 = 0x10

	RX_PW_P0 uint8 = 0x11
	RX_PW_P1 uint8 = 0x12
	RX_PW_P2 uint8 = 0x13
	RX_PW_P3 uint8 = 0x14
	RX_PW_P4 uint8 = 0x15
	RX_PW_P5 uint8 = 0x16

	FIFO_STATUS uint8 = 0x17

	DYNPD   uint8 = 0x1C
	FEATURE uint8 = 0x1D
)
