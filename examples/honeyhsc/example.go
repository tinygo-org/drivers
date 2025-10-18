package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/honeyhsc"
)

// Data taken from https://github.com/rodan/honeywell_hsc_ssc_i2c/blob/master/hsc_ssc_i2c.cpp
// these defaults are valid for the HSCMRNN030PA2A3 chip
const (
	i2cAddress = 0x28
	// 10%
	outputMinimum = 0x666
	// 90% of 2^14 - 1
	outputMax = 0x399A
	// min is 0 for sensors that give absolute values
	pressureMin = 0
	// 30psi (and we want results in millipascals)
	// pressureMax = 206842.7
	pressureMax = 206843 * 1000
)

func main() {
	bus := machine.I2C0
	err := bus.Configure(machine.I2CConfig{
		Frequency: 400_000, // 100kHz minimum and 400kHz I2C maximum clock. 50 to 800 for SPI.
		SDA:       machine.I2C0_SDA_PIN,
		SCL:       machine.I2C0_SCL_PIN,
	})
	if err != nil {
		panic(err.Error())
	}
	sensor := honeyhsc.NewDevI2C(bus, i2cAddress, outputMinimum, outputMax, pressureMin, pressureMax)
	for {
		time.Sleep(time.Second)
		const measuremask = drivers.Pressure | drivers.Temperature
		err := sensor.Update(measuremask)
		if err != nil {
			println("error updating measurements:", err.Error())
			continue
		}
		P := sensor.Pressure()
		T := sensor.Temperature()
		println("pressure:", P, "temperature:", T)
	}
}
