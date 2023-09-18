// Package pcf8523 implements a driver for the PCF8523 CMOS Real-Time Clock (RTC)
//
// Datasheet: https://www.nxp.com/docs/en/data-sheet/PCF8523.pdf
package pcf8523

import (
	"time"
	"tinygo.org/x/drivers"
)

const DefaultAddress = 0x68

// constants for all internal registers
const (
	rControl1               = 0x00 // Control_1
	rControl2               = 0x01 // Control_2
	rControl3               = 0x02 // Control_3
	rSeconds                = 0x03 // Seconds
	rMinutes                = 0x04 // Minutes
	rHours                  = 0x05 // Hours
	rDays                   = 0x06 // Days
	rWeekdays               = 0x07 // Weekdays
	rMonths                 = 0x08 // Months
	rYears                  = 0x09 // Years
	rMinuteAlarm            = 0x0A // Minute_alarm
	rHourAlarm              = 0x0B // Hour_alarm
	rDayAlarm               = 0x0C // Day_alarm
	rWeekdayAlarm           = 0x0D // Weekday_alarm
	rOffset                 = 0x0E // Offset
	rTimerClkoutControl     = 0x0F // Tmr_CLKOUT_ctrl
	rTimerAFrequencyControl = 0x10 // Tmr_A_freq_ctrl
	rTimerARegister         = 0x11 // Tmr_A_reg
	rTimerBFrequencyControl = 0x12 // Tmr_B_freq_ctrl
	rTimerBRegister         = 0x13 // Tmr_B_reg
)

// datasheet 8.5 Power management functions, table 11
type PowerManagement byte

const (
	PowerManagement_SwitchOver_ModeStandard_LowDetection PowerManagement = 0b000
	PowerManagement_SwitchOver_ModeDirect_LowDetection   PowerManagement = 0b001
	PowerManagement_VddOnly_LowDetection                 PowerManagement = 0b010
	PowerManagement_SwitchOver_ModeStandard              PowerManagement = 0b100
	PowerManagement_SwitchOver_ModeDirect                PowerManagement = 0b101
	PowerManagement_VddOnly                              PowerManagement = 0b101
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
