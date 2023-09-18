// Package pcf8523 implements a driver for the PCF8523 CMOS Real-Time Clock (RTC)
//
// Datasheet: https://www.nxp.com/docs/en/data-sheet/PCF8523.pdf
package pcf8523

import (
	"time"
	"tinygo.org/x/drivers"
)

type Device struct {
	bus     drivers.I2C
	Address uint8
}

func New(i2c drivers.I2C) Device {
	return Device{
		bus:     i2c,
		Address: DefaultAddress,
	}
}

// Reset resets the device according to the datasheet section 8.3
// This does not wipe the time registers, but resets control registers.
func (d *Device) Reset() (err error) {
	return d.bus.Tx(uint16(d.Address), []byte{rControl1, 0x58}, nil)
}

// SetPowerManagement configures how the device makes use of the backup battery, see
// datasheet section 8.5
func (d *Device) SetPowerManagement(b PowerManagement) error {
	return d.setRegister(rControl3, byte(b)<<5, 0xE0)
}

func (d *Device) setRegister(reg uint8, value, mask uint8) error {
	var buf [1]byte
	err := d.bus.Tx(uint16(d.Address), []byte{reg}, buf[:])
	if err != nil {
		return err
	}
	buf[0] = (value & mask) | (buf[0] & (^mask))
	return d.bus.Tx(uint16(d.Address), []byte{reg, buf[0]}, nil)
}

// SetTime sets the time and date
func (d *Device) SetTime(t time.Time) error {
	buf := []byte{
		rSeconds,
		bin2bcd(t.Second()),
		bin2bcd(t.Minute()),
		bin2bcd(t.Hour()),
		bin2bcd(t.Day()),
		bin2bcd(int(t.Weekday())),
		bin2bcd(int(t.Month())),
		bin2bcd(t.Year() - 2000),
	}

	return d.bus.Tx(uint16(d.Address), buf, nil)
}

// ReadTime returns the date and time
func (d *Device) ReadTime() (time.Time, error) {
	buf := make([]byte, 9)
	err := d.bus.Tx(uint16(d.Address), []byte{rSeconds}, buf)
	if err != nil {
		return time.Time{}, err
	}

	seconds := bcd2bin(buf[0] & 0x7F)
	minute := bcd2bin(buf[1] & 0x7F)
	hour := bcd2bin(buf[2] & 0x3F)
	day := bcd2bin(buf[3] & 0x3F)
	//skipping weekday buf[4]
	month := time.Month(bcd2bin(buf[5] & 0x1F))
	year := int(bcd2bin(buf[6])) + 2000

	t := time.Date(year, month, day, hour, minute, seconds, 0, time.UTC)
	return t, nil
}

// bin2bcd converts binary to BCD
func bin2bcd(dec int) uint8 {
	return uint8(dec + 6*(dec/10))
}

// bcd2bin converts BCD to binary
func bcd2bin(bcd uint8) int {
	return int(bcd - 6*(bcd>>4))
}
