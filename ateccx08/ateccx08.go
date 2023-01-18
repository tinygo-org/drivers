// Package ateccx08 provides a driver for the ATECCx08 I2C cryptographic co-processor.
//
// Datasheet: https://datasheet.octopart.com/ATSAMA5D27-WLSOM1-Microchip-datasheet-149595509.pdf
package ateccx08 // import "tinygo.org/x/drivers/ateccx08"

import (
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

var (
	maxCommandTime = (200 + 50) * time.Millisecond
)

var (
	ErrWakeup          = errors.New("error on wakeup")
	ErrInvalidCRCCheck = errors.New("invalid CRC check")
	ErrLockFailed      = errors.New("locked failed")
)

type Device struct {
	bus     drivers.I2C
	Address uint8
}

// New returns ATECCx08 device for the provided I2C bus using default address.
func New(i2c drivers.I2C) *Device {
	return &Device{
		bus:     i2c,
		Address: Address,
	}
}

// Configure the ATECCx08 device.
func (d *Device) Configure() error {
	return nil
}

// Connected returns whether ATECCx08 has been found.
func (d *Device) Connected() bool {
	if err := d.Wakeup(); err != nil {
		return false
	}

	v, err := d.Version()
	if err != nil {
		return false
	}

	return (v == ATECC508 || v == ATECC608)
}

// Wakeup the ATECC by trying to write something to address 0x00
func (d *Device) Wakeup() error {
	d.bus.Tx(uint16(0x0), []byte{0x00}, nil)
	time.Sleep(1500 * time.Microsecond)
	d.bus.Tx(uint16(d.Address), []byte{0x00}, nil)
	time.Sleep(maxCommandTime)

	var status [4]byte
	if err := d.readResponse(status[:]); err != nil {
		return err
	}

	if status[0] != StatusAfterWake {
		return ErrWakeup
	}

	return nil
}

// Sleep puts the ATECC to sleep.
func (d *Device) Sleep() {
	d.bus.Tx(uint16(d.Address), []byte{0x01}, nil)
	time.Sleep(time.Millisecond)
}

// Idle puts the ATECC in idle mode.
func (d *Device) Idle() {
	d.bus.Tx(uint16(d.Address), []byte{0x02}, nil)
	time.Sleep(time.Millisecond)
}

type ATECCVersion uint16

func (at ATECCVersion) String() string {
	switch at {
	case ATECC508:
		return "ATECC508"
	case ATECC608:
		return "ATECC608"
	case ATECCNone:
		return "No ATECCx08"
	default:
		return "Unknown"
	}
}

// Version returns what version of ATECC is being used.
// Either ATECC508, ATECC608, or ATECCNone.
func (d *Device) Version() (ATECCVersion, error) {
	var version [4]byte
	d.Wakeup()
	defer d.Idle()

	d.sendCommand(cmdInfo, 0x00, 0, nil)

	time.Sleep(maxCommandTime)
	if err := d.readResponse(version[:]); err != nil {
		return ATECCNone, err
	}

	return ATECCVersion(uint16(version[2])<<8 | uint16(version[3])&0xf000), nil
}

// Random returns an array of 32 byte-sized random numbers.
func (d *Device) Random() ([32]byte, error) {
	var random [32]byte
	d.Wakeup()
	defer d.Idle()

	d.sendCommand(cmdRandom, 0x00, 0, nil)
	time.Sleep(23 * time.Millisecond)

	err := d.readResponse(random[:])
	return random, err
}

// Read reads from the device memory.
func (d *Device) Read(zone, address int, data []byte) error {
	d.Wakeup()
	defer d.Idle()

	d.sendCommand(cmdRead, byte(zone), uint16(address), nil)
	time.Sleep(5 * time.Millisecond)

	return d.readResponse(data)
}

// IsLocked checks to see if the ATECC is locked.
// Config zone (0) must be locked to generate random numbers.
func (d *Device) IsLocked() bool {
	return d.IsZoneLocked(0)
}

// IsZoneLocked checks to see if a specific zone in the ATECC is locked.
func (d *Device) IsZoneLocked(zone int) bool {
	var config [4]byte

	if zone < 0 || zone > 8 {
		return false
	}

	switch zone {
	case 0, 1:
		if err := d.Read(0, 0x15, config[:]); err != nil {
			return false
		}

		// LockConfig
		loc := 3

		// LockData
		if zone == 1 {
			loc = 2
		}

		if config[loc] == 0 {
			return true
		}

	default:
		if err := d.Read(0, 0x16, config[:]); err != nil {
			return false
		}

		slot := byte(zone<<2) | 2

		if (config[0] & slot) == 0 {
			return true
		}

		return false
	}

	return false
}

// Lock locks a zone in the device.
// Note that you cannot unlock a device zone once locked,
// so make sure you know what you are doing!
func (d *Device) Lock(zone int) error {
	var status [1]byte
	d.Wakeup()
	defer d.Idle()

	d.sendCommand(cmdLock, byte(zone)|0x80, 0, nil)
	time.Sleep(32 * time.Millisecond)

	d.readResponse(status[:])

	if status[0] != 0 {
		return ErrLockFailed
	}

	return nil
}

var cmdBuf [64]byte

func (d *Device) sendCommand(opcode, param1 byte, param2 uint16, data []byte) error {
	cmdBuf[0] = 0x03
	cmdBuf[1] = byte(8 + len(data) - 1)
	cmdBuf[2] = opcode
	cmdBuf[3] = param1
	cmdBuf[4] = byte(param2 & 0xff)
	cmdBuf[5] = byte(param2 >> 8)
	copy(cmdBuf[6:], data)

	crc := crc16(cmdBuf[1 : 6+len(data)])
	cmdBuf[6+len(data)] = crc[0]
	cmdBuf[6+len(data)+1] = crc[1]

	if err := d.bus.Tx(uint16(d.Address), cmdBuf[:6+len(data)+2], nil); err != nil {
		return err
	}

	time.Sleep(time.Millisecond)
	return nil
}

func (d *Device) readResponse(data []byte) error {
	var sz [1]byte
	if err := d.bus.Tx(uint16(d.Address), []byte{cmdAddress}, sz[:]); err != nil {
		return err
	}

	rx := make([]byte, sz[0])
	if err := d.bus.Tx(uint16(d.Address), []byte{cmdAddress}, rx); err != nil {
		return err
	}

	size := len(rx) - 2
	payload := rx[:size]
	payloaddata := rx[1:size]
	payloadcrc := rx[size:]

	crcCheck := crc16(payload)
	if !(crcCheck[0] == payloadcrc[0] &&
		crcCheck[1] == payloadcrc[1]) {
		return ErrInvalidCRCCheck
	}

	copy(data, payloaddata)
	return nil
}
