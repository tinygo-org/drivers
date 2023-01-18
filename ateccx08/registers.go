package ateccx08

const (
	// Address is default I2C address.
	Address = 0x60
)

const (
	ATECCNone = 0
	ATECC508  = 0x5000
	ATECC608  = 0x6000
)

const (
	cmdAddress = 0x03
	cmdCounter = 0x24
	cmdGenKey  = 0x40
	cmdInfo    = 0x30
	cmdLock    = 0x17
	cmdNonce   = 0x16
	cmdRandom  = 0x1B
	cmdSHA     = 0x47
	cmdSign    = 0x41
	cmdWrite   = 0x12
	cmdRead    = 0x02
)

const (
	StatusSuccess         = 0x00
	StatusMiscompare      = 0x01
	StatusParseError      = 0x03
	StatusECCFault        = 0x05
	StatusSelfTestError   = 0x07
	StatusHealthTestError = 0x08
	StatusExecutionError  = 0x0f
	StatusAfterWake       = 0x11
	StatusWatchdogExpire  = 0xee
	StatusCRCError        = 0xff
)
