package main

import (
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

	for {
		rtc.SetOscillatorFrequency(pcf8563.RTC_COT_1HZ)
		time.Sleep(3 * time.Second)
		rtc.SetOscillatorFrequency(pcf8563.RTC_COT_32HZ)
		time.Sleep(3 * time.Second)
		rtc.SetOscillatorFrequency(pcf8563.RTC_COT_1KHZ)
		time.Sleep(3 * time.Second)
		rtc.SetOscillatorFrequency(pcf8563.RTC_COT_32KHZ)
		time.Sleep(3 * time.Second)
		rtc.SetOscillatorFrequency(pcf8563.RTC_COT_DISABLE)
		time.Sleep(3 * time.Second)
	}
}
