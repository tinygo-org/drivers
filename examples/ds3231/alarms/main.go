// Connects to an DS3231 I2C Real Time Clock (RTC) and sets both alarms. It then repeatedly checks
// if the alarms are firing and prints out a message if that is the case.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/ds3231"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	rtc := ds3231.New(machine.I2C0)
	rtc.Configure()

	valid := rtc.IsTimeValid()
	if !valid {
		date := time.Date(2019, 12, 05, 20, 34, 12, 0, time.UTC)
		rtc.SetTime(date)
	}

	// Set alarm1 so it triggers when the seconds match 59 => repeats every minute at dd:hh:mm:59
	if err := rtc.SetAlarm1(time.Date(0, 0, 0, 0, 0, 59, 0, time.UTC), ds3231.A1_SECOND); err != nil {
		println("Error while setting Alarm1")
	}
	if err := rtc.SetEnabledAlarm1(true); err != nil {
		println("Error while enabling Alarm1")
	}

	// Set alarm2 so it triggers when the minutes match 35 => repeats every hour at dd:hh:35:ss
	if err := rtc.SetAlarm2(time.Date(0, 0, 0, 0, 35, 0, 0, time.UTC), ds3231.A2_MINUTE); err != nil {
		println("Error while setting Alarm2")
	}
	if err := rtc.SetEnabledAlarm2(true); err != nil {
		println("Error while enabling Alarm2")
	}

	running := rtc.IsRunning()
	if !running {
		err := rtc.SetRunning(true)
		if err != nil {
			println("Error configuring RTC")
		}
	}

	for {
		dt, err := rtc.ReadTime()
		if err != nil {
			println("Error reading date:", err)
			continue
		}

		a1 := rtc.IsAlarm1Fired()
		a2 := rtc.IsAlarm2Fired()

		println(dt.Format(time.DateTime), "A1:", a1, "A2:", a2)

		if a1 {
			if err := rtc.ClearAlarm1(); err != nil {
				println("Error while clearing alarm1")
			}
		}
		if a2 {
			if err := rtc.ClearAlarm2(); err != nil {
				println("Error while clearing alarm2")
			}

		}

		time.Sleep(time.Second * 1)
	}
}
