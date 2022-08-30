package main

import (
	"machine"

	"tinygo.org/x/drivers/vl6180x"

	"time"
)

func main() {
	time.Sleep(3 * time.Second)
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400000,
	})
	sensor := vl6180x.New(machine.I2C0)
	connected := sensor.Connected()
	if !connected {
		println("VL6180X device not found")
		return
	}
	println("VL6180X device found")
	sensor.Configure(true)
	var value uint16
	var status uint8
	for {
		value = sensor.Read()
		status = sensor.ReadStatus()
		println("Distance (mm):", value)
		println("Status:", status)
		println("---")
		time.Sleep(100 * time.Millisecond)
	}
}
