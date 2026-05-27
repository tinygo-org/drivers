package main

import (
	"machine"
	"strconv"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/bh1745"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	sensor := bh1745.New(machine.I2C0)

	if !sensor.Connected() {
		println("BH1745 not detected")
		return
	}
	println("BH1745 detected")

	if err := sensor.Configure(); err != nil {
		println("configure error:", err.Error())
		return
	}

	for {
		if err := sensor.Update(drivers.Luminosity); err != nil {
			println("read error:", err.Error())
			time.Sleep(500 * time.Millisecond)
			continue
		}

		println("R:", strconv.Itoa(int(sensor.R())),
			" G:", strconv.Itoa(int(sensor.G())),
			" B:", strconv.Itoa(int(sensor.B())),
			" C:", strconv.Itoa(int(sensor.C())))
		println("Lux:", strconv.Itoa(int(sensor.Luminosity())))
		println("CCT:", strconv.Itoa(int(sensor.ColorTemperature())), "K")
		println("---")

		time.Sleep(500 * time.Millisecond)
	}
}
