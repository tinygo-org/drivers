package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/apds9960"
)

func main() {

	// use Nano 33 BLE Sense's internal I2C bus
	machine.I2C1.Configure(machine.I2CConfig{
		SCL:       machine.SCL1_PIN,
		SDA:       machine.SDA1_PIN,
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	sensor := apds9960.New(machine.I2C1)

	sensor.Configure(apds9960.Configuration{}) // use default settings

	if !sensor.Connected() {
		println("APDS-9960 not connected!")
		return
	}

	sensor.EnableGesture() // enable gesture engine

	for {

		// wave your hand (not too slow) about 10 cm above the sensor
		if sensor.GestureAvailable() {

			gesture := sensor.ReadGesture()
			print("Detected gesture: ")
			switch gesture {
			case apds9960.GESTURE_UP: // the nRF52 chip is "up"
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
