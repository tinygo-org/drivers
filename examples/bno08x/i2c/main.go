// Package main provides a basic example of using the BNO08x driver
// to read rotation vector (quaternion) data from the sensor.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/bno08x"
)

func main() {
	time.Sleep(2 * time.Second) // Wait for sensor to power up
	// Initialize I2C bus
	i2c := machine.I2C0
	err := i2c.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
	})
	if err != nil {
		println("Failed to configure I2C:", err.Error())
		return
	}

	println("Initializing BNO08x sensor...")

	// Create and configure sensor using I2C
	sensor := bno08x.NewI2C(i2c)
	err = sensor.Configure(bno08x.Config{})
	if err != nil {
		println("Failed to configure sensor:", err.Error())
		return
	}

	println("Sensor initialized successfully")

	// Enable Game Rotation Vector reports at 100Hz (10000 microseconds = 10ms interval)
	// Using Game Rotation Vector (0x08) to match the working channel_debug test
	err = sensor.EnableReport(bno08x.SensorGameRotationVector, 10000)
	if err != nil {
		println("Failed to enable game rotation vector:", err.Error())
		return
	}

	println("Reading rotation vectors...")
	println("Format: Real I J K Accuracy")

	// Add a delay after enabling reports (Arduino does this)
	time.Sleep(100 * time.Millisecond)

	// Main loop - read and display quaternion data
	for {
		event, ok := sensor.GetSensorEvent()
		if ok && (event.ID() == bno08x.SensorRotationVector || event.ID() == bno08x.SensorGameRotationVector) {
			q := event.Quaternion()
			if event.ID() == bno08x.SensorRotationVector {
				println(q.Real, q.I, q.J, q.K, event.QuaternionAccuracy())
			} else {
				// GameRotationVector doesn't have accuracy
				println(q.Real, q.I, q.J, q.K)
			}
		}

		// Arduino uses 10ms delay in loop
		time.Sleep(10 * time.Millisecond)
	}
}
