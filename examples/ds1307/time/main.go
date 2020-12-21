package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/ds1307"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})
	rtc := ds1307.New(machine.I2C0)
	rtc.SetTime(time.Date(2019, 5, 15, 20, 34, 12, 0, time.UTC))

	for {
		t, err := rtc.ReadTime()
		if err != nil {
			println("Error reading date:", err)
			break
		}
		println(t.Hour(), ":", t.Minute(), ":", t.Second(), " ", t.Day(), "/", t.Month(), "/", t.Year())

	}

}
