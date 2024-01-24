package epd2in13

import "tinygo.org/x/drivers"

// Registers
const (
	DRIVER_OUTPUT_CONTROL                = 0x01
	BOOSTER_SOFT_START_CONTROL           = 0x0C
	GATE_SCAN_START_POSITION             = 0x0F
	DEEP_SLEEP_MODE                      = 0x10
	DATA_ENTRY_MODE_SETTING              = 0x11
	SW_RESET                             = 0x12
	TEMPERATURE_SENSOR_CONTROL           = 0x1A
	MASTER_ACTIVATION                    = 0x20
	DISPLAY_UPDATE_CONTROL_1             = 0x21
	DISPLAY_UPDATE_CONTROL_2             = 0x22
	WRITE_RAM                            = 0x24
	WRITE_VCOM_REGISTER                  = 0x2C
	WRITE_LUT_REGISTER                   = 0x32
	SET_DUMMY_LINE_PERIOD                = 0x3A
	SET_GATE_TIME                        = 0x3B
	BORDER_WAVEFORM_CONTROL              = 0x3C
	SET_RAM_X_ADDRESS_START_END_POSITION = 0x44
	SET_RAM_Y_ADDRESS_START_END_POSITION = 0x45
	SET_RAM_X_ADDRESS_COUNTER            = 0x4E
	SET_RAM_Y_ADDRESS_COUNTER            = 0x4F
	TERMINATE_FRAME_READ_WRITE           = 0xFF

	NO_ROTATION  = drivers.Rotation0
	ROTATION_90  = drivers.Rotation90 // 90 degrees clock-wise rotation
	ROTATION_180 = drivers.Rotation180
	ROTATION_270 = drivers.Rotation270
)
