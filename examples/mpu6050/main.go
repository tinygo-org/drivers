// Connects to an MPU6050 I2C accelerometer/gyroscope.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/mpu6050"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	mpuDevice := mpu6050.New(machine.I2C0, mpu6050.DefaultAddress)

	// Configure the device with default configuration.
	err := mpuDevice.Configure(mpu6050.Config{})
	if err != nil {
		panic(err.Error())
	}
	for {
		time.Sleep(time.Millisecond * 100)
		err := mpuDevice.Update(drivers.AllMeasurements)
		if err != nil {
			println("error reading from mpu6050:", err.Error())
			continue
		}
		print("acceleration: ")
		println(mpuDevice.Acceleration())
		print("angular velocity:")
		println(mpuDevice.AngularVelocity())
		print("temperature celsius:")
		println(mpuDevice.Temperature() / 1000)
	}
}
