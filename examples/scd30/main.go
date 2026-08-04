package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/scd30"
)

var sensor = scd30.New(machine.I2C0)

func main() {
	// The SCD30 requires clock stretching and supports I2C speeds up to 100kHz.
	if err := machine.I2C0.Configure(machine.I2CConfig{Frequency: 50 * machine.KHz}); err != nil {
		println("could not configure I2C:", err.Error())
		return
	}

	if !sensor.Connected() {
		println("SCD30 not detected")
		return
	}
	if err := sensor.Configure(scd30.DefaultConfig); err != nil {
		println("could not configure SCD30:", err.Error())
		return
	}
	if err := sensor.StartContinuousMeasurement(0); err != nil {
		println("could not start SCD30:", err.Error())
		return
	}

	for {
		ready, err := sensor.DataReady()
		if err != nil {
			println("could not read SCD30 status:", err.Error())
			time.Sleep(time.Second)
			continue
		}
		if !ready {
			time.Sleep(250 * time.Millisecond)
			continue
		}

		if err := sensor.Update(drivers.AllMeasurements); err != nil {
			println("could not read SCD30 measurement:", err.Error())
			time.Sleep(time.Second)
			continue
		}

		println("CO2 (ppm):", sensor.CO2())
		println("temperature (mC):", sensor.Temperature())
		println("humidity (0.01%):", sensor.Humidity())
	}
}
