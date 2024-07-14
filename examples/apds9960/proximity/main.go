package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/apds9960"
)

func main() {
	// Sleep to catch any errors through the serial monitor.
	time.Sleep(1000 * time.Millisecond)
	bus := machine.I2C0
	// use Nano 33 BLE Sense's internal I2C bus
	err := bus.Configure(machine.I2CConfig{
		SCL:       machine.GP1,
		SDA:       machine.GP0,
		Frequency: 400 * machine.KHz,
	})
	if err != nil {
		panic(err.Error())
	}

	sensor := apds9960.New(bus)

	// use default settings
	sensor.Configure(apds9960.Configuration{})

	if !sensor.Connected() {
		println("APDS-9960 not connected!")
		println("err:", sensor.Err())
		return
	}
	println("APDS connected!")
	err = sensor.EnableProximity() // enable proximity engine
	if err != nil {
		panic(err.Error())
	}
	for {
		if sensor.ProximityAvailable() {
			p := sensor.ReadProximity()
			println("Proximity:", p)
		}
		if err := sensor.Err(); err != nil {
			println(err.Error())
		}
		time.Sleep(time.Millisecond * 100)
	}
}
