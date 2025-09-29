// Driver works for max7219 and 7221
// Datasheet: https://datasheets.maximintegrated.com/en/ds/MAX7219-MAX7221.pdf
package max72xx

import (
	"machine"

	"tinygo.org/x/drivers"
)

type Device struct {
	bus drivers.SPI
	cs  machine.Pin
	n   uint8 // Number of MAX7219 devices in series
}

const maxNumberOfDevices = 8

// NewDevice creates a new max7219 connection. The SPI wire must already be configured
// The SPI frequency must not be higher than 10MHz.
// parameter cs: the datasheet also refers to this pin as "load" pin.
func NewDevice(bus drivers.SPI, cs machine.Pin) *Device {
	return &Device{
		bus: bus,
		cs:  cs,
		n:   1,
	}
}

// NewDeviceN creates a new max7219 connection with n devices in series. The SPI wire must already be configured
func NewDeviceN(bus drivers.SPI, cs machine.Pin, n uint8) *Device {
	if n < 1 || n > maxNumberOfDevices {
		n = 1
	}

	return &Device{
		bus: bus,
		cs:  cs,
		n:   n,
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
	driver.writeToAll(REG_SCANLIMIT, digitNumber-1)
}

// SetIntensity sets the intensity of the diplays.
// There are 16 possible intensity levels. The valid range is 0x00-0x0F
func (driver *Device) SetIntensity(intensity uint8) {
	if intensity > 0x0F {
		intensity = 0x0F
	}
	driver.writeToAll(REG_INTENSITY, intensity)
}

// SetDecodeMode sets the decode mode for 7 segment displays.
// digitNumber = 1 -> 1 digit gets decoded
// digitNumber = 2 or 3, or 4 -> 4 digit are being decoded
// digitNumber = 8 -> 8 digits are being decoded
// digitNumber 0 || digitNumber > 8 -> no decoding is being used
func (driver *Device) SetDecodeMode(digitNumber uint8) {
	switch digitNumber {
	case 1: // only decode first digit
		driver.writeToAll(REG_DECODE_MODE, 0x01)
	case 2, 3, 4: //  decode digits 3-0
		driver.writeToAll(REG_DECODE_MODE, 0x0F)
	case 8: // decode 8 digits
		driver.writeToAll(REG_DECODE_MODE, 0xFF)
	default:
		driver.writeToAll(REG_DECODE_MODE, 0x00)
	}
}

// StartShutdownMode sets the IC into a low power shutdown mode.
func (driver *Device) StartShutdownMode() {
	driver.writeToAll(REG_SHUTDOWN, 0x00)
}

// StartShutdownMode sets the IC into normal operation mode.
func (driver *Device) StopShutdownMode() {
	driver.writeToAll(REG_SHUTDOWN, 0x01)
}

// StartDisplayTest starts a display test.
func (driver *Device) StartDisplayTest() {
	driver.writeToAll(REG_DISPLAY_TEST, 0x01)
}

// StopDisplayTest stops the display test and gets into normal operation mode.
func (driver *Device) StopDisplayTest() {
	driver.writeToAll(REG_DISPLAY_TEST, 0x00)
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

// WriteCommandN sends a command to a specific MAX7219 in the chain
//
// The MAX7219 uses a 16-bit serial protocol divided into two 8-bit segments:
// - The first 8 bits (D15 to D8) represent the address of the register to write to.
// - The second 8 bits (D7 to D0) represent the data to be written to that register.
//
// Register Addresses:
// - 0x00: No-Op Register
// - 0x01 to 0x08: Digit Registers (control individual digits/segments)
// - 0x09: Decode Mode Register (configures decoding mode for digits)
// - 0x0A: Intensity Register (sets brightness level)
// - 0x0B: Scan Limit Register (sets number of digits to display)
// - 0x0C: Shutdown Register (controls shutdown mode)
// - 0x0F: Display Test Register (tests the display)
//
// Examples:
// To set the intensity to a medium level, send the following 16-bit data:
//
//	WriteCommandN(0, Command{Register: REG_INTENSITY, Data: 10})
//
// Example: To set digit 1 on the first max72xx to display the number 5 and DP, send the following 16-bit data:
//
//	WriteCommandN(0, REG_DIGIT1, 5 | BcdDot})
func (driver *Device) WriteCommandN(deviceNum uint8, register, data byte) {
	if deviceNum >= driver.n {
		deviceNum = 0
	}

	tmp := make([]struct {
		Register byte
		Data     byte
	}, driver.n)
	for i := range tmp {
		tmp[i].Register = REG_NOOP
		tmp[i].Data = 0x00
	}
	tmp[deviceNum].Data = data
	tmp[deviceNum].Register = register

	driver.cs.Low()
	for _, d := range tmp {
		driver.writeByte(d.Register)
		driver.writeByte(d.Data)
	}
	driver.cs.High()
}

// writeToAll sends the same command to all devices in the chain
func (driver *Device) writeToAll(register, data byte) {
	tmp := make([]struct {
		Register byte
		Data     byte
	}, driver.n)
	for i := range tmp {
		tmp[i].Register = register
		tmp[i].Data = data
	}

	driver.cs.Low()
	for _, d := range tmp {
		driver.writeByte(d.Register)
		driver.writeByte(d.Data)
	}
	driver.cs.High()
}
