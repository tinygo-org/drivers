package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/apds9960"
)

func main() {

	machine.I2C1.Configure(machine.I2CConfig{
		SCL:       machine.P0_15, // SCL1 on Nano 33 BLE Sense
		SDA:       machine.P0_14, // SDA1 on Nano 33 BLE Sense
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	sensor := apds9960.New(machine.I2C1)

	if !sensor.Connected() {
		println("APDS-9960 not connected!")
		return
	}

	sensor.Configure(apds9960.Configuration{}) // use default settings

	sensor.EnableGesture() // enable gesture engine

	for {

		// wave your hand (not too slow) about 10 cm above the sensor
		if sensor.GestureAvailable() {

			gesture := sensor.ReadGesture()
			print("Detected gesture: ")
			switch gesture {
			case apds9960.GESTURE_UP:
				println("Up")
			case apds9960.GESTURE_DOWN:
				println("Down")
			case apds9960.GESTURE_LEFT:
				println("Left")
			case apds9960.GESTURE_RIGHT:
				println("Right")
			}
		}
		// note: the delay shouldn't be too long, otherwise new gesture data might be lost
		time.Sleep(time.Millisecond * 250)
	}

}
