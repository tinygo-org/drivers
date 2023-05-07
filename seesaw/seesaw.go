package seesaw

import (
	"errors"
	"strconv"
	"time"
)

const DefaultSeesawAddress = 0x49

// empirically determined delay, the one from the official library seems to be too short (250us)
const defaultDelay = 10 * time.Millisecond

const (
	seesawHwIdCodeSAMD09  = 0x55 // HW ID code for SAMD09
	seesawHwIdCodeTINY8x7 = 0x87 // HW ID code for ATtiny817
)

type Device struct {
	bus  I2C
	addr uint16
	hwid byte
}

func New(addr uint16, bus I2C) *Device {
	return &Device{
		bus:  bus,
		addr: addr,
	}
}

// Begin resets and initializes the seesaw chip
func (d *Device) Begin() error {

	err := d.SoftReset()
	if err != nil {
		return err
	}

	var lastErr error
	for i := 0; i < 10; i++ {
		hwid, err := d.ReadHardwareID()
		if err == nil {
			d.hwid = hwid
			lastErr = nil
			break
		}
		lastErr = err
		time.Sleep(10 * time.Millisecond)
	}

	if lastErr != nil {
		return lastErr
	}

	return nil
}

// ReadHardwareID reads the ID of the seesaw device
func (d *Device) ReadHardwareID() (byte, error) {
	hwid, err := d.ReadRegister(ModuleStatusBase, FunctionStatusHwId)
	if err != nil {
		return 0, err
	}

	if hwid == seesawHwIdCodeSAMD09 || hwid == seesawHwIdCodeTINY8x7 {
		return hwid, nil
	}

	return 0, errors.New("unknown hardware ID: " + strconv.Itoa(int(hwid)))
}

// SoftReset triggers a soft-reset of seesaw
func (d *Device) SoftReset() error {
	return d.WriteRegister(ModuleStatusBase, FunctionStatusSwrst, 0xFF)
}

// WriteRegister writes a single seesaw register
func (d *Device) WriteRegister(module ModuleBaseAddress, function FunctionAddress, value byte) error {
	buf := []byte{byte(module), byte(function), value}
	return d.bus.Tx(d.addr, buf, nil)
}

// ReadRegister reads a single register from seesaw
func (d *Device) ReadRegister(module ModuleBaseAddress, function FunctionAddress) (byte, error) {
	buf := make([]byte, 1)
	err := d.Read(module, function, buf, defaultDelay)
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

// Read reads a number of bytes from the device after sending the read command and waiting 'delay'. The delays depend
// on the module and function and are documented in the seesaw datasheet
func (d *Device) Read(module ModuleBaseAddress, function FunctionAddress, buf []byte, delay time.Duration) error {
	prefix := []byte{byte(module), byte(function)}
	err := d.bus.Tx(d.addr, prefix, nil)
	if err != nil {
		return err
	}

	//see seesaw datasheet
	time.Sleep(delay)

	return d.bus.Tx(d.addr, nil, buf)
}

func (d *Device) Write(module ModuleBaseAddress, function FunctionAddress, buf []byte) error {
	prefix := []byte{byte(module), byte(function)}
	data := append(prefix, buf...)
	return d.bus.Tx(d.addr, data, nil)
}
