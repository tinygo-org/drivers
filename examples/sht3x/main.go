package main

import (
	"fmt"
	"machine"
	"time"

	"github.com/tinygo-org/drivers/sht3x"
)

func main() {
	println("Hello")
	machine.I2C0.Configure(machine.I2CConfig{})
	println("Hello1")
	sensor := sht3x.New(machine.I2C0)
	println("Hello2")

	temp, humidity, _ := sensor.ReadTemperatureHumidity()
	println("Hello3")
	println("Temperature:", temp, "milliºC")
	println("Relative Humidity:", humidity, "%")
	// println(uint16(sensor.ReadTemperature()))
	// println(uint16(sensor.ReadHumidity()))

	// this doesn't work on Arduino
	for {
		temp, humidity, _ := sensor.ReadTemperatureHumidity()
		t := fmt.Sprintf("%.2f", float32(temp)/1000)
		h := fmt.Sprintf("%.2f", float32(humidity)/100)
		println("Temperature:", t, "ºC")
		println("Humidity", h, "%")
		time.Sleep(2 * time.Second)
	}
}
