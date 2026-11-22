package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/ltr390"
)

func main() {
	time.Sleep(2 * time.Second)

	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := ltr390.New(machine.I2C0)

	for !sensor.Connected() {
		println("LTR390 not found")
		time.Sleep(time.Second)
	}

	if err := sensor.Configure(ltr390.Config{}); err != nil {
		println("configure error:", err.Error())
	}

	for {
		lux, err := sensor.ReadIlluminance()
		if err != nil {
			println("illuminance error:", err.Error())
		}
		uvi, err := sensor.ReadUVIndex()
		if err != nil {
			println("uv index error:", err.Error())
		}
		fmt.Printf("Illuminance: %d.%03d lx\n", lux/1000, lux%1000)
		fmt.Printf("UV index:    %d.%03d\n", uvi/1000, uvi%1000)
		time.Sleep(time.Second)
	}
}
