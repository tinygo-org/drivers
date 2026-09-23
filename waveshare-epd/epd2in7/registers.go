package epd2in7

// Derived from https://github.com/waveshare/e-Paper/blob/master/RaspberryPi_JetsonNano/c/lib/e-Paper/EPD_2in7.c

// Registers
const (
	// Display resolution
	EPD_WIDTH  = 176
	EPD_HEIGHT = 264

	// EPD2IN7 commands
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
	PARTIAL_DISPLAY_REFRESH        = 0x16
	VCOM_AND_DATA_INTERVAL_SETTING = 0x50
	LOW_POWER_DETECTION            = 0x51
	TCON_SETTING                   = 0x60
	RESOLUTION_SETTING             = 0x61
	GET_STATUS                     = 0x71
	AUTO_MEASUREMENT_VCOM          = 0x80
	READ_VCOM_VALUE                = 0x81
	VCM_DC_SETTING                 = 0x82
)
