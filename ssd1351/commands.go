package ssd1351

// Commands
const (
	SET_COLUMN_ADDRESS          = 0x15
	SET_ROW_ADDRESS             = 0x75
	WRITE_RAM                   = 0x5C
	READ_RAM                    = 0x5D
	SET_REMAP_COLORDEPTH        = 0xA0
	SET_DISPLAY_START_LINE      = 0xA1
	SET_DISPLAY_OFFSET          = 0xA2
	SET_DISPLAY_MODE_ALLOFF     = 0xA4
	SET_DISPLAY_MODE_ALLON      = 0xA5
	SET_DISPLAY_MODE_RESET      = 0xA6
	SET_DISPLAY_MODE_INVERT     = 0xA7
	FUNCTION_SELECTION          = 0xAB
	SLEEP_MODE_DISPLAY_OFF      = 0xAE
	SLEEP_MODE_DISPLAY_ON       = 0xAF
	SET_PHASE_PERIOD            = 0xB1
	ENHANCED_DRIVING_SCHEME     = 0xB2
	SET_FRONT_CLOCK_DIV         = 0xB3
	SET_SEGMENT_LOW_VOLTAGE     = 0xB4
	SET_GPIO                    = 0xB5
	SET_SECOND_PRECHARGE_PERIOD = 0xB6
	GRAY_SCALE_LOOKUP           = 0xB8
	LINEAR_LUT                  = 0xB9
	SET_PRECHARGE_VOLTAGE       = 0xBB
	SET_VCOMH_VOLTAGE           = 0xBE
	SET_CONTRAST                = 0xC1
	MASTER_CONTRAST             = 0xC7
	SET_MUX_RATIO               = 0xCA
	NOP0                        = 0xD1
	NOP1                        = 0xE3
	SET_COMMAND_LOCK            = 0xFD
	HORIZONTAL_SCROLL           = 0x96
	STOP_MOVING                 = 0x9E
	START_MOVING                = 0x9F
)
