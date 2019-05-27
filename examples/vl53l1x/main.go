package main

import (
	"machine"

	"time"

	"tinygo.org/x/drivers/vl53l1x"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400000,
	})
	sensor := vl53l1x.New(machine.I2C0)
	connected := sensor.Connected()
	if !connected {
		println("VL53L1X device not found")
		return
	}
	println("VL53L1X device found")
	sensor.Configure(true)
	sensor.SetMeasurementTimingBudget(50000)
	sensor.StartContinuous(50)
	for {
		sensor.Read(true)
		println("Distance (mm):", sensor.Distance())
		println("Status:", sensor.Status())
		println("Peak signal rate (cps):", sensor.SignalRate())
		println("Ambient rate (cps):", sensor.AmbientRate())
		println("---")
		time.Sleep(100 * time.Millisecond)
	}
}
