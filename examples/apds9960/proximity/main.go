package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/apds9960"
)

func main() {

	machine.I2C1.Configure(machine.I2CConfig{
		SCL:       machine.P0_15, // SCL1 on Nano 33 BLE Sense
		SDA:       machine.P0_14, // SDA1 on Nano 33 BLE Sense
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	sensor := apds9960.New(machine.I2C1)

	if !sensor.Connected() {
		println("APDS-9960 not connected!")
		return
	}

	sensor.Configure(apds9960.Configuration{}) // use default settings

	sensor.EnableProximity() // enable proximity engine

	for {

		if sensor.ProximityAvailable() {
			p := sensor.ReadProximity()
			println("Proximity:", p)
		}
		time.Sleep(time.Millisecond * 100)
	}

}
