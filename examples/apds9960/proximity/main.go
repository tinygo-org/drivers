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

	// use default settings
	sensor.Configure(apds9960.Configuration{})

	if !sensor.Connected() {
		println("APDS-9960 not connected!")
		return
	}

	sensor.EnableProximity() // enable proximity engine

	for {

		if sensor.ProximityAvailable() {
			p := sensor.ReadProximity()
			println("Proximity:", p)
		}
		time.Sleep(time.Millisecond * 100)
	}

}
