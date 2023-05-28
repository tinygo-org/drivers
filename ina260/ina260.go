package ina260

import (
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// Device wraps an I2C connection to an INA260 device.
type Device struct {
	bus     drivers.I2C
	Address uint16
}

// Config holds the configuration of the INA260 device.
type Config struct {
	// One of AVGMODE_XXX
	AverageMode byte

	// One of CONVTIME_XXXXUSEC
	VoltConvTime byte

	// One of CONVTIME_XXXXUSEC
	CurrentConvTime byte

	// Multiple of MODE_XXXX
	Mode byte
}

// New creates a new INA260 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
	}
}

// Configure sets up the device.
//
// This only needs to be called to override built-in defaults. By default,
// the device starts with:
//
// * AverageMode = AVGMODE_1
// * VoltConvTime = CONVTIME_1100USEC
// * CurrentConvTime = CONVTIME_1100USEC
// * Mode = MODE_CONTINUOUS | MODE_VOLTAGE | MODE_CURRENT
func (d *Device) Configure(cfg Config) {
	var val uint16

	val = uint16(cfg.AverageMode&0x7) << 9
	val |= uint16(cfg.VoltConvTime&0x7) << 6
	val |= uint16(cfg.CurrentConvTime&0x7) << 3
	val |= uint16(cfg.Mode & 0x7)

	d.WriteRegister(REG_CONFIG, val)
}

// Resets the device, setting all registers to default values
func (d *Device) Reset() {
	d.WriteRegister(REG_CONFIG, 0x8000)
}

// Connected returns whether an INA260 has been found.
func (d *Device) Connected() bool {
	return d.ReadRegister(REG_MANF_ID) == MANF_ID &&
		(d.ReadRegister(REG_DIE_ID)&DEVICE_ID_MASK) == DEVICE_ID
}

// Gets the measured current in µA (max resolution 1.25mA)
func (d *Device) Current() int32 {
	val := d.ReadRegister(REG_CURRENT)

	if val&0x8000 == 0 {
		return int32(val) * 1250
	}

	// Two's complement, convert to signed int
	return -(int32(^val) + 1) * 1250
}

// Gets the measured voltage in µV (max resolution 1.25mV)
func (d *Device) Voltage() int32 {
	val := d.ReadRegister(REG_BUSVOLTAGE)

	if val&0x8000 == 0 {
		return int32(val) * 1250
	}

	// Two's complement, convert to signed int
	return -(int32(^val) + 1) * 1250
}

// Gets the measured power in µW (max resolution 10mW)
func (d *Device) Power() int32 {
	return int32(d.ReadRegister(REG_POWER)) * 10000
}

// Read a register
func (d *Device) ReadRegister(reg uint8) uint16 {
	data := []byte{0, 0}
	legacy.ReadRegister(d.bus, uint8(d.Address), reg, data)
	return (uint16(data[0]) << 8) | uint16(data[1])
}

// Write to a register
func (d *Device) WriteRegister(reg uint8, v uint16) {
	data := []byte{0, 0}
	data[0] = byte(v >> 8)
	data[1] = byte(v & 0xff)

	legacy.WriteRegister(d.bus, uint8(d.Address), reg, data)
}
