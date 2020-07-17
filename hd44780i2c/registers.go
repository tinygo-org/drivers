package hd44780i2c

const (

	// commands
	DISPLAY_CLEAR        = 0x01
	CURSOR_HOME          = 0x02
	ENTRY_MODE           = 0x04
	DISPLAY_ON_OFF       = 0x08
	CURSOR_DISPLAY_SHIFT = 0x10
	FUNCTION_MODE        = 0x20
	CGRAM_SET            = 0x40
	DDRAM_SET            = 0x80

	// flags for display entry mode
	// CURSOR_DECREASE  = 0x00
	CURSOR_INCREASE = 0x02
	// DISPLAY_SHIFT    = 0x01
	DISPLAY_NO_SHIFT = 0x00

	// flags for display on/off control
	DISPLAY_ON       = 0x04
	DISPLAY_OFF      = 0x00
	CURSOR_ON        = 0x02
	CURSOR_OFF       = 0x00
	CURSOR_BLINK_ON  = 0x01
	CURSOR_BLINK_OFF = 0x00

	// flags for function set
	// DATA_LENGTH_8BIT = 0x10
	DATA_LENGTH_4BIT = 0x00
	TWO_LINE         = 0x08
	ONE_LINE         = 0x00
	FONT_5X10        = 0x04
	FONT_5X8         = 0x00

	// flags for backlight control
	BACKLIGHT_ON  = 0x08
	BACKLIGHT_OFF = 0x00

	En = 0x04 // Enable bit
	// Rw = 0x02 // Read/Write bit
	Rs = 0x01 // Register select bit
)
