// Connects to an DS3231 I2C Real Time Clock (RTC) and sets both alarms. It then repeatedly checks
// if the alarms are firing and prints out a message if that is the case.
package main

import (
	"machine"
	"time"

	"fmt"

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
		fmt.Println("Error while setting Alarm1")
	}
	if err := rtc.EnableAlarm1(); err != nil {
		fmt.Println("Error while enabling Alarm1")
	}

	// Set alarm2 so it triggers when the minutes match 59 => repeats every hour at dd:hh:59:ss
	if err := rtc.SetAlarm2(time.Date(0, 0, 0, 0, 59, 0, 0, time.UTC), ds3231.A2_MINUTE); err != nil {
		fmt.Println("Error while setting Alarm2")
	}
	if err := rtc.EnableAlarm2(); err != nil {
		fmt.Println("Error while enabling Alarm2")
	}

	running := rtc.IsRunning()
	if !running {
		err := rtc.SetRunning(true)
		if err != nil {
			fmt.Println("Error configuring RTC")
		}
	}

	for {
		dt, err := rtc.ReadTime()
		if err != nil {
			fmt.Println("Error reading date:", err)
			continue
		}

		if rtc.IsAlarm1Fired() {
			fmt.Printf(
				"Alarm1 fired at %d/%s/%02d %02d:%02d:%02d \r\n",
				dt.Year(),
				dt.Month(),
				dt.Day(),
				dt.Hour(),
				dt.Minute(),
				dt.Second(),
			)
			if err := rtc.ClearAlarm1(); err != nil {
				fmt.Println("Error while clearing alarm1")
			}
		}

		if rtc.IsAlarm2Fired() {
			fmt.Printf(
				"Alarm2 fired at %d/%s/%02d %02d:%02d:%02d \r\n",
				dt.Year(),
				dt.Month(),
				dt.Day(),
				dt.Hour(),
				dt.Minute(),
				dt.Second(),
			)
			if err := rtc.ClearAlarm2(); err != nil {
				fmt.Println("Error while clearing alarm2")
			}
		}

		time.Sleep(time.Second * 1)
	}
}
