package main

import (
	"fmt"
	"machine"
	"time"
	"tinygo.org/x/drivers/bmi160"
)

func main() {
	time.Sleep(5 * time.Second)

	machine.SPI0.Configure(machine.SPIConfig{})
	sensor := bmi160.NewSPI(machine.A5, machine.SPI0)
	sensor.Configure()

	if !sensor.Connected() {
		println("BMI160 not connected")
		return
	}

	for {
		time.Sleep(time.Second)

		t, err := sensor.ReadTemperature()
		if err != nil {
			println("Error reading temperature", err)
			continue
		}
		fmt.Printf("Temperature: %.2f °C\n", float32(t)/1000)

		accelX, accelY, accelZ, err := sensor.ReadAcceleration()
		if err != nil {
			println("Error reading acceleration", err)
			continue
		}
		fmt.Printf("Acceleration: %.2fg %.2fg %.2fg\n", float32(accelX)/1e6, float32(accelY)/1e6, float32(accelZ)/1e6)

		gyroX, gyroY, gyroZ, err := sensor.ReadRotation()
		if err != nil {
			println("Error reading rotation", err)
			continue
		}
		fmt.Printf("Rotation: %.2f°/s %.2f°/s %.2f°/s\n", float32(gyroX)/1e6, float32(gyroY)/1e6, float32(gyroZ)/1e6)
	}
}
