package main

import (
	"machine"
	"strconv"
	"time"

	"tinygo.org/x/drivers/bmp388"
)

func main() {

	machine.I2C0.Configure(machine.I2CConfig{})

	sensor := bmp388.New(machine.I2C0)
	if !sensor.Connected() {
		println("Uh oh, BMP388 not detected")
		return
	}

	// The accuracy of the sensor can be increased, at the cost of a slower output rate. Table 9 in Section 3.5 of the
	// datasheet has recommended settings for common use cases. If increasing the sampling rate, the output data rate
	// (ODR) will likely have to be decreased. Configure() will return an error if there's a problem with the
	// configuration settings - keep decreasing the ODR and cycling the power to the sensor until it is happy.
	err := sensor.Configure(bmp388.Config{
		Pressure:    bmp388.Sampling8X,
		Temperature: bmp388.Sampling2X,
		ODR:         bmp388.Odr25,
		IIR:         bmp388.Coeff0,
		Mode:        bmp388.Normal,
	})

	// This is also fine
	// err := sensor.Configure(bmp388.BMP388Config{})

	if err != nil {
		println(err)
	}

	for {
		temp, err := sensor.ReadTemperature() // returns the temperature in centicelsius
		press, err := sensor.ReadPressure()   // returns the pressure in centipascals

		if err != nil {
			println(err)
		} else {
			println("Temperature: " + strconv.FormatInt(int64(temp), 10) + " cC")
			println("Pressure:    " + strconv.FormatInt(int64(press), 10) + " cPa\n")
		}

		time.Sleep(time.Second)
	}
}
