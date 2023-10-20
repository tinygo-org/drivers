package l3gd20

import (
	"encoding/binary"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

const (
	fifoLen = 32 * 3 * 2
)

type DevI2C struct {
	addr uint8
	// sensitivity or range.
	mul int32
	bus drivers.I2C
	buf [1]byte
	// gyro databuf.
	databuf [6]byte
	data    [3]int32
}

func NewI2C(bus drivers.I2C, addr uint8) *DevI2C {
	return &DevI2C{
		addr: addr,
		bus:  bus,
		mul:  sensMul250,
	}
}

// Initializes and configures the device.
func (d *DevI2C) Configure(cfg Config) error {
	err := cfg.validate()
	if err != nil {
		return err
	}
	err = d.Reboot()
	if err != nil {
		return err
	}
	// Reset then switch to normal mode and enable all three channels.
	err = d.write8(CTRL_REG1, 0)
	if err != nil {
		return err
	}
	err = d.write8(CTRL_REG1, reg1NormalBits)
	if err != nil {
		return err
	}
	// Reset REG2 values to default
	err = d.write8(CTRL_REG3, 0)
	if err != nil {
		return err
	}

	// Set sensitivity
	switch cfg.Range {
	case 1: // debugging range
		d.mul = 1
		cfg.Range = Range_2000
	case Range_250:
		d.mul = sensMul250
	case Range_500:
		d.mul = sensMul500
	case Range_2000:
		d.mul = sensMul2000
	default:
		return ErrBadRange
	}
	err = d.write8(CTRL_REG4, cfg.Range)
	if err != nil {
		return err
	}
	// Finally verify whomai register and return error if
	// board is not who it says it is. Some counterfeit boards
	// have incorrect whomai but can still be used.
	whoami, err := d.read8(WHOAMI)
	if err != nil {
		return err
	}
	if whoami != expectedWHOAMI && whoami != expectedWHOAMI_H {
		return ErrBadIdentity
	}
	return nil
}

func (d *DevI2C) Update() error {
	err := legacy.ReadRegister(d.bus, d.addr, OUT_X_L, d.databuf[:2])
	if err != nil {
		return err
	}
	err = legacy.ReadRegister(d.bus, d.addr, OUT_Y_L, d.databuf[2:4])
	if err != nil {
		return err
	}
	err = legacy.ReadRegister(d.bus, d.addr, OUT_Z_L, d.databuf[4:6])
	if err != nil {
		return err
	}
	x := int16(binary.LittleEndian.Uint16(d.databuf[0:]))
	y := int16(binary.LittleEndian.Uint16(d.databuf[2:]))
	z := int16(binary.LittleEndian.Uint16(d.databuf[4:]))
	d.data[0] = d.mul * int32(x)
	d.data[1] = d.mul * int32(y)
	d.data[2] = d.mul * int32(z)
	return nil
}

// Reboot sets reboot bit in CTRL_REG5 to true and unsets it.
func (d *DevI2C) Reboot() error {
	reg5, err := d.read8(CTRL_REG5)
	if err != nil {
		return err
	}
	// Write reboot bit and then unset it.
	err = d.write8(CTRL_REG5, reg5|reg5RebootBit)
	if err != nil {
		return err
	}
	time.Sleep(50 * time.Microsecond)
	return d.write8(CTRL_REG5, reg5&^reg5RebootBit)
}

// AngularVelocity returns result in microradians per second.
func (d *DevI2C) AngularVelocity() (x, y, z int32) {
	return d.data[0], d.data[1], d.data[2]
}

// func (d DevI2C) Update(measurement)

func (d DevI2C) read8(reg uint8) (byte, error) {
	err := legacy.ReadRegister(d.bus, d.addr, reg, d.buf[:1])
	return d.buf[0], err
}

func (d DevI2C) write8(reg uint8, val byte) error {
	d.buf[0] = val
	return legacy.WriteRegister(d.bus, d.addr, reg, d.buf[:1])
}
