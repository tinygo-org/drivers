// Package bmm150 provides a driver for the Bosch BMM150 3-axis magnetometer.
//
// Datasheet: https://www.bosch-sensortec.com/media/boschsensortec/downloads/datasheets/bst-bmm150-ds001.pdf
//
// The trim compensation is ported from the Bosch BMM150 Sensor API at
// https://github.com/boschsensortec/BMM150_SensorAPI (see LICENSE in this directory).
package bmm150

import (
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

var (
	errNotConnected = errors.New("bmm150: not connected")
	errOverflow     = errors.New("bmm150: measurement overflow")
)

// Config holds the measurement settings. Zero values select the regular
// preset at 10Hz.
type Config struct {
	DataRate DataRate
	Preset   Preset
}

type trimData struct {
	x1, y1 int8
	x2, y2 int8
	xy1    uint8
	xy2    int8
	z1     uint16
	z2, z3 int16
	z4     int16
	xyz1   uint16
}

// Device is a BMM150 magnetometer connected over I2C.
type Device struct {
	bus     drivers.I2C
	Address uint16
	trim    trimData
	buf     [trimLength]byte
	wbuf    [2]byte

	x, y, z int32
}

// New creates a new BMM150 connection. The I2C bus must already be configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) *Device {
	return &Device{bus: bus, Address: DefaultAddress}
}

// Connected returns whether a BMM150 has been found. The chip ID can only be
// read after power on, so this moves the device from suspend to sleep mode.
// Datasheet section 4.2.1.
func (d *Device) Connected() bool {
	if d.write(regPowerCtrl, powerOn) != nil {
		return false
	}
	time.Sleep(3 * time.Millisecond)
	id, err := d.read1(regChipID)
	return err == nil && id == chipID
}

// Configure resets the device, reads the trim data, and starts continuous
// measurements.
func (d *Device) Configure(cfg Config) error {
	if !d.Connected() {
		return errNotConnected
	}

	if err := d.write(regPowerCtrl, softReset|powerOn); err != nil {
		return err
	}
	time.Sleep(3 * time.Millisecond)

	if err := d.readTrim(); err != nil {
		return err
	}

	// Register values for nXY = 1+2*REP_XY and nZ = 1+REP_Z. Datasheet table 6.
	var repXY, repZ uint8
	switch cfg.Preset {
	case LowPower:
		repXY, repZ = 0x01, 0x02
	case Enhanced:
		repXY, repZ = 0x07, 0x1A
	case HighAccuracy:
		repXY, repZ = 0x17, 0x52
	default:
		repXY, repZ = 0x04, 0x0E
	}
	if err := d.write(regRepXY, repXY); err != nil {
		return err
	}
	if err := d.write(regRepZ, repZ); err != nil {
		return err
	}

	return d.write(regOpMode, uint8(cfg.DataRate&0x07)<<3|opModeNormal)
}

// Update reads the magnetic field from the sensor.
func (d *Device) Update(which drivers.Measurement) error {
	if which&drivers.MagneticField == 0 {
		return nil
	}

	data := d.buf[:8]
	if err := d.readn(regDataX, data); err != nil {
		return err
	}

	// X and Y are 13 bit, Z is 15 bit, RHALL is 14 bit. Datasheet section 5.2.
	rawX := int16(uint16(data[1])<<8|uint16(data[0])) >> 3
	rawY := int16(uint16(data[3])<<8|uint16(data[2])) >> 3
	rawZ := int16(uint16(data[5])<<8|uint16(data[4])) >> 1
	rhall := (uint16(data[7])<<8 | uint16(data[6])) >> 2

	x, okX := d.compensateXY(rawX, rhall, d.trim.x1, d.trim.x2)
	y, okY := d.compensateXY(rawY, rhall, d.trim.y1, d.trim.y2)
	z, okZ := d.compensateZ(rawZ, rhall)
	if !okX || !okY || !okZ {
		return errOverflow
	}

	// Compensated values are in 1/16 µT.
	d.x = x * 1000 / 16
	d.y = y * 1000 / 16
	d.z = z * 1000 / 16
	return nil
}

// MagneticField returns the last read magnetic field in nT (nanotesla).
// 1 G (gauss) = 100_000 nT.
func (d *Device) MagneticField() (x, y, z int32) {
	return d.x, d.y, d.z
}

func (d *Device) readTrim() error {
	t := d.buf[:trimLength]
	if err := d.readn(regTrim, t); err != nil {
		return err
	}

	// Offsets from regTrim. Bosch BMM150 Sensor API bmm150_defs.h.
	d.trim = trimData{
		x1:   int8(t[0]),
		y1:   int8(t[1]),
		z4:   int16(uint16(t[6])<<8 | uint16(t[5])),
		x2:   int8(t[7]),
		y2:   int8(t[8]),
		z2:   int16(uint16(t[12])<<8 | uint16(t[11])),
		z1:   uint16(t[14])<<8 | uint16(t[13]),
		xyz1: uint16(t[16]&0x7F)<<8 | uint16(t[15]),
		z3:   int16(uint16(t[18])<<8 | uint16(t[17])),
		xy2:  int8(t[19]),
		xy1:  t[20],
	}
	return nil
}

func (d *Device) compensateXY(raw int16, rhall uint16, dig1, dig2 int8) (int32, bool) {
	if raw == overflowXY {
		return 0, false
	}
	r := rhall
	if r == 0 {
		r = d.trim.xyz1
	}
	if r == 0 {
		return 0, false
	}

	v := int32(int16(uint16(int32(d.trim.xyz1)*16384/int32(r)) - 0x4000))
	c := (int32(d.trim.xy2)*(v*v/128)+v*int32(d.trim.xy1)*128)/512 + 0x100000
	c = c * (int32(dig2) + 0xA0) / 4096
	return int32(int16(int32(raw)*c/8192)) + int32(dig1)*8, true
}

func (d *Device) compensateZ(raw int16, rhall uint16) (int32, bool) {
	t := &d.trim
	if raw == overflowZ || t.z2 == 0 || t.z1 == 0 || rhall == 0 || t.xyz1 == 0 {
		return 0, false
	}

	z0 := int16(rhall) - int16(t.xyz1)
	z1 := int32(t.z3) * int32(z0) / 4
	z2 := (int32(raw) - int32(t.z4)) * 32768
	z3 := int32(t.z1) * int32(int16(rhall)) * 2
	z4 := int16((z3 + 32768) / 65536)
	v := (z2 - z1) / (int32(t.z2) + int32(z4))

	// Saturate to the range of the Bosch API.
	if v > 32767 {
		v = 32767
	} else if v < -32767 {
		v = -32767
	}
	return v, true
}

func (d *Device) read1(reg uint8) (uint8, error) {
	err := d.readn(reg, d.buf[:1])
	return d.buf[0], err
}

func (d *Device) readn(reg uint8, data []byte) error {
	d.wbuf[0] = reg
	return d.bus.Tx(d.Address, d.wbuf[:1], data)
}

func (d *Device) write(reg, value uint8) error {
	d.wbuf[0] = reg
	d.wbuf[1] = value
	return d.bus.Tx(d.Address, d.wbuf[:2], nil)
}
