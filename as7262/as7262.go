// Package as7262 provides a driver for the as7262 6-channel visible spectral_id device
//
// Datasheet: https://ams.com/documents/20143/36005/AS7262_DS000486_5-00.pdf

package as7262 // import "tinygo.org/x/drivers/as7262"

import (
	"math"
	"time"
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

type Device struct {
	bus     drivers.I2C
	buf     []byte
	Address uint8
}

// New returns AS7262 device for the provided I2C bus using default address of 0x49 (1001001)
func New(i2c drivers.I2C) *Device {
	return &Device{
		bus:     i2c,
		buf:     make([]byte, 4),
		Address: DefaultAddress,
	}
}

// Configure soft resets device and returns
func (d *Device) Configure() (err error) {
	controlRegValue := d.readByte(ControlReg)
	controlRegValue |= 0x80

	// soft reset device 0x04:7
	d.writeByte(ControlReg, controlRegValue)
	time.Sleep(100 * time.Millisecond)
	return
}

// Connected returns if HardwareVersion (Device type == 01000000)
func (d *Device) Connected() bool {
	data := []byte{0}
	err := legacy.ReadRegister(d.bus, d.Address, HardwareVersionReg, data)
	if err != nil {
		return false
	}
	return data[0] == 0x40
}

/*
	Communication Functions
*/

// readByte read byte from device register
func (d *Device) readByte(reg uint8) byte {
	legacy.ReadRegister(d.bus, d.Address, reg, d.buf)
	return d.buf[0]
}

func (d *Device) readUint32(reg uint8) uint32 {
	legacy.ReadRegister(d.bus, d.Address, reg, d.buf)
	// shift bytes for uint32 from reg (start) + 3 more regs
	return uint32(d.buf[0])<<24 | uint32(d.buf[1])<<16 | uint32(d.buf[2])<<8 | uint32(d.buf[3])
}

// writeByte write byte to device register
func (d *Device) writeByte(reg uint8, data byte) {
	d.buf[0] = reg
	d.buf[1] = data
	d.bus.Tx(uint16(d.Address), d.buf, nil)
}

/*
	Data Reader Functions
*/

func (d *Device) ReadColors() [6]float32 {
	v := d.ReadViolet()
	b := d.ReadBlue()
	g := d.ReadGreen()
	y := d.ReadYellow()
	o := d.ReadOrange()
	r := d.ReadRed()
	return [6]float32{v, b, g, y, o, r}
}

// ReadRGB returns RGB Values
func (d *Device) ReadRGB() [3]float32 {
	return [3]float32{d.ReadRed(), d.ReadGreen(), d.ReadBlue()}
}

// ReadViolet returns Violet measurement
func (d *Device) ReadViolet() float32 {
	value := d.readUint32(VCalReg)
	return math.Float32frombits(value)
}

// ReadBlue returns Blue measurement
func (d *Device) ReadBlue() float32 {
	value := d.readUint32(BCalReg)
	return math.Float32frombits(value)
}

// ReadGreen returns Green measurement
func (d *Device) ReadGreen() float32 {
	value := d.readUint32(GCalReg)
	return math.Float32frombits(value)
}

// ReadYellow returns Yellow measurement
func (d *Device) ReadYellow() float32 {
	value := d.readUint32(YCalReg)
	return math.Float32frombits(value)
}

// ReadOrange returns Orange measurement
func (d *Device) ReadOrange() float32 {
	value := d.readUint32(OCalReg)
	return math.Float32frombits(value)
}

// ReadRed returns Red measurement
func (d *Device) ReadRed() float32 {
	value := d.readUint32(RCalReg)
	return math.Float32frombits(value)
}

// ReadTemp returns Temperature of Sensor in °C
func (d *Device) ReadTemp() int {
	value := d.readByte(TempReg)
	return int(value)
}
