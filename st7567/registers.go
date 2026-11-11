package st7567

// Commands from the ST7567 datasheet V1.5, table "Instruction Table".
const (
	cmdDisplayOff     = 0xAE
	cmdDisplayOn      = 0xAF
	cmdStartLine      = 0x40
	cmdPageAddress    = 0xB0
	cmdColumnHigh     = 0x10
	cmdColumnLow      = 0x00
	cmdSEGNormal      = 0xA0
	cmdSEGReverse     = 0xA1
	cmdInverseOff     = 0xA6
	cmdInverseOn      = 0xA7
	cmdAllPixelOff    = 0xA4
	cmdBias9          = 0xA2
	cmdBias7          = 0xA3
	cmdReset          = 0xE2
	cmdCOMNormal      = 0xC0
	cmdCOMReverse     = 0xC8
	cmdPowerControl   = 0x28
	cmdRegulation     = 0x20
	cmdSetEV          = 0x81
	powerBoosterOn    = 0x04
	powerRegulatorOn  = 0x02
	powerFollowerOn   = 0x01
	defaultRegulation = 4  // 5.0
	defaultContrast   = 36 // V0 = 5.0 * (99 + 36) / 162 * 2.1 = 8.75V
	maxContrast       = 63
	maxColumns        = 132
)
