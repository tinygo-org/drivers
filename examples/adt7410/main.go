package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/adt7410"
)

var (
	i2c    = &machine.I2C0
	sensor = adt7410.New(i2c)
)

func main() {

	i2c.Configure(machine.I2CConfig{Frequency: machine.TWI_FREQ_400KHZ})
	sensor.Configure()

	for {
		temp := sensor.ReadTempF()
		fmt.Printf("temperature: %f\r\n", temp)
		time.Sleep(time.Second)
	}

}
