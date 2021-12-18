package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/lps22hb"
)

func main() {

	// use Nano 33 BLE Sense's internal I2C bus
	machine.I2C1.Configure(machine.I2CConfig{
		SCL:       machine.SCL1_PIN,
		SDA:       machine.SDA1_PIN,
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	sensor := lps22hb.New(machine.I2C1)
	sensor.Configure()

	if !sensor.Connected() {
		println("LPS22HB not connected!")
		return
	}

	for {
		p, _ := sensor.ReadPressure()
		t, _ := sensor.ReadTemperature()
		println("p =", float32(p)/1000.0, "hPa / t =", float32(t)/1000.0, "*C")
		time.Sleep(time.Second)
		// note: the device would power down itself after each query
	}

}
