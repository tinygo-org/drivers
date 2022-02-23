// Driver works for max7219 and 7221
// Datasheet: https://datasheets.maximintegrated.com/en/ds/MAX7219-MAX7221.pdf
package max72xx

import (
	"machine"
)

type Device struct {
	bus machine.SPI
	cs  machine.Pin
}

// NewDriver creates a new max7219 connection. The SPI wire must already be configured
// The SPI frequency must not be higher than 10MHz.
// parameter cs: the datasheet also refers to this pin as "load" pin.
func NewDevice(bus machine.SPI, cs machine.Pin) *Device {
	return &Device{
		bus: bus,
		cs:  cs,
	}
}

// Configure setups the pins.
func (driver *Device) Configure() {
	outPutConfig := machine.PinConfig{Mode: machine.PinOutput}

	driver.cs.Configure(outPutConfig)
}

// SetScanLimit sets the scan limit. Maximum is 8.
// Example: a 4 digit 7SegmentDisplay has a scan limit of 4
func (driver *Device) SetScanLimit(digitNumber uint8) {
	driver.WriteCommand(REG_SCANLIMIT, digitNumber-1)
}

// SetIntensity sets the intensity of the diplays.
// There are 16 possible intensity levels. The valid range is 0x00-0x0F
func (driver *Device) SetIntensity(intensity uint8) {
	if intensity > 0x0F {
		intensity = 0x0F
	}
	driver.WriteCommand(REG_INTENSITY, intensity)
}

// SetDecodeMode sets the decode mode for 7 segment displays.
// digitNumber = 1 -> 1 digit gets decoded
// digitNumber = 2 or 3, or 4 -> 4 digit are being decoded
// digitNumber = 8 -> 8 digits are being decoded
// digitNumber 0 || digitNumber > 8 -> no decoding is being used
func (driver *Device) SetDecodeMode(digitNumber uint8) {
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
func (driver *Device) StartShutdownMode() {
	driver.WriteCommand(REG_SHUTDOWN, 0x00)
}

// StartShutdownMode sets the IC into normal operation mode.
func (driver *Device) StopShutdownMode() {
	driver.WriteCommand(REG_SHUTDOWN, 0x01)
}

// StartDisplayTest starts a display test.
func (driver *Device) StartDisplayTest() {
	driver.WriteCommand(REG_DISPLAY_TEST, 0x01)
}

// StopDisplayTest stops the display test and gets into normal operation mode.
func (driver *Device) StopDisplayTest() {
	driver.WriteCommand(REG_DISPLAY_TEST, 0x00)
}

func (driver *Device) writeByte(data byte) {
	driver.bus.Transfer(data)
}

// WriteCommand write data to a given register.
func (driver *Device) WriteCommand(register, data byte) {
	driver.cs.Low()
	driver.writeByte(register)
	driver.writeByte(data)
	driver.cs.High()
}
