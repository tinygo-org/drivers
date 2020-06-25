package adt7410 // import "tinygo.org/x/drivers/adt7410"

import (
	"machine"
	"time"
)

type Error uint8

const (
	ErrInvalidID Error = 0x1
)

func (e Error) Error() string {
	switch e {
	case ErrInvalidID:
		return "Invalid chip ID"
	default:
		return "Unknown error"
	}
}

type Device struct {
	bus  *machine.I2C
	buf  []byte
	addr uint8
}

// New returns ADT7410 device for the provided I2C bus and address. The ADT7410
// has a default address of 0x48 (1001000).  The last 2 bits of the address
// can be set using by connecting to the A1 and A0 pins to VDD or GND (for a
// total of up to 4 devices on a I2C bus).  Also note that 10k pullups are
// recommended for the SDA and SCL lines.
func New(i2c *machine.I2C, addressBits uint8) *Device {
	return &Device{
		bus:  i2c,
		buf:  make([]byte, 2),
		addr: Address | (addressBits & 0x3),
	}
}

func (dev *Device) Configure() (err error) {

	// verify the chip ID
	// TODO: According to datasheet, the check below should work; however
	//       this does not seem to be working right, but is not exactly
	//       necessary, so can revisit later to see if there is a bug
	//id := dev.ReadByte(RegID) & 0xF8
	//if id != 0xC8 {
	//	err = ErrInvalidID
	//}

	// reset the chip
	dev.writeByte(RegReset, 0xFF)
	time.Sleep(10 * time.Millisecond)
	return

}

// ReadTemperature returns the temperature in celsius milli degrees (°C/1000)
func (d *Device) ReadTemperature() (temperature int32, err error) {
	return (int32(d.readUint16(RegTempValueMSB)) * 1000) / 128, nil
}

// ReadTempC returns the value in the temperature value register, in Celcius
func (d *Device) ReadTempC() float32 {
	t := d.readUint16(RegTempValueMSB)
	return float32(int(t)) / 128.0
}

// ReadTempF returns the value in the temperature value register, in Fahrenheit
func (d *Device) ReadTempF() float32 {
	return d.ReadTempC()*1.8 + 32.0
}

func (d *Device) writeByte(reg uint8, data byte) {
	d.buf[0] = reg
	d.buf[1] = data
	d.bus.Tx(uint16(d.addr), d.buf, nil)
}

func (d *Device) readByte(reg uint8) byte {
	d.bus.ReadRegister(d.addr, reg, d.buf)
	return d.buf[0]
}

func (d *Device) readUint16(reg uint8) uint16 {
	d.bus.ReadRegister(d.addr, reg, d.buf)
	return uint16(d.buf[0])<<8 | uint16(d.buf[1])
}
