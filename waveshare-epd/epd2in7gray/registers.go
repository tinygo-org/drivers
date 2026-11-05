package epd2in7gray

const (
	commandPanelSetting     = 0x00
	commandPowerSetting     = 0x01
	commandPowerOff         = 0x02
	commandPowerOn          = 0x04
	commandBoosterSoftStart = 0x06
	commandDeepSleep        = 0x07
	commandDataStart1       = 0x10
	commandDisplayRefresh   = 0x12
	commandDataStart2       = 0x13
	commandPartialRefresh   = 0x16
	commandLUTVCOM          = 0x20
	commandLUTWW            = 0x21
	commandLUTBW            = 0x22
	commandLUTWB            = 0x23
	commandLUTBB            = 0x24
	commandLUTExtra         = 0x25
	commandPLL              = 0x30
	commandVCOMDataInterval = 0x50
	commandResolution       = 0x61
	commandGetStatus        = 0x71
	commandVCOMDC           = 0x82
	commandPowerOptimize    = 0xf8
)

type registerSetting struct {
	command byte
	data    []byte
}

type initSetting struct {
	stage   string
	command byte
	data    []byte
}

// EPD_2IN7_Init_4Gray in Waveshare's EPD_2in7.cpp.
// https://github.com/waveshareteam/e-Paper/blob/a824b4f8f34dee7e721183c0154788dcde41c460/Arduino_R4/src/e-Paper/EPD_2in7.cpp
var initialization = [...]initSetting{
	{"power setting", commandPowerSetting, []byte{0x03, 0x00, 0x2b, 0x2b}},
	{"booster soft start", commandBoosterSoftStart, []byte{0x07, 0x07, 0x17}},
	{"power optimization 0x60", commandPowerOptimize, []byte{0x60, 0xa5}},
	{"power optimization 0x89", commandPowerOptimize, []byte{0x89, 0xa5}},
	{"power optimization 0x90", commandPowerOptimize, []byte{0x90, 0x00}},
	{"power optimization 0x93", commandPowerOptimize, []byte{0x93, 0x2a}},
	{"power optimization 0xa0", commandPowerOptimize, []byte{0xa0, 0xa5}},
	{"power optimization 0xa1", commandPowerOptimize, []byte{0xa1, 0x00}},
	{"power optimization 0x73", commandPowerOptimize, []byte{0x73, 0x41}},
	{"partial refresh disabled", commandPartialRefresh, []byte{0x00}},
	{"power on", commandPowerOn, nil},
	{"panel setting 0xbf", commandPanelSetting, []byte{0xbf}},
	{"PLL setting 0x90", commandPLL, []byte{0x90}},
	{"resolution", commandResolution, nil},
	{"VCOM DC setting 0x12", commandVCOMDC, []byte{0x12}},
	{"data interval 0x97", commandVCOMDataInterval, []byte{0x97}},
}
