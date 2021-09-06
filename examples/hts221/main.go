package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/hts221"
)

func main() {

	machine.I2C1.Configure(machine.I2CConfig{
		SCL:       machine.P0_15, // SCL1 on Nano 33 BLE Sense
		SDA:       machine.P0_14, // SDA1 on Nano 33 BLE Sense
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	sensor := hts221.New(machine.I2C1, hts221.ON_NANO_33_BLE)
	// for normal sensor module, use
	// sensor := hts221.New(machine.I2C1, hts221.STANDARD)

	if !sensor.Connected() {
		println("HTS221 not connected!")
		return
	}

	sensor.Configure() // power on and calibrate

	for {

		h, _ := sensor.ReadHumidity()
		t, _ := sensor.ReadTemperature()
		println("h =", float32(h)/100.0, "% / t =", float32(t)/1000.0, "*C")
		time.Sleep(time.Second)

	}

}
