package main

import (
	"machine"
	"time"
	"tinygo.org/x/drivers/bme280"
	"tinygo.org/x/drivers/pca9548a"
)

func main() {
	time.Sleep(5 * time.Second)
	err := machine.I2C0.Configure(machine.I2CConfig{})
	if err != nil {
		panic(err.Error())
	}
	mux := pca9548a.New(machine.I2C0, pca9548a.Address)
	if !mux.IsConnected() {
		println("NO DEVICE DETECTED")
		return
	}

	port := mux.GetPortState()
	println("GET PORT", port)
	mux.DisablePort()
	mux.SetPort(0)
	port = mux.GetPortState()
	println("GET PORT", port)
	mux.SetPort(1)
	port = mux.GetPortState()
	println("GET PORT", port)

	tmpSensors := make([]bme280.Device, 2)

	for i := uint8(0); i < 2; i++ {
		mux.SetPort(i)
		time.Sleep(10 * time.Millisecond)
		tmpSensors[i] = bme280.New(machine.I2C0)
		tmpSensors[i].Configure()

		connected := tmpSensors[i].Connected()
		if !connected {
			println("\nBME280 Sensor not detected\n", i)
		} else {
			println("\nBME280 Sensor detected\n", i)
		}
	}
	time.Sleep(10000 * time.Millisecond)

	for {
		for i := uint8(0); i < 2; i++ {
			mux.SetPort(i)
			t, err := tmpSensors[i].ReadTemperature()
			if err != nil {
				println(i, "Error reading temperature")
			}
			println(i, "temperature", t)

			p, err := tmpSensors[i].ReadPressure()
			if err != nil {
				println("Error reading pressure")
			}
			println(i, "pressure", p)

		}
		time.Sleep(40 * time.Millisecond)
	}
}
