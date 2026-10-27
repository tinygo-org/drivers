package ili9325

// Registers from the ILI9325 datasheet register list.
const (
	DRIVER_OUTPUT_CONTROL    = 0x01
	LCD_DRIVING_CONTROL      = 0x02
	ENTRY_MODE               = 0x03
	RESIZE_CONTROL           = 0x04
	DISPLAY_CONTROL1         = 0x07
	DISPLAY_CONTROL2         = 0x08
	DISPLAY_CONTROL3         = 0x09
	DISPLAY_CONTROL4         = 0x0A
	RGB_INTERFACE_CONTROL1   = 0x0C
	FRAME_MARKER_POSITION    = 0x0D
	RGB_INTERFACE_CONTROL2   = 0x0F
	POWER_CONTROL1           = 0x10
	POWER_CONTROL2           = 0x11
	POWER_CONTROL3           = 0x12
	POWER_CONTROL4           = 0x13
	GRAM_HORIZONTAL_ADDRESS  = 0x20
	GRAM_VERTICAL_ADDRESS    = 0x21
	WRITE_DATA_TO_GRAM       = 0x22
	POWER_CONTROL7           = 0x29
	FRAME_RATE_CONTROL       = 0x2B
	HORIZONTAL_START         = 0x50
	HORIZONTAL_END           = 0x51
	VERTICAL_START           = 0x52
	VERTICAL_END             = 0x53
	GATE_SCAN_CONTROL1       = 0x60
	GATE_SCAN_CONTROL2       = 0x61
	GATE_SCAN_CONTROL3       = 0x6A
	PANEL_INTERFACE_CONTROL1 = 0x90
	PANEL_INTERFACE_CONTROL2 = 0x92
)

// Entry mode (R03h) bits.
const (
	entryModeBGR = 1 << 12
	entryModeIDY = 1 << 5
	entryModeIDX = 1 << 4
	entryModeAM  = 1 << 3
)
