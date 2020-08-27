// Package ds3231 provides a driver for the DS3231 RTC
//
// Datasheet:
// https://datasheets.maximintegrated.com/en/ds/DS3231.pdf
package ds3231 // import "tinygo.org/x/drivers/ds3231"

import (
	"time"

	"tinygo.org/x/drivers"
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
	err := d.bus.ReadRegister(uint8(d.Address), REG_STATUS, data)
	if err != nil {
		return false
	}
	return (data[0] & (1 << OSF)) == 0x00
}

// IsRunning returns if the oscillator is running
func (d *Device) IsRunning() bool {
	data := []uint8{0}
	err := d.bus.ReadRegister(uint8(d.Address), REG_CONTROL, data)
	if err != nil {
		return false
	}
	return (data[0] & (1 << EOSC)) == 0x00
}

// SetRunning starts the internal oscillator
func (d *Device) SetRunning(isRunning bool) error {
	data := []uint8{0}
	err := d.bus.ReadRegister(uint8(d.Address), REG_CONTROL, data)
	if err != nil {
		return err
	}
	if isRunning {
		data[0] &^= uint8(1 << EOSC)
	} else {
		data[0] |= 1 << EOSC
	}
	err = d.bus.WriteRegister(uint8(d.Address), REG_CONTROL, data)
	if err != nil {
		return err
	}
	return nil
}

// SetTime sets the date and time in the DS3231
func (d *Device) SetTime(dt time.Time) error {
	data := []byte{0}
	err := d.bus.ReadRegister(uint8(d.Address), REG_STATUS, data)
	if err != nil {
		return err
	}
	data[0] &^= 1 << OSF
	err = d.bus.WriteRegister(uint8(d.Address), REG_STATUS, data)
	if err != nil {
		return err
	}

	data = make([]uint8, 7)
	data[0] = uint8ToBCD(uint8(dt.Second()))
	data[1] = uint8ToBCD(uint8(dt.Minute()))
	data[2] = uint8ToBCD(uint8(dt.Hour()))

	year := uint8(dt.Year() - 2000)
	centuryFlag := uint8(0)
	if year >= 100 {
		year -= 100
		centuryFlag = 1 << 7
	}

	data[3] = uint8ToBCD(uint8(dt.Weekday()))
	data[4] = uint8ToBCD(uint8(dt.Day()))
	data[5] = uint8ToBCD(uint8(dt.Month()) | centuryFlag)
	data[6] = uint8ToBCD(year)

	err = d.bus.WriteRegister(uint8(d.Address), REG_TIMEDATE, data)
	if err != nil {
		return err
	}

	return nil
}

// ReadTime returns the date and time
func (d *Device) ReadTime() (dt time.Time, err error) {
	data := make([]uint8, 7)
	err = d.bus.ReadRegister(uint8(d.Address), REG_TIMEDATE, data)
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
	err := d.bus.ReadRegister(uint8(d.Address), REG_TEMP, data)
	if err != nil {
		return 0, err
	}
	return int32(data[0])*1000 + int32((data[1]>>6)*25)*10, nil
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
