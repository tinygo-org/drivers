// Package ds3231 provides a driver for the DS3231 RTC
//
// Datasheet:
// https://datasheets.maximintegrated.com/en/ds/DS3231.pdf
package ds3231 // import "tinygo.org/x/drivers/ds3231"

import (
	"encoding/binary"
	"errors"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/regmap"
)

type Mode uint8

// Device wraps an I2C connection to a DS3231 device.
type Device struct {
	bus     drivers.I2C
	Address uint16
	d       regmap.Device8I2C
}

// New creates a new DS3231 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	d := Device{
		bus:     bus,
		Address: Address,
	}
	d.Configure()
	return d
}

// Configure sets up the device for communication
func (d *Device) Configure() bool {
	d.d.SetBus(d.bus, d.Address, binary.BigEndian)
	return true
}

// IsTimeValid return true/false is the time in the device is valid
func (d *Device) IsTimeValid() bool {
	status, err := d.d.Read8(REG_STATUS)
	if err != nil {
		return false
	}
	return (status & (1 << OSF)) == 0x00
}

// IsRunning returns if the oscillator is running
func (d *Device) IsRunning() bool {
	control, err := d.d.Read8(REG_CONTROL)
	if err != nil {
		return false
	}
	return (control & (1 << EOSC)) == 0x00
}

// SetRunning starts the internal oscillator
func (d *Device) SetRunning(isRunning bool) error {
	control, err := d.d.Read8(REG_CONTROL)
	if err != nil {
		return err
	}
	if isRunning {
		control &^= uint8(1 << EOSC)
	} else {
		control |= 1 << EOSC
	}
	return d.d.Write8(REG_CONTROL, control)
}

// SetTime sets the date and time in the DS3231. The DS3231 hardware supports
// only a 2-digit year field, so the current year will be stored as an offset
// from the year 2000, which supports the year 2000 until 2100.
//
// The DS3231 also supports a one-bit 'century' flag which is set by the chip
// when the year field rolls over from 99 to 00. The current code interprets
// this flag to be the year 2100, which appears to extend the range of years
// until the year 2200. However the DS3231 does not incorporate the 'century'
// flag in its leap year calculation, so it will incorrectly identify the year
// 2100 as a leap year, causing it to increment from 2100-02-28 to 2100-02-29
// instead of 2100-03-01.
func (d *Device) SetTime(dt time.Time) error {
	status, err := d.d.Read8(REG_STATUS)
	if err != nil {
		return err
	}
	status &^= 1 << OSF
	if err = d.d.Write8(REG_STATUS, status); err != nil {
		return err
	}

	data := make([]uint8, 7)
	data[0] = uint8ToBCD(uint8(dt.Second()))
	data[1] = uint8ToBCD(uint8(dt.Minute()))
	data[2] = uint8ToBCD(uint8(dt.Hour()))

	year := uint8(dt.Year() - 2000)
	// This code interprets the centuryFlag to be the year 2100. Warning: The
	// DS3231 does not incorporate the centuryFlag in its leap year calculation.
	// It will increment from 2100-02-28 to 2100-02-29, which is incorrect because
	// the year 2100 is not a leap year in the Gregorian calendar.
	centuryFlag := uint8(0)
	if year >= 100 {
		year -= 100
		centuryFlag = 1 << 7
	}

	data[3] = uint8ToBCD(uint8(dt.Weekday()))
	data[4] = uint8ToBCD(uint8(dt.Day()))
	data[5] = uint8ToBCD(uint8(dt.Month()) | centuryFlag)
	data[6] = uint8ToBCD(year)

	return d.bus.Tx(d.Address, append([]byte{REG_TIMEDATE}, data...), nil)
}

// ReadTime returns the date and time
func (d *Device) ReadTime() (dt time.Time, err error) {
	data := make([]uint8, 7)
	if err = d.d.ReadData(REG_TIMEDATE, data); err != nil {
		return
	}

	second := bcdToInt(data[0] & 0x7F)
	minute := bcdToInt(data[1])
	hour := hoursBCDToInt(data[2])
	day := bcdToInt(data[4])
	monthRaw := data[5]
	year := bcdToInt(data[6]) + 2000
	if monthRaw&(1<<7) != 0x00 {
		year += 100
	}
	month := time.Month(bcdToInt(monthRaw & 0x7F))

	dt = time.Date(year, month, day, hour, minute, second, 0, time.UTC)
	return
}

// ReadTemperature returns the temperature in millicelsius (mC)
func (d *Device) ReadTemperature() (int32, error) {
	temp, err := d.d.Read16(REG_TEMP)
	if err != nil {
		return 0, err
	}
	return milliCelsius(temp), nil
}

// GetSqwPinMode returns the current square wave output frequency
func (d *Device) GetSqwPinMode() SqwPinMode {
	control, err := d.d.Read8(REG_CONTROL)
	if err != nil {
		return SQW_OFF
	}

	control &= 0x1C // turn off INTCON
	if control&0x04 != 0 {
		return SQW_OFF
	}

	return SqwPinMode(control)
}

// SetSqwPinMode sets the square wave output mode to the given frequency
func (d *Device) SetSqwPinMode(mode SqwPinMode) error {
	control, err := d.d.Read8(REG_CONTROL)
	if err != nil {
		return err
	}

	control &^= 0x04 // turn off INTCON
	control &^= 0x18 // set freq bits to 0

	control |= uint8(mode)

	return d.d.Write8(REG_CONTROL, control)
}

// SetAlarm1 sets alarm1 to the given time and mode
func (d *Device) SetAlarm1(dt time.Time, mode Alarm1Mode) error {
	control, err := d.d.Read8(REG_CONTROL)
	if err != nil {
		return err
	}
	if control&(1<<INTCN) == 0x00 {
		return errors.New("INTCN has to be disabled")
	}

	A1M1 := uint8((mode & 0x01) << 7)
	A1M2 := uint8((mode & 0x02) << 6)
	A1M3 := uint8((mode & 0x04) << 5)
	A1M4 := uint8((mode & 0x08) << 4)
	DY_DT := uint8((mode & 0x10) << 2)

	day := dt.Day()
	if DY_DT > 0 {
		day = dowToDS3231(int(dt.Weekday()))
	}

	alarm1 := uint32(uint8ToBCD(uint8(dt.Second()))|A1M1) << 24
	alarm1 |= uint32(uint8ToBCD(uint8(dt.Minute()))|A1M2) << 16
	alarm1 |= uint32(uint8ToBCD(uint8(dt.Hour()))|A1M3) << 8
	alarm1 |= uint32(uint8ToBCD(uint8(day)) | A1M4 | DY_DT)

	if err := d.d.Write32(REG_ALARMONE, alarm1); err != nil {
		return err
	}

	control |= AlarmFlag_Alarm1
	return d.d.Write8(REG_CONTROL, control)
}

// ReadAlarm1 returns the alarm1 time
func (d *Device) ReadAlarm1() (dt time.Time, err error) {
	data := make([]uint8, 4)
	if err = d.d.ReadData(REG_ALARMONE, data); err != nil {
		return
	}
	second := bcdToInt(data[0] & 0x7F)
	minute := bcdToInt(data[1] & 0x7F)
	hour := hoursBCDToInt(data[2] & 0x3F)

	isDayOfWeek := (data[3] & 0x40) >> 6
	var day int
	if isDayOfWeek > 0 {
		day = bcdToInt(data[3] & 0x0F)
	} else {
		day = bcdToInt(data[3] & 0x3F)
	}

	dt = time.Date(2000, 5, day, hour, minute, second, 0, time.UTC)
	return
}

// SetAlarm2 sets alarm2 to the given time and mode
func (d *Device) SetAlarm2(dt time.Time, mode Alarm2Mode) error {
	control, err := d.d.Read8(REG_CONTROL)
	if err != nil {
		return err
	}
	if control&(1<<INTCN) == 0x00 {
		return errors.New("INTCN has to be disabled")
	}

	A2M2 := uint8((mode & 0x01) << 7)
	A2M3 := uint8((mode & 0x02) << 6)
	A2M4 := uint8((mode & 0x04) << 5)
	DY_DT := uint8((mode & 0x08) << 3)

	day := dt.Day()
	if DY_DT > 0 {
		day = dowToDS3231(int(dt.Weekday()))
	}

	data := make([]uint8, 4)
	data[0] = uint8ToBCD(uint8(dt.Minute())) | A2M2
	data[1] = uint8ToBCD(uint8(dt.Hour())) | A2M3
	data[2] = uint8ToBCD(uint8(day)) | A2M4 | DY_DT
	if err = d.bus.Tx(d.Address, append([]byte{REG_ALARMTWO}, data...), nil); err != nil {
		return err
	}

	control |= AlarmFlag_Alarm2
	return d.d.Write8(REG_CONTROL, control)
}

// ReadAlarm2 returns the alarm2 time
func (d *Device) ReadAlarm2() (dt time.Time, err error) {
	data := make([]uint8, 3)
	if err = d.d.ReadData(REG_ALARMTWO, data); err != nil {
		return
	}
	minute := bcdToInt(data[0] & 0x7F)
	hour := hoursBCDToInt(data[1] & 0x3F)

	isDayOfWeek := (data[2] & 0x40) >> 6
	var day int
	if isDayOfWeek > 0 {
		day = bcdToInt(data[2] & 0x0F)
	} else {
		day = bcdToInt(data[2] & 0x3F)
	}

	dt = time.Date(2000, 5, day, hour, minute, 0, 0, time.UTC)
	return
}

// IsEnabledAlarm1 returns true when alarm1 is enabled
func (d *Device) IsEnabledAlarm1() bool {
	return d.isEnabledAlarm(1)
}

// SetEnabledAlarm1 sets the enabled status of alarm1
func (d *Device) SetEnabledAlarm1(enable bool) error {
	if enable {
		return d.enableAlarm(1)
	}
	return d.disableAlarm(1)
}

// IsEnabledAlarm2 returns true when alarm2 is enabled
func (d *Device) IsEnabledAlarm2() bool {
	return d.isEnabledAlarm(2)
}

// SetEnabledAlarm2 sets the enabled status of alarm2
func (d *Device) SetEnabledAlarm2(enable bool) error {
	if enable {
		return d.enableAlarm(2)
	}
	return d.disableAlarm(2)
}

// ClearAlarm1 clears status of alarm1
func (d *Device) ClearAlarm1() error {
	return d.clearAlarm(1)
}

// ClearAlarm2 clears status of alarm2
func (d *Device) ClearAlarm2() error {
	return d.clearAlarm(2)
}

// IsAlarm1Fired returns true when alarm1 is firing
func (d *Device) IsAlarm1Fired() bool {
	return d.isAlarmFired(1)
}

// IsAlarm2Fired returns true when alarm2 is firing
func (d *Device) IsAlarm2Fired() bool {
	return d.isAlarmFired(2)
}

// SetEnabled32K sets the enabled status of the 32KHz output
func (d *Device) SetEnabled32K(enable bool) error {
	status, err := d.d.Read8(REG_STATUS)
	if err != nil {
		return err
	}

	if enable {
		status |= 1 << EN32KHZ
	} else {
		status &^= 1 << EN32KHZ
	}

	return d.d.Write8(REG_STATUS, status)
}

// IsEnabled32K returns true when the 32KHz output is enabled
func (d *Device) IsEnabled32K() bool {
	status, err := d.d.Read8(REG_STATUS)
	if err != nil {
		return false
	}
	return (status & (1 << EN32KHZ)) != 0x00
}

func (d *Device) disableAlarm(alarm_num uint8) error {
	control, err := d.d.Read8(REG_CONTROL)
	if err != nil {
		return err
	}
	control &^= (1 << (alarm_num - 1))
	return d.d.Write8(REG_CONTROL, control)
}

func (d *Device) enableAlarm(alarm_num uint8) error {
	control, err := d.d.Read8(REG_CONTROL)
	if err != nil {
		return err
	}
	control |= (1 << (alarm_num - 1))
	return d.d.Write8(REG_CONTROL, control)
}

func (d *Device) isEnabledAlarm(alarm_num uint8) bool {
	control, err := d.d.Read8(REG_CONTROL)
	if err != nil {
		return false
	}
	return (control & (1 << (alarm_num - 1))) != 0x00
}

func (d *Device) clearAlarm(alarm_num uint8) error {
	status, err := d.d.Read8(REG_STATUS)
	if err != nil {
		return err
	}
	status &^= (1 << (alarm_num - 1))
	return d.d.Write8(REG_STATUS, status)
}

func (d *Device) isAlarmFired(alarm_num uint8) bool {
	status, err := d.d.Read8(REG_STATUS)
	if err != nil {
		return false
	}
	return (status & (1 << (alarm_num - 1))) != 0x00
}

// milliCelsius converts the raw temperature bytes (msb and lsb) from the DS3231
// into a 32-bit signed integer in units of milli Celsius (1/1000 deg C).
//
// According to the DS3231 datasheet: "Temperature is represented as a 10-bit
// code with a resolution of 0.25 deg C and is accessible at location 11h and
// 12h. The temperature is encoded in two's complement format. The upper 8 bits,
// the integer portion, are at location 11h and the lower 2 bits, the fractional
// portion, are in the upper nibble at location 12h."
//
// In other words, the msb and lsb bytes should be treated as a signed 16-bit
// integer in units of (1/256 deg C). It is possible to convert this into a
// 16-bit signed integer in units of centi Celsius (1/100 deg C) with no loss of
// precision or dynamic range. But for backwards compatibility, let's instead
// convert this into a 32-bit signed integer in units of milli Celsius.
func milliCelsius(tempBytes uint16) int32 {
	t256 := int16(uint16(tempBytes>>8)<<8 | uint16(tempBytes&0xFF))
	t1000 := int32(t256) / 64 * 250
	return t1000
}

// uint8ToBCD converts a byte to BCD for the DS3231
func uint8ToBCD(value uint8) uint8 {
	return value + 6*(value/10)
}

// bcdToInt converts BCD from the DS3231 to int
func bcdToInt(value uint8) int {
	return int(value - 6*(value>>4))
}

// hoursBCDToInt converts the BCD hours to int
func hoursBCDToInt(value uint8) (hour int) {
	if value&0x40 != 0x00 {
		hour = bcdToInt(value & 0x1F)
		if (value & 0x20) != 0x00 {
			hour += 12
		}
	} else {
		hour = bcdToInt(value)
	}
	return
}

// dowToDS3231 converts the day of the week to internal DS3231 format
func dowToDS3231(d int) int {
	if d == 0 {
		return 7
	}
	return d
}
