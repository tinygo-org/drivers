// Connects to a pcf8591 ADC via I2C.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/pcf8591"
)

var (
	i2c = machine.I2C0
)

func main() {
	i2c.Configure(machine.I2CConfig{})
	adc := pcf8591.New(i2c)
	adc.Configure()

	// get "CH0" aka "machine.ADC" interface to channel 0 from ADC.
	p := adc.CH0

	for {
		val := p.Get()
		println(val)
		time.Sleep(50 * time.Millisecond)
	}
}
