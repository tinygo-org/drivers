package main

import (
	"machine"

	"github.com/tinygo-org/drivers/sht3x"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := sht3x.New(machine.I2C0)

	temp, humidity, _ := sensor.ReadTemperatureHumidity()
	println(uint16(temp))
	println(uint16(humidity))
	// println(uint16(sensor.ReadTemperature()))
	// println(uint16(sensor.ReadHumidity()))

	// this doesn't work on Arduino
	// for {
	// 	sensor.Read()
	// 	t := fmt.Sprintf("%.2f", temp)
	// 	h := fmt.Sprintf("%.2f", humidity)
	// 	println("Temperature:", t, "ºC")
	// 	println("Humidity", h, "%")
	// 	time.Sleep(2 * time.Second)
	// }
}
