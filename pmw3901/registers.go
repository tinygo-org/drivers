package pmw3901

// Register map, see datasheet section 5.1.
const (
	RegProductID        = 0x00
	RegRevisionID       = 0x01
	RegMotion           = 0x02
	RegDeltaXL          = 0x03
	RegDeltaXH          = 0x04
	RegDeltaYL          = 0x05
	RegDeltaYH          = 0x06
	RegSqual            = 0x07
	RegRawDataSum       = 0x08
	RegMaximumRawData   = 0x09
	RegMinimumRawData   = 0x0A
	RegShutterLower     = 0x0B
	RegShutterUpper     = 0x0C
	RegObservation      = 0x15
	RegMotionBurst      = 0x16
	RegPowerUpReset     = 0x3A
	RegShutdown         = 0x3B
	RegRawDataGrab      = 0x58
	RegRawDataGrabStat  = 0x59
	RegInverseProductID = 0x5F
)

const (
	// ProductID is the value of RegProductID.
	ProductID = 0x49
	// InverseProductID is the value of RegInverseProductID.
	InverseProductID = 0xB6
)

const (
	powerUpResetValue = 0x5A

	regBank = 0x7F

	spiWrite = 0x80

	motionMOT = 0x80

	rawDataUpper = 0x40
	rawDataLower = 0x80
	rawDataMask  = 0xC0
)

// Performance optimization registers as register and value pairs, from the
// Bitcraze PMW3901 driver, see LICENSE-bitcraze.
// https://github.com/bitcraze/Bitcraze_PMW3901/blob/master/src/Bitcraze_PMW3901.cpp
const initRegisters1 = "" +
	"\x7F\x00\x61\xAD\x7F\x03\x40\x00\x7F\x05" +
	"\x41\xB3\x43\xF1\x45\x14\x5B\x32\x5F\x34" +
	"\x7B\x08\x7F\x06\x44\x1B\x40\xBF\x4E\x3F" +
	"\x7F\x08\x65\x20\x6A\x18\x7F\x09\x4F\xAF" +
	"\x5F\x40\x48\x80\x49\x80\x57\x77\x60\x78" +
	"\x61\x78\x62\x08\x63\x50\x7F\x0A\x45\x60" +
	"\x7F\x00\x4D\x11\x55\x80\x74\x1F\x75\x1F" +
	"\x4A\x78\x4B\x78\x44\x08\x45\x50\x64\xFF" +
	"\x65\x1F\x7F\x14\x65\x60\x66\x08\x63\x78" +
	"\x7F\x15\x48\x58\x7F\x07\x41\x0D\x43\x14" +
	"\x4B\x0E\x45\x0F\x44\x42\x4C\x80\x7F\x10" +
	"\x5B\x02\x7F\x07\x40\x41\x70\x00"

const initRegisters2 = "" +
	"\x32\x44\x7F\x07\x40\x40\x7F\x06\x62\xF0" +
	"\x63\x00\x7F\x0D\x48\xC0\x6F\xD5\x7F\x00" +
	"\x5B\xA0\x4E\xA8\x5A\x50\x40\x80"

// Frame capture setup as register and value pairs, from frame_capture in the
// Pimoroni pmw3901-python driver, see LICENSE-pimoroni.
// https://github.com/pimoroni/pmw3901-python/blob/main/pmw3901/__init__.py
const frameRegisters1 = "" +
	"\x7F\x07\x4C\x00\x7F\x08\x6A\x38\x7F\x00" +
	"\x55\x04\x40\x80\x4D\x11"

const frameRegisters2 = "\x7F\x00\x58\xFF"
