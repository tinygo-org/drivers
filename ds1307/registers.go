package ds1307

const (
	I2CAddress = 0x68
	TimeDate   = 0x00
	Control    = 0x7
	//CH is oscillator halt bit
	CH              = 0x7
	SRAMBeginAddres = 0x8
	SRAMEndAddress  = 0x3F
)

const (
	SQW_OFF   = 0x0
	SQW_1HZ   = 0x10
	SQW_4KHZ  = 0x11
	SQW_8KHZ  = 0x12
	SQW_32KHZ = 0x13
)
