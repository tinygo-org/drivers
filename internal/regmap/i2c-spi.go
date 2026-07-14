package regmap

import (
	"encoding/binary"

	"tinygo.org/x/drivers"
)

// Device8SPI implements common logic to most 8-bit peripherals with an SPI bus.
// All methods expect the target to support conventional register read and write operations
// where the first byte sent is the register address being accessed.
//
// All methods use an internal buffer and perform no dynamic memory allocation.
type Device8SPI struct {
	bus   drivers.SPI
	order binary.ByteOrder
	d     Device8
}

// SetBus sets the SPI bus and byte order for the Device8SPI.
//
// As a hint, most SPI devices use big-endian (MSB) byte order.
//   - Big endian: A value of 0x1234 is transmitted as 0x12 followed by 0x34.
//   - Little endian: A value of 0x1234 is transmitted as 0x34 followed by 0x12.
func (d *Device8SPI) SetBus(bus drivers.SPI, order binary.ByteOrder) {
	d.bus = bus
	d.order = order
}

// Read8 reads a single byte from register addr.
func (d *Device8SPI) Read8(addr uint8) (byte, error) {
	return d.d.Read8SPI(d.bus, addr)
}

// Read16 reads a 16-bit value from register addr.
func (d *Device8SPI) Read16(addr uint8) (uint16, error) {
	return d.d.Read16SPI(d.bus, addr, d.order)
}

// Read32 reads a 32-bit value from register addr.
func (d *Device8SPI) Read32(addr uint8) (uint32, error) {
	return d.d.Read32SPI(d.bus, addr, d.order)
}

// ReadData reads dataLength bytes from register addr. Due to the internal functioning of
// SPI, an auxiliary buffer must be provided to perform the operation and avoid memory allocation.
// The returned slice is a subslice of auxBuffer containing the read data.
func (d *Device8SPI) ReadData(addr uint8, datalength int, auxBuffer []byte) ([]byte, error) {
	return d.d.ReadDataSPI(d.bus, addr, datalength, auxBuffer)
}

// Write8 writes a single byte value to register addr.
func (d *Device8SPI) Write8(addr, value uint8) error {
	return d.d.Write8SPI(d.bus, addr, value)
}

// Write16 writes a 16-bit value to register addr.
func (d *Device8SPI) Write16(addr uint8, value uint16) error {
	return d.d.Write16SPI(d.bus, addr, value, d.order)
}

// Write32 writes a 32-bit value to register addr.
func (d *Device8SPI) Write32(addr uint8, value uint32) error {
	return d.d.Write32SPI(d.bus, addr, value, d.order)
}

// Device8I2C implements common logic to most 8-bit peripherals with an I2C bus.
// All methods expect the target to support conventional register read and write operations
// where the first byte sent is the register address being accessed.
//
// All methods use an internal buffer and perform no dynamic memory allocation.
type Device8I2C struct {
	bus     drivers.I2C
	i2cAddr uint16
	order   binary.ByteOrder
	d       Device8
}

// SetBus sets the I2C bus, device address, and byte order for the Device8I2C.
//
// As a hint, most I2C devices use big-endian (MSB) byte order.
//   - Big endian: A value of 0x1234 is transmitted as 0x12 followed by 0x34.
//   - Little endian: A value of 0x1234 is transmitted as 0x34 followed by 0x12.
func (d *Device8I2C) SetBus(bus drivers.I2C, i2cAddr uint16, order binary.ByteOrder) {
	d.bus = bus
	d.i2cAddr = i2cAddr
	d.order = order
}

// Read8 reads a single byte from register addr.
func (d *Device8I2C) Read8(addr uint8) (byte, error) {
	return d.d.Read8I2C(d.bus, d.i2cAddr, addr)
}

// Read16 reads a 16-bit value from register addr.
func (d *Device8I2C) Read16(addr uint8) (uint16, error) {
	return d.d.Read16I2C(d.bus, d.i2cAddr, addr, d.order)
}

// Read32 reads a 32-bit value from register addr.
func (d *Device8I2C) Read32(addr uint8) (uint32, error) {
	return d.d.Read32I2C(d.bus, d.i2cAddr, addr, d.order)
}

// ReadData reads dataLength bytes from register addr.
func (d *Device8I2C) ReadData(addr uint8, dataDestination []byte) error {
	return d.d.ReadDataI2C(d.bus, d.i2cAddr, addr, dataDestination)
}

// Write8 writes a single byte value to register addr.
func (d *Device8I2C) Write8(addr, value uint8) error {
	return d.d.Write8I2C(d.bus, d.i2cAddr, addr, value)
}

// Write16 writes a 16-bit value to register addr.
func (d *Device8I2C) Write16(addr uint8, value uint16) error {
	return d.d.Write16I2C(d.bus, d.i2cAddr, addr, value, d.order)
}

// Write32 writes a 32-bit value to register addr.
func (d *Device8I2C) Write32(addr uint8, value uint32) error {
	return d.d.Write32I2C(d.bus, d.i2cAddr, addr, value, d.order)
}
