package main

import (
	"machine"
	"time"
	"tinygo.org/x/drivers/as7262"
)

var (
	i2c    = machine.I2C0
	sensor = as7262.New(i2c)
)

func main() {
	i2c.Configure(machine.I2CConfig{Frequency: machine.KHz * 100})
	sensor.Configure(true, 64, 30, 2)

	println("Starting ...")
	for {
		println("need more context")
		time.Sleep(time.Millisecond * 800)
	}
}
