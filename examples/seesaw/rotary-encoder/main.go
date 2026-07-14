package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/seesaw"
)

// example reading the position of a rotary encoder (4991) powered by a seesaw
// https://learn.adafruit.com/adafruit-i2c-qt-rotary-encoder/arduino
func main() {
	// This assumes you are using an Adafruit QT Py RP2040 for its Stemma QT connector
	// https://www.adafruit.com/product/4900
	i2c := machine.I2C1
	i2c.Configure(machine.I2CConfig{
		SCL: machine.I2C1_QT_SCL_PIN,
		SDA: machine.I2C1_QT_SDA_PIN,
	})

	dev := seesaw.New(i2c)
	dev.Address = 0x36

	for {
		time.Sleep(time.Second)

		pos, err := dev.GetEncoderPosition(0, false)
		if err != nil {
			println(err)
			continue
		}

		println(pos)
	}
}
