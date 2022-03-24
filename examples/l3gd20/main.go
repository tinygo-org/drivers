package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/l3gd20"
)

func main() {
	const (
		// Default address on most breakout boards.
		pcaAddr = 0x40
	)

	bus := machine.I2C0

	err := bus.Configure(machine.I2CConfig{})
	if err != nil {
		panic(err.Error())
	}

	gyro := l3gd20.NewI2C(bus, 105)
	err = gyro.Configure(l3gd20.Config{Range: l3gd20.Range_250})
	if err != nil {
		println(err.Error())
	}

	var x, y, z int32
	for {
		err = gyro.Update()
		if err != nil {
			println(err.Error())
		}
		x, y, z = gyro.AngularVelocity()
		println(x, y, z)
		time.Sleep(500 * time.Millisecond)
	}
}
