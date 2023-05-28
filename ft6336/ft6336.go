// Package ft6336 provides a driver for the FT6336 I2C Self-Capacitive touch
// panel controller.
//
// Datasheet: https://focuslcds.com/content/FT6236.pdf
package ft6336

import (
	"machine"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
	"tinygo.org/x/drivers/touch"
)

// Device wraps FT6336 I2C Self-Capacitive touch
type Device struct {
	bus     drivers.I2C
	buf     []byte
	Address uint8
	intPin  machine.Pin
}

// New returns FT6336 device for the provided I2C bus using default address.
func New(i2c drivers.I2C, intPin machine.Pin) *Device {
	return &Device{
		bus:     i2c,
		buf:     make([]byte, 11),
		Address: Address,
		intPin:  intPin,
	}
}

// Config contains settings for FT6636.
type Config struct {
}

// Configure the FT6336 device.
func (d *Device) Configure(config Config) error {
	d.write1Byte(0xA4, 0x00)
	d.intPin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	return nil
}

// SetGMode sets interrupt mode.
//
//	0x00 : Interrupt Polling mode
//	0x01 : Interrupt Trigger mode (default)
func (d *Device) SetGMode(v uint8) {
	d.write1Byte(RegGMode, v)
}

// GetGMode gets interrupt mode.
func (d *Device) GetGMode() uint8 {
	return d.read8bit(RegGMode)
}

// SetPeriodActive sets report rate in Active mode.
func (d *Device) SetPeriodActive(v uint8) {
	d.write1Byte(RegPeriodActive, v)
}

// GetPeriodActive gets report rate in Active mode.
func (d *Device) GetPeriodActive() uint8 {
	return d.read8bit(RegPeriodActive)
}

// GetFirmwareID gets firmware version.
func (d *Device) GetFirmwareID() uint8 {
	return d.read8bit(RegFirmid)
}

// Read reads the registers.
func (d *Device) Read() []byte {
	d.bus.Tx(uint16(d.Address), []byte{0x02}, d.buf[:])
	return d.buf[:]
}

// ReadTouchPoint reads a single touch.Point from the device. The maximum value
// for each touch.Point is 0xFFFF.
func (d *Device) ReadTouchPoint() touch.Point {
	d.Read()
	z := 0xFFFFF
	if d.buf[0] == 0 {
		z = 0
	}

	//Scale X&Y to 16 bit for consistency across touch drivers
	return touch.Point{
		X: (int(d.buf[1]&0x0F)<<8 + int(d.buf[2])) * ((1 << 16) / 320),
		Y: (int(d.buf[3]&0x0F)<<8 + int(d.buf[4])) * ((1 << 16) / 270),
		Z: z,
	}
}

// Touched returns if touched or not.
func (d *Device) Touched() bool {
	p := d.ReadTouchPoint()
	return p.Z > 0
}

func (d *Device) write1Byte(reg, data uint8) {
	legacy.WriteRegister(d.bus, d.Address, reg, []byte{data})
}

func (d *Device) read8bit(reg uint8) uint8 {
	legacy.ReadRegister(d.bus, d.Address, reg, d.buf[:1])
	return d.buf[0]
}
