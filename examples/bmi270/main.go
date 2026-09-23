package main

import (
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
		println("acc (mg):", accelX/1000, accelY/1000, accelZ/1000)

		gyroX, gyroY, gyroZ, err := sensor.ReadRotation()
		if err != nil {
			println("Error reading rotation:", err.Error())
			continue
		}
		println("gyr (mdps):", gyroX/1000, gyroY/1000, gyroZ/1000)
	}
}
