// Package pcf8563 implements a driver for the PCF8563 CMOS Real-Time Clock (RTC)
//
// Datasheet: https://www.nxp.com/docs/en/data-sheet/PCF8563.pdf
//

package pcf8563

import (
	"time"

	"tinygo.org/x/drivers"
)

// Device wraps an I2C connection to a PCF8563 device.
type Device struct {
	bus     drivers.I2C
	Address uint16
}

// New creates a new PCF8563 connection. I2C bus must be already configured.
func New(i2c drivers.I2C) Device {
	return Device{
		bus:     i2c,
		Address: PCF8563_ADDR,
	}
}

// Reset resets the `control and status registers`. When this method is
// called, it writes `0x00` to the `control and status registers`. This will
// cause `Alarm` and `Timer` to become Inactive. Please refer to the datasheet
// for details.
func (d *Device) Reset() (err error) {
	return d.bus.Tx(d.Address, []byte{0x00, 0x00, 0x00}, nil)
}

// SetTime sets the time and date
func (d *Device) SetTime(t time.Time) error {
	var buf [9]byte
	buf[0] = 0x02
	buf[1] = decToBcd(t.Second())
	buf[2] = decToBcd(t.Minute())
	buf[3] = decToBcd(t.Hour())
	buf[4] = decToBcd(t.Day())
	buf[5] = decToBcd(int(t.Weekday() + 1))
	buf[6] = decToBcd(int(t.Month()))
	buf[7] = decToBcd(t.Year() - 2000)
	err := d.bus.Tx(d.Address, buf[:], nil)
	return err
}

// ReadTime returns the date and time
func (d *Device) ReadTime() (time.Time, error) {
	var buf [9]byte
	err := d.bus.Tx(d.Address, []byte{0x00}, buf[:])
	if err != nil {
		return time.Time{}, err
	}

	seconds := bcdToDec(buf[2] & 0x7F)
	minute := bcdToDec(buf[3] % 0x7F)
	hour := bcdToDec(buf[4] & 0x3F)
	day := bcdToDec(buf[5] & 0x3F)
	month := time.Month(bcdToDec(buf[7] & 0x0F))
	year := int(bcdToDec(buf[8])) + 2000

	t := time.Date(year, month, day, hour, minute, seconds, 0, time.UTC)
	return t, nil
}

// SetAlarm sets the alarm
func (d *Device) SetAlarm(t time.Time) error {
	var buf [5]byte
	buf[0] = 0x09
	buf[1] = RTC_ALARM_ENABLE | decToBcd(t.Minute())
	buf[2] = RTC_ALARM_ENABLE | decToBcd(t.Hour())
	buf[3] = RTC_ALARM_ENABLE | decToBcd(t.Day())
	buf[4] = RTC_ALARM_DISABLE
	err := d.bus.Tx(d.Address, buf[:], nil)
	if err != nil {
		return err
	}

	// enable alarm
	buf[0] = 0x01
	err = d.bus.Tx(d.Address, buf[:1], buf[1:])
	if err != nil {
		return err
	}

	buf[1] |= RTC_CTRL_AF
	err = d.bus.Tx(d.Address, buf[:2], nil)
	return err
}

// ClearAlarm disables alarm.
func (d *Device) ClearAlarm() error {
	var buf [2]byte
	buf[0] = 0x01
	err := d.bus.Tx(d.Address, buf[:1], buf[1:])
	if err != nil {
		return err
	}

	buf[1] &= ^uint8(RTC_CTRL_AF)
	err = d.bus.Tx(d.Address, buf[:], nil)
	return err
}

// EnableAlarmInterrupt enables alarm interrupt. When triggered, INT pin (3)
// goes low.
func (d *Device) EnableAlarmInterrupt() error {
	var buf [2]byte
	buf[0] = 0x01
	err := d.bus.Tx(d.Address, buf[:1], buf[1:])
	if err != nil {
		return err
	}

	buf[1] |= RTC_CTRL_AIE
	err = d.bus.Tx(d.Address, buf[:], nil)
	return err
}

// DisableAlarmInterrupt disable alarm interrupt.
func (d *Device) DisableAlarmInterrupt() error {
	var buf [2]byte
	buf[0] = 0x01
	err := d.bus.Tx(d.Address, buf[:1], buf[1:])
	if err != nil {
		return err
	}

	buf[1] &= ^uint8(RTC_CTRL_AIE)
	err = d.bus.Tx(d.Address, buf[:], nil)
	return err
}

// AlarmTriggered returns whether or not an Alarm has been triggered.
func (d *Device) AlarmTriggered() bool {
	var buf [1]byte
	buf[0] = 0x01
	err := d.bus.Tx(d.Address, buf[:], buf[:])
	if err != nil {
		return false
	}
	return (buf[0] & RTC_CTRL_AF) != 0
}

// SetTimer sets timer. The available durations are 1 to 127 seconds.  If any
// other value is specified, it will be truncated.
func (d *Device) SetTimer(dur time.Duration) error {
	var buf [3]byte

	sec := dur / time.Second
	if sec > 127 {
		sec = 127
	}

	// Treat as sec timer.
	buf[0] = 0x0E
	buf[1] = RTC_TIMER_1S
	buf[2] = byte(sec)
	err := d.bus.Tx(d.Address, buf[:], nil)
	if err != nil {
		return err
	}

	// enable alarm
	buf[0] = 0x01
	err = d.bus.Tx(d.Address, buf[:1], buf[1:])
	if err != nil {
		return err
	}

	buf[1] |= RTC_CTRL_TF
	err = d.bus.Tx(d.Address, buf[:2], nil)
	return err
}

// ClearTimer disables timer.
func (d *Device) ClearTimer() error {
	var buf [2]byte
	buf[0] = 0x01
	err := d.bus.Tx(d.Address, buf[:1], buf[1:])
	if err != nil {
		return err
	}

	buf[1] &= ^uint8(RTC_CTRL_TF)
	err = d.bus.Tx(d.Address, buf[:], nil)
	return err
}

// EnableTimerInterrupt enables timer interrupt. When triggered, INT pin (3)
// goes low.
func (d *Device) EnableTimerInterrupt() error {
	var buf [2]byte
	buf[0] = 0x01
	err := d.bus.Tx(d.Address, buf[:1], buf[1:])
	if err != nil {
		return err
	}

	buf[1] |= RTC_CTRL_TIE
	err = d.bus.Tx(d.Address, buf[:], nil)
	return err
}

// DisableTimerInterrupt disable timer interrupt.
func (d *Device) DisableTimerInterrupt() error {
	var buf [2]byte
	buf[0] = 0x01
	err := d.bus.Tx(d.Address, buf[:1], buf[1:])
	if err != nil {
		return err
	}

	buf[1] &= ^uint8(RTC_CTRL_TIE)
	err = d.bus.Tx(d.Address, buf[:], nil)
	return err
}

// TimerTriggered returns whether or not an Alarm has been triggered.
func (d *Device) TimerTriggered() bool {
	var buf [1]byte
	buf[0] = 0x01
	err := d.bus.Tx(d.Address, buf[:], buf[:])
	if err != nil {
		return false
	}
	return (buf[0] & RTC_CTRL_TF) != 0
}

// SetOscillatorFrequency sets output oscillator frequency
// Available modes: RTC_COT_DISABLE, RTC_COT_32KHZ, RTC_COT_1KHZ,
// RTC_COT_32Hz, RTC_COT_1HZ.
func (d *Device) SetOscillatorFrequency(sqw uint8) error {
	var buf [2]byte
	buf[0] = 0x0D
	buf[1] = sqw
	return d.bus.Tx(d.Address, buf[:], nil)
}

// decToBcd converts int to BCD
func decToBcd(dec int) uint8 {
	return uint8(dec + 6*(dec/10))
}

// bcdToDec converts BCD to int
func bcdToDec(bcd uint8) int {
	return int(bcd - 6*(bcd>>4))
}
