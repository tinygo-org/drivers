package ds3231

// The I2C address which this device listens to.
const Address = 0x68

// Registers
const (
	REG_TIMEDATE = 0x00
	REG_ALARMONE = 0x07
	REG_ALARMTWO = 0x0B

	REG_CONTROL = 0x0E
	REG_STATUS  = 0x0F
	REG_AGING   = 0x10

	REG_TEMP = 0x11

	REG_ALARMONE_SIZE = 4
	REG_ALARMTWO_SIZE = 3

	// DS3231 Control Register Bits
	A1IE  = 0
	A2IE  = 1
	INTCN = 2
	RS1   = 3
	RS2   = 4
	CONV  = 5
	BBSQW = 6
	EOSC  = 7

	// DS3231 Status Register Bits
	A1F     = 0
	A2F     = 1
	BSY     = 2
	EN32KHZ = 3
	OSF     = 7

	AlarmFlag_Alarm1    = 0x01
	AlarmFlag_Alarm2    = 0x02
	AlarmFlag_AlarmBoth = 0x03

	None          Mode = 0
	BatteryBackup Mode = 1
	Clock         Mode = 2
	AlarmOne      Mode = 3
	AlarmTwo      Mode = 4
	ModeAlarmBoth Mode = 5
)
