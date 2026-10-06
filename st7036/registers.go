package st7036

// ST7036 commands. See https://www.crystalfontz.com/controllers/Sitronix/ST7036/
const (
	commandClear          = 0b00000001
	commandHome           = 0b00000010
	commandScroll         = 0b00010000
	commandDouble         = 0b00010000
	commandBias           = 0b00010100
	commandSetDisplayMode = 0b00001000

	blinkOn    = 0b00000001
	cursorOn   = 0b00000010
	displayOn  = 0b00000100
	setDDRAM   = 0b10000000
	setCGRAM   = 0b01000000
	entryShift = 0b00000010
	entryMove  = 0b00000100
)

// The function-set opcode includes instruction-set and double-height bits.
const defaultInstructionSetTemplate = 0b00111000
