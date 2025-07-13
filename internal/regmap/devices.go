package regmap

import (
	"encoding/binary"
	"io"

	"tinygo.org/x/drivers"
)

// Device8 implements common logic to most 8-bit peripherals with an I2C or SPI bus.
type Device8 struct {
	buf [10]byte
}

// clear zeroes Device8's buffers.
func (d *Device8) clear() {
	d.buf = [10]byte{}
}

// I2C methods.

func (d *Device8) Read8I2C(bus drivers.I2C, i2cAddr uint16, addr uint8) (byte, error) {
	d.buf[0] = addr
	err := bus.Tx(i2cAddr, d.buf[0:1], d.buf[1:2])
	return d.buf[1], err
}

func (d *Device8) Read16I2C(bus drivers.I2C, i2cAddr uint16, addr uint8, order binary.ByteOrder) (uint16, error) {
	d.buf[0] = addr
	err := bus.Tx(i2cAddr, d.buf[0:1], d.buf[1:3])
	return order.Uint16(d.buf[1:3]), err
}

func (d *Device8) Read32I2C(bus drivers.I2C, i2cAddr uint16, addr uint8, order binary.ByteOrder) (uint32, error) {
	d.buf[0] = addr
	err := bus.Tx(i2cAddr, d.buf[0:1], d.buf[1:5])
	return order.Uint32(d.buf[1:5]), err
}

func (d *Device8) ReadDataI2C(bus drivers.I2C, i2cAddr uint16, addr uint8, dataDestination []byte) error {
	d.buf[0] = addr
	return bus.Tx(i2cAddr, d.buf[:1], dataDestination)
}

func (d *Device8) Write8I2C(bus drivers.I2C, i2cAddr uint16, addr, value uint8) error {
	d.buf[0] = addr
	d.buf[1] = value
	return bus.Tx(i2cAddr, d.buf[:2], nil)
}

func (d *Device8) Write16I2C(bus drivers.I2C, i2cAddr uint16, addr uint8, value uint16, order binary.ByteOrder) error {
	d.buf[0] = addr
	order.PutUint16(d.buf[1:3], value)
	return bus.Tx(i2cAddr, d.buf[0:3], nil)
}

func (d *Device8) Write32I2C(bus drivers.I2C, i2cAddr uint16, addr uint8, value uint32, order binary.ByteOrder) error {
	d.buf[0] = addr
	order.PutUint32(d.buf[1:5], value)
	return bus.Tx(i2cAddr, d.buf[0:5], nil)
}

// SPI methods.

func (d *Device8) Read8SPI(bus drivers.SPI, addr uint8) (byte, error) {
	d.clear()
	d.buf[0] = addr
	err := bus.Tx(d.buf[0:1], d.buf[1:2]) // We suppose data is returned after first byte in SPI.
	return d.buf[1], err
}

func (d *Device8) Read16SPI(bus drivers.SPI, addr uint8, order binary.ByteOrder) (uint16, error) {
	d.clear()
	d.buf[0] = addr
	err := bus.Tx(d.buf[0:3], d.buf[3:6]) // We suppose data is returned after first byte in SPI.
	return order.Uint16(d.buf[4:6]), err
}

func (d *Device8) Read32SPI(bus drivers.SPI, addr uint8, order binary.ByteOrder) (uint32, error) {
	d.clear()
	d.buf[0] = addr
	err := bus.Tx(d.buf[0:5], d.buf[5:10]) // We suppose data is returned after first byte in SPI.
	return order.Uint32(d.buf[6:10]), err
}

// ReadDataSPI reads data from a 8bit device address. It assumes data at register address is sent back
// from device after first byte is written as address.
// It needs the auxiliary buffer length to be large enough to contain both the write and read portions of buffer,
// so 2*(dataLength+1) < len(auxiliaryBuf) must hold.
func (d *Device8) ReadDataSPI(bus drivers.SPI, addr uint8, dataLength int, auxiliaryBuf []byte) ([]byte, error) {
	split := len(auxiliaryBuf) / 2
	if split < dataLength+1 {
		return nil, io.ErrShortBuffer
	}

	wbuf, rbuf := auxiliaryBuf[:split], auxiliaryBuf[split:]
	wbuf[0] = addr
	err := bus.Tx(wbuf, rbuf)
	return rbuf[1:], err
}

func (d *Device8) Write8SPI(bus drivers.SPI, addr, value uint8) error {
	d.clear()
	d.buf[0] = addr
	d.buf[1] = value
	return bus.Tx(d.buf[:2], nil)
}

func (d *Device8) Write16SPI(bus drivers.SPI, addr uint8, value uint16, order binary.ByteOrder) error {
	d.clear()
	d.buf[0] = addr
	order.PutUint16(d.buf[1:3], value)
	return bus.Tx(d.buf[:3], nil)
}

func (d *Device8) Write32SPI(bus drivers.SPI, addr uint8, value uint32, order binary.ByteOrder) error {
	d.clear()
	d.buf[0] = addr
	order.PutUint32(d.buf[1:5], value)
	return bus.Tx(d.buf[:5], nil)
}
