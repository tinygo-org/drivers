package amg88xx

// The I2C address which this device listens to.
const AddressHigh = 0x69
const AddressLow = 0x68

const (
	PCTL         = 0x00
	RST          = 0x01
	FPSC         = 0x02
	INTC         = 0x03
	STAT         = 0x04
	SCLR         = 0x05
	AVE          = 0x07
	INTHL        = 0x08
	INTHH        = 0x09
	INTLL        = 0x0A
	INTLH        = 0x0B
	IHYSL        = 0x0C
	IHYSH        = 0x0D
	TTHL         = 0x0E
	TTHH         = 0x0F
	INT_OFFSET   = 0x010
	PIXEL_OFFSET = 0x80

	// power modes
	NORMAL_MODE = 0x00
	SLEEP_MODE  = 0x01
	STAND_BY_60 = 0x20
	STAND_BY_10 = 0x21

	// resets
	FLAG_RESET    = 0x30
	INITIAL_RESET = 0x3F

	// frame rates
	FPS_10 = 0x00
	FPS_1  = 0x01

	// interrupt modes
	DIFFERENCE     InterruptMode = 0x00
	ABSOLUTE_VALUE InterruptMode = 0x01

	PIXEL_TEMP_CONVERSION = 250
	THERMISTOR_CONVERSION = 625
)
