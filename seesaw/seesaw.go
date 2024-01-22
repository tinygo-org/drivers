// Package seesaw provides a driver implementation to communicate with Adafruit's seesaw chip.
// There are many Adafruit boards that use a seesaw. Soil moisture sensors, LED keyboards, etc.
//
//   - Documentation: https://learn.adafruit.com/adafruit-seesaw-atsamd09-breakout/overview
//   - Arduino driver: https://github.com/adafruit/Adafruit_Seesaw
//   - Seesaw firmware: https://github.com/adafruit/seesaw
package seesaw

import (
	"errors"
	"strconv"
	"time"

	"tinygo.org/x/drivers"
)

// DefaultAddress is the I2C address the chips have by default. Most boards
// built on top of it come with their own respective default addresses.
const DefaultAddress = 0x49

// DefaultReadDelay is an empirically determined delay used when reading from the device,
// the one from the official library seems to be too short (250us)
const DefaultReadDelay = 100 * time.Millisecond

const (
	seesawHwIdCodeSAMD09  = 0x55 // HW ID code for SAMD09
	seesawHwIdCodeTINY8x7 = 0x87 // HW ID code for ATtiny817
)

type Device struct {
	bus       drivers.I2C
	Address   uint16
	ReadDelay time.Duration
}

func New(bus drivers.I2C) *Device {
	return &Device{
		bus:       bus,
		Address:   DefaultAddress,
		ReadDelay: DefaultReadDelay,
	}
}

// SoftReset triggers a soft-reset of seesaw and waits for it to be ready
func (d *Device) SoftReset() error {
	err := d.WriteRegister(ModuleStatusBase, FunctionStatusSwrst, 0xFF)
	if err != nil {
		return errors.New("failed sending soft-reset command: " + err.Error())
	}

	return d.waitForReset()
}

func (d *Device) waitForReset() error {
	// give the device a little bit of time to reset
	time.Sleep(time.Second)

	var lastErr error
	tries := 0
	for ; tries < 20; tries++ {
		_, err := d.readHardwareID()
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(20 * time.Millisecond)
	}
	return errors.New("failed to wait for device to start: " + lastErr.Error())
}

func (d *Device) readHardwareID() (byte, error) {
	hwid, err := d.ReadRegister(ModuleStatusBase, FunctionStatusHwId)
	if err != nil {
		return 0, err
	}

	if hwid == seesawHwIdCodeSAMD09 || hwid == seesawHwIdCodeTINY8x7 {
		return hwid, nil
	}

	return 0, errors.New("unknown hardware ID: " + strconv.FormatUint(uint64(hwid), 16))
}

// WriteRegister writes a single seesaw register
func (d *Device) WriteRegister(module ModuleBaseAddress, function FunctionAddress, value byte) error {
	var buf [3]byte
	buf[0] = byte(module)
	buf[1] = byte(function)
	buf[2] = value
	return d.bus.Tx(d.Address, buf[:], nil)
}

// ReadRegister reads a single register from seesaw
func (d *Device) ReadRegister(module ModuleBaseAddress, function FunctionAddress) (byte, error) {
	var buf [1]byte
	err := d.Read(module, function, buf[:])
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

// Read reads a number of bytes from the device after sending the read command and waiting 'ReadDelay'. The delays depend
// on the module and function and are documented in the seesaw datasheet
func (d *Device) Read(module ModuleBaseAddress, function FunctionAddress, buf []byte) error {
	var cmd [2]byte
	cmd[0] = byte(module)
	cmd[1] = byte(function)

	err := d.bus.Tx(d.Address, cmd[:], nil)
	if err != nil {
		return err
	}

	// This is needed for the client seesaw device to flush its RX buffer and process the command.
	// See seesaw datasheet for timings for specific modules.
	time.Sleep(d.ReadDelay)

	return d.bus.Tx(d.Address, nil, buf)
}

// Write writes data into a given module and function
func (d *Device) Write(module ModuleBaseAddress, function FunctionAddress, buf []byte) error {
	cmd := []byte{byte(module), byte(function)}
	data := append(cmd, buf...)
	return d.bus.Tx(d.Address, data, nil)
}
