package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/ndir"
)

var (
	ndirBus = machine.I2C0
)

func main() {
	err := ndirBus.Configure(machine.I2CConfig{
		Frequency: 100_000,
	})
	if err != nil {
		panic("i2c config fail:" + err.Error())
	}
	// Set the address based on how the resistors are soldered.
	// True means the left and middle pads are joined.
	ndirAddr := ndir.Addr(true, false)
	dev := ndir.NewDevI2C(ndirBus, ndirAddr)
	err = dev.Init()
	if err != nil {
		panic("ndir init fail:" + err.Error())
	}
	// Datasheet tells us to wait 12 seconds before reading from the sensor.
	time.Sleep(12 * time.Second)
	for {
		time.Sleep(time.Second)
		err := dev.Update(drivers.AllMeasurements)
		if err != nil {
			println(err.Error())
			continue
		}
		println("PPM:", dev.PPMCO2())
	}
}
