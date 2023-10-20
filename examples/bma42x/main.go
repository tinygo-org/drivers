package main

// Smoke test for the BMA421/BMA425 sensors.
// Warning: this code has _not been tested_. It's only here as a smoke test.

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/bma42x"
)

func main() {
	time.Sleep(5 * time.Second)

	i2cBus := machine.I2C1
	i2cBus.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
		SDA:       machine.SDA_PIN,
		SCL:       machine.SCL_PIN,
	})

	sensor := bma42x.NewI2C(i2cBus, bma42x.Address)
	err := sensor.Configure(bma42x.Config{
		Device:   bma42x.DeviceBMA421 | bma42x.DeviceBMA425,
		Features: bma42x.FeatureStepCounting,
	})
	if err != nil {
		println("could not configure BMA421/BMA425:", err)
		return
	}

	if !sensor.Connected() {
		println("BMA42x not connected")
		return
	}

	for {
		time.Sleep(time.Second)

		err := sensor.Update(drivers.Acceleration | drivers.Temperature)
		if err != nil {
			println("Error reading sensor", err)
			continue
		}

		fmt.Printf("Temperature: %.2f °C\n", float32(sensor.Temperature())/1000)

		accelX, accelY, accelZ := sensor.Acceleration()
		fmt.Printf("Acceleration: %.2fg %.2fg %.2fg\n", float32(accelX)/1e6, float32(accelY)/1e6, float32(accelZ)/1e6)
	}
}
