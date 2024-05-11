package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/mcp9808"
)

func main() {

	//tinygo monitor
	time.Sleep(time.Millisecond * 5000)

	//Configure I2C (in this case, I2C0 on RPI Pico), and wire the module accordingly
	machine.I2C0.Configure(machine.I2CConfig{
		SCL: machine.GP1,
		SDA: machine.GP0,
	})

	//Create sensor
	sensor := mcp9808.New(machine.I2C0)
	if !sensor.Connected() {
		println("MCP9808 not found")
		return
	} else {
		println("MCP9808 found")
	}

	time.Sleep(time.Millisecond * 1000)

	//Set resolution
	sensor.SetResolution(mcp9808.Maximum)

	time.Sleep(time.Millisecond * 1000)

	//Read temp.
	temp, err := sensor.ReadTemperature()
	if err != nil {
		println("MCP9808 error reading temperature")
		println(err.Error())
		return
	} else {
		fmt.Printf("Temperature: %.2f \n", temp)
	}
	return
}
