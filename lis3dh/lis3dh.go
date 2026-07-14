// Package lis3dh provides a driver for the LIS3DH digital accelerometer.
//
// Datasheet: https://www.st.com/resource/en/datasheet/lis3dh.pdf
package lis3dh // import "tinygo.org/x/drivers/lis3dh"

import (
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// Device wraps an I2C connection to a LIS3DH device.
type Device struct {
	bus     drivers.I2C
	address uint16
	r       Range
	accel   [6]byte // stored acceleration data (from the Update call)
}

// Driver configuration, used for the Configure call. All fields are optional.
type Config struct {
	Address uint16
}

// New creates a new LIS3DH connection. The I2C bus must already be configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{bus: bus, address: Address0}
}

// Configure sets up the device for communication
func (d *Device) Configure(config Config) error {
	if config.Address != 0 {
		d.address = config.Address
	}

	// enable all axes, normal mode
	err := legacy.WriteRegister(d.bus, uint8(d.address), REG_CTRL1, []byte{0x07})
	if err != nil {
		return err
	}

	// 400Hz rate
	err = d.SetDataRate(DATARATE_400_HZ)
	if err != nil {
		return err
	}

	// High res & BDU enabled
	err = legacy.WriteRegister(d.bus, uint8(d.address), REG_CTRL4, []byte{0x88})
	if err != nil {
		return err
	}

	// get current range
	d.r, err = d.ReadRange()
	return err
}

// Connected returns whether a LIS3DH has been found.
// It does a "who am I" request and checks the response.
func (d *Device) Connected() bool {
	data := []byte{0}
	err := legacy.ReadRegister(d.bus, uint8(d.address), WHO_AM_I, data)
	if err != nil {
		return false
	}
	return data[0] == 0x33
}

// SetDataRate sets the speed of data collected by the LIS3DH.
func (d *Device) SetDataRate(rate DataRate) error {
	ctl1 := []byte{0}
	err := legacy.ReadRegister(d.bus, uint8(d.address), REG_CTRL1, ctl1)
	if err != nil {
		return err
	}
	// mask off bits
	ctl1[0] &^= 0xf0
	ctl1[0] |= (byte(rate) << 4)
	return legacy.WriteRegister(d.bus, uint8(d.address), REG_CTRL1, ctl1)
}

// SetRange sets the G range for LIS3DH.
func (d *Device) SetRange(r Range) error {
	ctl := []byte{0}
	err := legacy.ReadRegister(d.bus, uint8(d.address), REG_CTRL4, ctl)
	if err != nil {
		return err
	}
	// mask off bits
	ctl[0] &^= 0x30
	ctl[0] |= (byte(r) << 4)
	err = legacy.WriteRegister(d.bus, uint8(d.address), REG_CTRL4, ctl)
	if err != nil {
		return err
	}

	// store the new range
	d.r = r

	return nil
}

// ReadRange returns the current G range for LIS3DH.
func (d *Device) ReadRange() (r Range, err error) {
	ctl := []byte{0}
	err = legacy.ReadRegister(d.bus, uint8(d.address), REG_CTRL4, ctl)
	if err != nil {
		return 0, err
	}
	// mask off bits
	r = Range(ctl[0] >> 4)
	r &= 0x03

	return r, nil
}

// ReadAcceleration reads the current acceleration from the device and returns
// it in µg (micro-gravity). When one of the axes is pointing straight to Earth
// and the sensor is not moving the returned value will be around 1000000 or
// -1000000.
func (d *Device) ReadAcceleration() (int32, int32, int32, error) {
	rawX, rawY, rawZ := d.ReadRawAcceleration()
	x, y, z := normalizeRange(rawX, rawY, rawZ, d.r)
	return x, y, z, nil
}

// ReadRawAcceleration returns the raw x, y and z axis from the LIS3DH
func (d *Device) ReadRawAcceleration() (x int16, y int16, z int16) {
	legacy.WriteRegister(d.bus, uint8(d.address), REG_OUT_X_L|0x80, nil)

	data := []byte{0, 0, 0, 0, 0, 0}
	d.bus.Tx(d.address, nil, data)

	x = int16((uint16(data[1]) << 8) | uint16(data[0]))
	y = int16((uint16(data[3]) << 8) | uint16(data[2]))
	z = int16((uint16(data[5]) << 8) | uint16(data[4]))

	return
}

// Update the sensor values of the 'which' parameter. Only acceleration is
// supported at the moment.
func (d *Device) Update(which drivers.Measurement) error {
	if which&drivers.Acceleration != 0 {
		// Read raw acceleration values and store them in the driver.
		err := legacy.WriteRegister(d.bus, uint8(d.address), REG_OUT_X_L|0x80, nil)
		if err != nil {
			return err
		}
		err = d.bus.Tx(d.address, nil, d.accel[:])
		if err != nil {
			return err
		}
	}
	return nil
}

// Acceleration returns the last read acceleration in µg (micro-gravity).
// When one of the axes is pointing straight to Earth and the sensor is not
// moving the returned value will be around 1000000 or -1000000.
func (d *Device) Acceleration() (x, y, z int32) {
	// Extract the raw 16-bit values.
	rawX := int16((uint16(d.accel[1]) << 8) | uint16(d.accel[0]))
	rawY := int16((uint16(d.accel[3]) << 8) | uint16(d.accel[2]))
	rawZ := int16((uint16(d.accel[5]) << 8) | uint16(d.accel[4]))

	// Normalize these values, to be in µg (micro-gravity).
	return normalizeRange(rawX, rawY, rawZ, d.r)
}

// Convert raw 16-bit values to normalized 32-bit values while avoiding floats
// and divisions.
func normalizeRange(rawX, rawY, rawZ int16, r Range) (x, y, z int32) {
	// We're going to convert the 16-bit raw values to values in the range
	// -1000_000..1000_000. For now we're going to assume a range of 16G, we'll
	// adjust that range later.
	// The formula is derived as follows, and carefully selected to avoid
	// overflow and integer divisions (the division will be optimized to a
	// bitshift):
	//   x = x * 1000_000      / 2048
	//   x = x * (1000_000/64) / (2048/64)
	//   x = x * 15625         / 32
	x = int32(rawX) * 15625 / 32
	y = int32(rawY) * 15625 / 32
	z = int32(rawZ) * 15625 / 32

	// Now we need to normalize the three values, since we assumed 16G before.
	shift := uint32(0)
	switch r {
	case RANGE_16_G:
		shift = 0
	case RANGE_8_G:
		shift = 1
	case RANGE_4_G:
		shift = 2
	case RANGE_2_G:
		shift = 3
	}
	x >>= shift
	y >>= shift
	z >>= shift

	return
}
