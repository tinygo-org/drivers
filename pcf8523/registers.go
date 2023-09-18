package pcf8523

const DefaultAddress = 0x68

// datasheet 8.5 Power management functions, table 11
type PowerManagement byte

const (
	PowerManagement_SwitchOver_ModeStandard_LowDetection PowerManagement = 0b000
	PowerManagement_SwitchOver_ModeDirect_LowDetection   PowerManagement = 0b001
	PowerManagement_VddOnly_LowDetection                 PowerManagement = 0b010
	PowerManagement_SwitchOver_ModeStandard              PowerManagement = 0b100
	PowerManagement_SwitchOver_ModeDirect                PowerManagement = 0b101
	PowerManagement_VddOnly                              PowerManagement = 0b101
)

// constants for all internal registers
const (
	rControl1               = 0x00 // Control_1
	rControl2               = 0x01 // Control_2
	rControl3               = 0x02 // Control_3
	rSeconds                = 0x03 // Seconds
	rMinutes                = 0x04 // Minutes
	rHours                  = 0x05 // Hours
	rDays                   = 0x06 // Days
	rWeekdays               = 0x07 // Weekdays
	rMonths                 = 0x08 // Months
	rYears                  = 0x09 // Years
	rMinuteAlarm            = 0x0A // Minute_alarm
	rHourAlarm              = 0x0B // Hour_alarm
	rDayAlarm               = 0x0C // Day_alarm
	rWeekdayAlarm           = 0x0D // Weekday_alarm
	rOffset                 = 0x0E // Offset
	rTimerClkoutControl     = 0x0F // Tmr_CLKOUT_ctrl
	rTimerAFrequencyControl = 0x10 // Tmr_A_freq_ctrl
	rTimerARegister         = 0x11 // Tmr_A_reg
	rTimerBFrequencyControl = 0x12 // Tmr_B_freq_ctrl
	rTimerBRegister         = 0x13 // Tmr_B_reg
)
