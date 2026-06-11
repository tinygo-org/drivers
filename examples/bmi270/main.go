package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/bmi270"
)

func main() {
	time.Sleep(5 * time.Second)

	machine.I2C0.Configure(machine.I2CConfig{
		SCL: machine.SCL0_PIN,
		SDA: machine.SDA0_PIN,
	})

	sensor := bmi270.NewI2C(machine.I2C0, bmi270.Address)
	if !sensor.Connected() {
		println("BMI270 not connected")
		return
	}

	cfg := bmi270.DefaultConfig()
	if err := sensor.Configure(cfg); err != nil {
		println("BMI270 configuration failed:", err.Error())
		return
	}

	for {
		time.Sleep(time.Second)

		accelX, accelY, accelZ, err := sensor.ReadAcceleration()
		if err != nil {
			println("Error reading acceleration:", err.Error())
			continue
		}
		fmt.Printf("Acceleration: %.2fg %.2fg %.2fg\n", float32(accelX)/1e6, float32(accelY)/1e6, float32(accelZ)/1e6)

		gyroX, gyroY, gyroZ, err := sensor.ReadRotation()
		if err != nil {
			println("Error reading rotation:", err.Error())
			continue
		}
		fmt.Printf("Rotation: %.2f°/s %.2f°/s %.2f°/s\n", float32(gyroX)/1e6, float32(gyroY)/1e6, float32(gyroZ)/1e6)
	}
}
