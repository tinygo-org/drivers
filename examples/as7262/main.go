package main

import (
	"fmt"
	"machine"
	"time"
	"tinygo.org/x/drivers/as7262"
)

var (
	i2c    = machine.I2C0
	sensor = as7262.New(i2c)
)

func main() {

	i2c.Configure(machine.I2CConfig{Frequency: machine.TWI_FREQ_100KHZ})
	sensor.Configure()

	if sensor.Connected() {
		for {
			// read sensor colors (will read all 6 colors)
			colors := sensor.ReadColors()
			fmt.Printf("Colors: %v", colors)
			time.Sleep(time.Second)

			// read sensor temperature
			temp := sensor.ReadTemp()
			fmt.Printf("Temperature: %d", temp)
			time.Sleep(time.Second)
		}
	} else {
		panic("as7262 not connected to I2C")
	}
}
