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

// SQW Pin Modes
type SqwPinMode uint8

const (
	SQW_OFF  SqwPinMode = 0x1C
	SQW_1HZ  SqwPinMode = 0x00
	SQW_1KHZ SqwPinMode = 0x08
	SQW_4KHZ SqwPinMode = 0x10
	SQW_8KHZ SqwPinMode = 0x18
)

// Alarm1 Modes define which parts of the set alarm time has to match the current timestamp of the clock device for
// alarm1 to fire
type Alarm1Mode uint8

const (
	// Alarm1 fires every second
	A1_PER_SECOND Alarm1Mode = 0x0F
	// Alarm1 fires when the seconds match
	A1_SECOND Alarm1Mode = 0x0E
	// Alarm1 fires when both seconds and minutes match
	A1_MINUTE Alarm1Mode = 0x0C
	// Alarm1 fires when seconds, minutes and hours match
	A1_HOUR Alarm1Mode = 0x08
	// Alarm1 fires when seconds, minutes, hours and the day of the month match
	A1_DATE Alarm1Mode = 0x00
	// Alarm1 fires when seconds, minutes, hours and the day of the week match
	A1_DAY Alarm1Mode = 0x10
)

// Alarm2 Modes define which parts of the set alarm time has to match the current timestamp of the clock device for
// alarm2 to fire.
//
// Alarm2 only supports matching down to the minute unlike alarm1 which supports matching down to the second.
type Alarm2Mode uint8

const (
	// Alarm2 fires every minute
	A2_PER_MINUTE Alarm2Mode = 0x07
	// Alarm2 fires when the minutes match
	A2_MINUTE Alarm2Mode = 0x06
	// Alarm2 fires when both minutes and hours match
	A2_HOUR Alarm2Mode = 0x04
	// Alarm2 fires when minutes, hours and the day of the month match
	A2_DATE Alarm2Mode = 0x00
	// Alarm2 fires when minutes, hours and the day of the week match
	A2_DAY Alarm2Mode = 0x08
)
