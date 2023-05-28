// Package ds3231 provides a driver for the DS3231 RTC
//
// Datasheet:
// https://datasheets.maximintegrated.com/en/ds/DS3231.pdf
package ds3231 // import "tinygo.org/x/drivers/ds3231"

import (
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

type Mode uint8

// Device wraps an I2C connection to a DS3231 device.
type Device struct {
	bus     drivers.I2C
	Address uint16
}

// New creates a new DS3231 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
	}
}

// Configure sets up the device for communication
func (d *Device) Configure() bool {
	return true
}

// IsTimeValid return true/false is the time in the device is valid
func (d *Device) IsTimeValid() bool {
	data := []byte{0}
	err := legacy.ReadRegister(d.bus, uint8(d.Address), REG_STATUS, data)
	if err != nil {
		return false
	}
	return (data[0] & (1 << OSF)) == 0x00
}

// IsRunning returns if the oscillator is running
func (d *Device) IsRunning() bool {
	data := []uint8{0}
	err := legacy.ReadRegister(d.bus, uint8(d.Address), REG_CONTROL, data)
	if err != nil {
		return false
	}
	return (data[0] & (1 << EOSC)) == 0x00
}

// SetRunning starts the internal oscillator
func (d *Device) SetRunning(isRunning bool) error {
	data := []uint8{0}
	err := legacy.ReadRegister(d.bus, uint8(d.Address), REG_CONTROL, data)
	if err != nil {
		return err
	}
	if isRunning {
		data[0] &^= uint8(1 << EOSC)
	} else {
		data[0] |= 1 << EOSC
	}
	err = legacy.WriteRegister(d.bus, uint8(d.Address), REG_CONTROL, data)
	if err != nil {
		return err
	}
	return nil
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
	data := []byte{0}
	err := legacy.ReadRegister(d.bus, uint8(d.Address), REG_STATUS, data)
	if err != nil {
		return err
	}
	data[0] &^= 1 << OSF
	err = legacy.WriteRegister(d.bus, uint8(d.Address), REG_STATUS, data)
	if err != nil {
		return err
	}

	data = make([]uint8, 7)
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

	err = legacy.WriteRegister(d.bus, uint8(d.Address), REG_TIMEDATE, data)
	if err != nil {
		return err
	}

	return nil
}

// ReadTime returns the date and time
func (d *Device) ReadTime() (dt time.Time, err error) {
	data := make([]uint8, 7)
	err = legacy.ReadRegister(d.bus, uint8(d.Address), REG_TIMEDATE, data)
	if err != nil {
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
	data := make([]uint8, 2)
	err := legacy.ReadRegister(d.bus, uint8(d.Address), REG_TEMP, data)
	if err != nil {
		return 0, err
	}
	return milliCelsius(data[0], data[1]), nil
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
func milliCelsius(msb uint8, lsb uint8) int32 {
	t256 := int16(uint16(msb)<<8 | uint16(lsb))
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
