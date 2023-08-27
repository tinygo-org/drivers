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
	i2c.Configure(machine.I2CConfig{Frequency: machine.TWI_FREQ_100KHZ})
	sensor.Configure()

	if sensor.Connected() {
		colorNames := [6]string{"Violet", "Blue", "Green", "Yellow", "Orange", "Red"}
		for {
			// read sensor colors (will read all 6 colors)
			colors := sensor.ReadColors()
			for i, c := range colors {
				println(colorNames[i], ": ", c)
			}

			// read sensor temperature
			temp := sensor.ReadTemp()
			println("Temperature: ", temp)
			time.Sleep(time.Second)
		}
	} else {
		panic("as7262 not connected to I2C")
	}
}
