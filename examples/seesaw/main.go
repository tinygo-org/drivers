package main

import (
	"machine"
	"strconv"
	"time"
	"tinygo.org/x/drivers/seesaw"
)

const readDelay = time.Microsecond * 3000

// example reading soil moisture with an Adafruit capacitive soil-sensor (4026) powered by a seesaw
// https://learn.adafruit.com/adafruit-stemma-soil-sensor-i2c-capacitive-moisture-sensor/overview
func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	dev := seesaw.New(machine.I2C0)

	dev.Address = 0x36

	var buf [2]byte
	err := dev.Read(seesaw.ModuleTouchBase, seesaw.FunctionTouchChannelOffset, buf[:], readDelay)
	if err != nil {
		panic(err)
	}
	moisture := uint16(buf[0])<<8 | uint16(buf[1])

	println("soil moisture: " + strconv.Itoa(int(moisture)))
}
