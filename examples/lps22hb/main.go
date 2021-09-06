package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/lps22hb"
)

func main() {

	machine.I2C1.Configure(machine.I2CConfig{
		SCL:       machine.P0_15, // SCL1 on Nano 33 BLE Sense
		SDA:       machine.P0_14, // SDA1 on Nano 33 BLE Sense
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	sensor := lps22hb.New(machine.I2C1)

	if !sensor.Connected() {
		println("LPS22HB not connected!")
		return
	}

	sensor.Configure()

	for {

		p, _ := sensor.ReadPressure()
		t, _ := sensor.ReadTemperature()
		println("p =", float32(p)/1000.0, "hPa / t =", float32(t)/1000.0, "*C")
		time.Sleep(time.Second)
		// note: the device would power down itself after each query

	}

}
