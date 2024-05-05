package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/mcp9808"
)

func main() {

	time.Sleep(time.Millisecond * 5000)

	err := machine.I2C0.Configure(machine.I2CConfig{
		SCL: machine.GP1,
		SDA: machine.GP0,
	})

	if err != nil {
		fmt.Println("i2c error")
		fmt.Println(err.Error())
	}

	sensor := mcp9808.New(machine.I2C0)
	fmt.Println("Device sensor created")

	for {
		if !sensor.Connected() {
			println("not connected!")
			return
		}

		temp, err := sensor.Temperature()
		if err != nil {
			println("temp read error")
			println(err.Error())
		} else {
			println(temp)
		}
		time.Sleep(time.Millisecond * 1000)
	}
}
