package gdew0154m09

const (
	commandPanelSetting       = 0x00
	commandPowerOff           = 0x02
	commandPowerOn            = 0x04
	commandDeepSleep          = 0x07
	commandDisplayRefresh     = 0x12
	commandDataStartOld       = 0x10
	commandDataStartNew       = 0x13
	commandVCOMDataInterval   = 0x50
	commandResolutionSetting  = 0x61
	commandTCONSetting        = 0x60
	commandPowerSaving        = 0xe3
	commandPowerSequence      = 0xf3
	commandBoosterSoftStart   = 0xb6
	commandUnknownAA          = 0xaa
	commandUnknownE9          = 0xe9
	commandFITIInternalCode   = 0x4d
	deepSleepCheckCode        = 0xa5
	defaultVCOMDataInterval   = 0xd7
	deepSleepVCOMDataInterval = 0xf7
)
