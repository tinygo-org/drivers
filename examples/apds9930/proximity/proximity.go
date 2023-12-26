package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/apds9930"
)

func main() {
	// Sleep to catch any errors through the serial monitor.
	time.Sleep(1000 * time.Millisecond)
	bus := machine.I2C0
	// use Nano 33 BLE Sense's internal I2C bus
	err := bus.Configure(machine.I2CConfig{
		SCL:       machine.GP1,
		SDA:       machine.GP0,
		Frequency: 100 * machine.KHz,
	})
	if err != nil {
		panic(err.Error())
	}
	sensor := apds9930.New(bus, 0x39)

	err = sensor.Init()
	if err != nil {
		panic(err)
	}
	err = sensor.EnableProximity()
	if err != nil {
		panic(err)
	}
	println("proximity enabled!")
	for {
		time.Sleep(50 * time.Millisecond)
		prox := sensor.ReadProximity()
		println("proximity:", prox)
	}
}
