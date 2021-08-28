package main

import (
	"machine"
	"time"

	"MyProject/lps22hb"
)

func main() {

	machine.I2C1.Configure(machine.I2CConfig{
		SCL:       machine.P0_15, // SCL1 on Nano 33 BLE Sense
		SDA:       machine.P0_14, // SDA1 on Nano 33 BLE Sense
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	sensor := lps22hb.New(machine.I2C1, lps22hb.ON_NANO_33_BLE)
	// for normal sensor module, use
	// sensor := lps22hb.New(machine.I2C1, lps22hb.STANDARD)

	if !sensor.Connected() {
		println("LPS22HB not connected!")
		return
	}

	sensor.Configure()

	for {

		p, _ := sensor.ReadPressure()
		t, _ := sensor.ReadTemperature()
		println("p =", float32(p)/1000.0, "/ t =", float32(t)/1000.0)
		time.Sleep(time.Second)
		// the device would power down itself after each query

	}

}
