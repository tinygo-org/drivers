package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/pcf8563"
)

var (
	i2c = machine.I2C0
	rtc = pcf8563.New(i2c)
)

func main() {
	i2c.Configure(machine.I2CConfig{Frequency: machine.TWI_FREQ_400KHZ})
	rtc.Reset()

	rtc.SetTime(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))

	prev := -1

	for {
		for {
			t, _ := rtc.ReadTime()
			if prev != t.Second() {
				fmt.Printf("%s\r\n", t.String())
				prev = t.Second()
				break
			}
			time.Sleep(time.Millisecond * 100)
		}
	}

}
