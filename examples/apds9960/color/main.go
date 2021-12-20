package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/apds9960"
)

func main() {

	// use Nano 33 BLE Sense's internal I2C bus
	machine.I2C1.Configure(machine.I2CConfig{
		SCL:       machine.SCL1_PIN,
		SDA:       machine.SDA1_PIN,
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	sensor := apds9960.New(machine.I2C1)

	sensor.Configure(apds9960.Configuration{}) // use default settings

	if !sensor.Connected() {
		println("APDS-9960 not connected!")
		return
	}

	sensor.EnableColor() // enable color engine

	for {

		if sensor.ColorAvailable() {
			r, g, b, c := sensor.ReadColor()
			println("Red =", r, "\tGreen =", g, "\tBlue =", b, "\tClear =", c)
		}
		time.Sleep(time.Millisecond * 100)
	}

}
