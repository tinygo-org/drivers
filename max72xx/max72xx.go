// Driver works for max7219 and 7221
// Datasheet: https://datasheets.maximintegrated.com/en/ds/MAX7219-MAX7221.pdf
package max72xx

import (
	"machine"
)

type Device interface {
	Configure()
	WriteCommand(register, data byte)

	StartShutdownMode()
	StopShutdownMode()
	StartDisplayTest()
	StopDisplayTest()
	SetDecodeMode(digitNumber uint8)
	SetScanLimit(digitNumber uint8)
}

type device struct {
	bus  machine.SPI
	load machine.Pin // load
}

// NewDriver creates a new max7219 connection. The SPI wire must already be configured
// The SPI frequency must not be higher than 10MHz.
func NewDevice(load machine.Pin, bus machine.SPI) Device {
	return &device{
		load: load,
		bus:  bus,
	}
}

// Configure setups the pins.
func (driver *device) Configure() {
	outPutConfig := machine.PinConfig{Mode: machine.PinOutput}

	driver.load.Configure(outPutConfig)
}

// SetScanLimit sets the scan limit. Maximum is 8.
// Example: a 4 digit 7SegmentDisplay has a scan limit of 4
func (driver *device) SetScanLimit(digitNumber uint8) {
	driver.WriteCommand(byte(REG_SCANLIMIT), byte(digitNumber-1))
}

// SetDecodeMode sets the decode mode for 7 segment displays.
// digitNumber = 1 -> 1 digit gets decoded
// digitNumber = 2 or 3, or 4 -> 4 digit are being decoded
// digitNumber = 8 -> 8 digits are being decoded
// digitNumber 0 || digitNumber > 8 -> no decoding is being used
func (driver *device) SetDecodeMode(digitNumber uint8) {
	switch digitNumber {
	case 1: // only decode first digit
		driver.WriteCommand(REG_DECODE_MODE, 0x01)
	case 2, 3, 4: //  decode digits 3-0
		driver.WriteCommand(REG_DECODE_MODE, 0x0F)
	case 8: // decode 8 digits
		driver.WriteCommand(REG_DECODE_MODE, 0xFF)
	default:
		driver.WriteCommand(REG_DECODE_MODE, 0x00)
	}
}

// StartShutdownMode sets the IC into a low power shutdown mode.
func (driver *device) StartShutdownMode() {
	driver.WriteCommand(REG_SHUTDOWN, 0x00)
}

// StartShutdownMode sets the IC into normal operation mode.
func (driver *device) StopShutdownMode() {
	driver.WriteCommand(REG_SHUTDOWN, 0x01)
}

// StartDisplayTest starts a display test.
func (driver *device) StartDisplayTest() {
	driver.WriteCommand(REG_DISPLAY_TEST, 0x01)
}

// StopDisplayTest stops the display test and gets into normal operation mode.
func (driver *device) StopDisplayTest() {
	driver.WriteCommand(REG_DISPLAY_TEST, 0x00)
}

func (driver *device) writeByte(data byte) {
	driver.bus.Transfer(data)
}

// WriteCommand write data to a given register.
func (driver *device) WriteCommand(register, data byte) {
	driver.load.Low()
	driver.writeByte(register)
	driver.writeByte(data)
	driver.load.High()
}
