// Datasheet: https://datasheets.maximintegrated.com/en/ds/MAX7219-MAX7221.pdf
package max7219

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
func NewDriver(load machine.Pin, bus machine.SPI) Device {
	return &device{
		load: load,
		bus:  bus,
	}
}

func (driver *device) Configure() {
	outPutConfig := machine.PinConfig{Mode: machine.PinOutput}

	driver.load.Configure(outPutConfig)
}

func (driver *device) SetScanLimit(digitNumber uint8) {
	driver.WriteCommand(byte(REG_SCANLIMIT), byte(digitNumber-1))
}

func (driver *device) SetDecodeMode(digitNumber uint8) {
	switch digitNumber {
	case 1: // only decode first digit
		driver.WriteCommand(REG_DECODE_MODE, 0x01)
	case 2, 3, 4: //  decode digits 3-0
		driver.WriteCommand(REG_DECODE_MODE, 0x0F)
	case 8: // decode 8 digits
		driver.WriteCommand(REG_DECODE_MODE, 0xFF)
	}
}

func (driver *device) StartShutdownMode() {
	driver.WriteCommand(REG_SHUTDOWN, 0x00)

}

func (driver *device) StopShutdownMode() {
	driver.WriteCommand(REG_SHUTDOWN, 0x01)
}

func (driver *device) StartDisplayTest() {
	driver.WriteCommand(REG_DISPLAY_TEST, 0x01)

}

func (driver *device) StopDisplayTest() {
	driver.WriteCommand(REG_DISPLAY_TEST, 0x00)
}

func (driver *device) writeByte(data byte) {
	driver.bus.Transfer(data)
}

func (driver *device) WriteCommand(register, data byte) {
	driver.load.Low()
	driver.writeByte(register)
	driver.writeByte(data)
	driver.load.High()
}
