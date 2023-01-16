// Package e220 implements a driver for the e220 LoRa module.
//
// https://dragon-torch.tech/wp-content/uploads/2022/08/data_sheet.pdf
// P23 - 27
package e220

// E220RegisterLength represents register length of E220
const (
	E220RegisterLength = 9
)

// UartSerialPortRate represents UART serial port rate of E220
const (
	UartSerialPortRate1200Bps   = 0b000
	UartSerialPortRate2400Bps   = 0b001
	UartSerialPortRate4800Bps   = 0b010
	UartSerialPortRate9600Bps   = 0b011
	UartSerialPortRate19200Bps  = 0b100
	UartSerialPortRate38400Bps  = 0b101
	UartSerialPortRate57600Bps  = 0b110
	UartSerialPortRate115200Bps = 0b111
)

// AirDataRate represents air data rate of E220
const (
	AirDataRate15625Bps = 0b00000
	AirDataRate9375Bps  = 0b00100
	AirDataRate5469Bps  = 0b01000
	AirDataRate3125Bps  = 0b01100
	AirDataRate1758Bps  = 0b10000
	AirDataRate31250Bps = 0b00001
	AirDataRate18750Bps = 0b00101
	AirDataRate10938Bps = 0b01001
	AirDataRate6250Bps  = 0b01101
	AirDataRate3516Bps  = 0b10001
	AirDataRate1953Bps  = 0b10101
	AirDataRate62500Bps = 0b00010
	AirDataRate37500Bps = 0b00110
	AirDataRate21875Bps = 0b01010
	AirDataRate12500Bps = 0b01110
	AirDataRate7031Bps  = 0b10010
	AirDataRate3906Bps  = 0b10110
	AirDataRate2148Bps  = 0b11010
)

// SubPacket represents sub packet size of E220
const (
	SubPacket200Bytes = 0b00
	SubPacket128Bytes = 0b01
	SubPacket64Bytes  = 0b10
	SubPacket32Bytes  = 0b11
)

// RSSIAmbient represents disable or enable of RSSI ambient of E220
const (
	RSSIAmbientDisable = 0b0
	RSSIAmbientEnable  = 0b1
)

// TransmitPower represents transmit power of E220
const (
	TransmitPowerUnavailable = 0b00
	TransmitPower13Dbm       = 0b01
	TransmitPower7Dbm        = 0b10
	TransmitPower0Dbm        = 0b11
)

// RSSIByte represents disable or enable of RSSI byte of E220
const (
	RSSIByteDisable = 0b0
	RSSIByteEnable  = 0b1
)

// TransmitMethod represents transmit method of E220
const (
	TransmitMethodTransparent = 0b0
	TransmitMethodFixed       = 0b1
)

// WorCycle represents WOR cycle of E220
const (
	WorCycleSetting500ms  = 0b000
	WorCycleSetting1000ms = 0b001
	WorCycleSetting1500ms = 0b010
	WorCycleSetting2000ms = 0b011
	WorCycleSetting2500ms = 0b100
	WorCycleSetting3000ms = 0b101
	WorCycleSetting3500ms = 0b110
	WorCycleSetting4000ms = 0b111
)
