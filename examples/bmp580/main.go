package main

import (
	"machine"
	"strconv"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/bmp580"
)

func main() {
	time.Sleep(2 * time.Second)

	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := bmp580.New(machine.I2C0)

	for !sensor.Connected() {
		println("BMP580 not found")
		time.Sleep(time.Second)
	}

	// 16 samples per pressure reading and 2 per temperature reading: noise of
	// about 0.5 Pa, enough for weather.
	err := sensor.Configure(bmp580.Config{Pressure: bmp580.Sampling16X, Temperature: bmp580.Sampling2X})
	if err != nil {
		println("configure error:", err.Error())
	}

	for {
		if err := sensor.Update(drivers.Temperature | drivers.Pressure); err != nil {
			println("update error:", err.Error())
		} else {
			println("Temperature:", strconv.FormatFloat(float64(sensor.Temperature())/1000, 'f', 2, 64), "°C")
			println("Pressure:   ", strconv.FormatFloat(float64(sensor.Pressure())/100_000, 'f', 2, 64), "hPa")
		}
		time.Sleep(time.Second)
	}
}
