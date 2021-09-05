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

	sensor := apds9960.New(machine.I2C1, apds9960.ON_NANO_33_BLE)
	// for normal sensor module, use
	// sensor := apds9960.New(machine.I2C1, apds9960.STANDARD)

	if !sensor.Connected() {
		println("APDS-9960 not connected!")
		return
	}

	// use default settings
	sensor.Configure(apds9960.Configuration{})

	// enable proximity engine
	sensor.EnableProximity()

	for {

		if sensor.ProximityAvailable() {
			p := sensor.ReadProximity()
			println("Proximity:", p)
		}
		time.Sleep(time.Millisecond * 100)
	}

}
