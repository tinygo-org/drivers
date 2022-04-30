package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/scd4x"
)

var (
	i2c    = machine.I2C0
	sensor = scd4x.New(i2c)
)

func main() {
	time.Sleep(1500 * time.Millisecond)

	i2c.Configure(machine.I2CConfig{})
	if err := sensor.Configure(); err != nil {
		println(err)
	}

	time.Sleep(1500 * time.Millisecond)

	if err := sensor.StartPeriodicMeasurement(); err != nil {
		println(err)
	}

	time.Sleep(1500 * time.Millisecond)

	for {
		co2, err := sensor.ReadCO2()
		if err != nil {
			println(err)
		}
		println("CO2", co2)
		time.Sleep(time.Second)
	}
}
