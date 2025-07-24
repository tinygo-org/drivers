// Connects to an DS3231 I2C Real Time Clock (RTC).
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
		} else {
			fmt.Printf("Date: %d/%s/%02d %02d:%02d:%02d \r\n", dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute(), dt.Second())
		}
		temp, _ := rtc.ReadTemperature()
		fmt.Printf("Temperature: %.2f °C \r\n", float32(temp)/1000)

		time.Sleep(time.Second * 1)
	}
}
