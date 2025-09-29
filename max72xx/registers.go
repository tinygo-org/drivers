package max72xx

const (
	REG_NOOP         byte = 0x00
	REG_DIGIT0       byte = 0x01
	REG_DIGIT1       byte = 0x02
	REG_DIGIT2       byte = 0x03
	REG_DIGIT3       byte = 0x04
	REG_DIGIT4       byte = 0x05
	REG_DIGIT5       byte = 0x06
	REG_DIGIT6       byte = 0x07
	REG_DIGIT7       byte = 0x08
	REG_DECODE_MODE  byte = 0x09 // turn of for led matrix, turn on for digits
	REG_INTENSITY    byte = 0x0A
	REG_SCANLIMIT    byte = 0x0B
	REG_SHUTDOWN     byte = 0x0C // turn on for no shutdown mode
	REG_DISPLAY_TEST byte = 0x0F // turn off for no display test
)

// BCD (B) codes for the 7-segment display
// B coding uses the first 4 bits to represent the character and bit 7 to represent the dot.
// 0 - 9 represent the numbers 0 to 9
const (
	BcdDash  byte = 10
	BcdE     byte = 11
	BcdH     byte = 12
	BcdL     byte = 13
	BcdP     byte = 14
	BcdBlank byte = 15
	BcdDot   byte = 128 // 7 bit must be | with the character
)
