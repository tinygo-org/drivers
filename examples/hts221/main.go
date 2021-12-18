package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/hts221"
)

func main() {

	// use Nano 33 BLE Sense's internal I2C bus
	machine.I2C1.Configure(machine.I2CConfig{
		SCL:       machine.SCL1_PIN,
		SDA:       machine.SDA1_PIN,
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	sensor := hts221.New(machine.I2C1)

	sensor.Configure() // power on and calibrate

	if !sensor.Connected() {
		println("HTS221 not connected!")
		return
	}

	for {
		h, _ := sensor.ReadHumidity()
		t, _ := sensor.ReadTemperature()
		println("h =", float32(h)/100.0, "% / t =", float32(t)/1000.0, "*C")
		time.Sleep(time.Second)
	}

}
