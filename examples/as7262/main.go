// tinygo flash -target arduino examples/as7262/main.go && tinygo monitor -baudrate 9600

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
	sensor.Configure(true, 64, 17.857, 2)
	//sensor.ConfigureLed(12.5, true, 8, true)

	println("Starting ...")

	for {
		println("Value: ", sensor.Temperature())
		time.Sleep(time.Millisecond * 800)
	}
}
