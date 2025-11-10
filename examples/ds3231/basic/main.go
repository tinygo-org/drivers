// Connects to an DS3231 I2C Real Time Clock (RTC).
package main

import (
	"machine"
	"strconv"
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
		} else {
			println(dt.Format(time.DateTime))
		}
		temp, _ := rtc.ReadTemperature()
		println("Temperature:", strconv.FormatFloat(float64(temp)/1000, 'f', -1, 32), "°C")

		time.Sleep(time.Second * 1)
	}
}
