package epd4in2

// Derived from https://github.com/waveshare/e-Paper/blob/master/Arduino/epd4in2/epd4in2.h

// Registers
const (
	// Display resolution
	EPD_WIDTH  = 400
	EPD_HEIGHT = 300

	// EPD4IN2 commands
	PANEL_SETTING                  = 0x00
	POWER_SETTING                  = 0x01
	POWER_OFF                      = 0x02
	POWER_OFF_SEQUENCE_SETTING     = 0x03
	POWER_ON                       = 0x04
	POWER_ON_MEASURE               = 0x05
	BOOSTER_SOFT_START             = 0x06
	DEEP_SLEEP                     = 0x07
	DATA_START_TRANSMISSION_1      = 0x10
	DATA_STOP                      = 0x11
	DISPLAY_REFRESH                = 0x12
	DATA_START_TRANSMISSION_2      = 0x13
	LUT_FOR_VCOM                   = 0x20
	LUT_WHITE_TO_WHITE             = 0x21
	LUT_BLACK_TO_WHITE             = 0x22
	LUT_WHITE_TO_BLACK             = 0x23
	LUT_BLACK_TO_BLACK             = 0x24
	PLL_CONTROL                    = 0x30
	TEMPERATURE_SENSOR_COMMAND     = 0x40
	TEMPERATURE_SENSOR_SELECTION   = 0x41
	TEMPERATURE_SENSOR_WRITE       = 0x42
	TEMPERATURE_SENSOR_READ        = 0x43
	VCOM_AND_DATA_INTERVAL_SETTING = 0x50
	LOW_POWER_DETECTION            = 0x51
	TCON_SETTING                   = 0x60
	RESOLUTION_SETTING             = 0x61
	GSST_SETTING                   = 0x65
	GET_STATUS                     = 0x71
	AUTO_MEASUREMENT_VCOM          = 0x80
	READ_VCOM_VALUE                = 0x81
	VCM_DC_SETTING                 = 0x82
	PARTIAL_WINDOW                 = 0x90
	PARTIAL_IN                     = 0x91
	PARTIAL_OUT                    = 0x92
	PROGRAM_MODE                   = 0xA0
	ACTIVE_PROGRAMMING             = 0xA1
	READ_OTP                       = 0xA2
	POWER_SAVING                   = 0xE3

	NO_ROTATION  Rotation = 0
	ROTATION_90  Rotation = 1 // 90 degrees clock-wise rotation
	ROTATION_180 Rotation = 2
	ROTATION_270 Rotation = 3
)
