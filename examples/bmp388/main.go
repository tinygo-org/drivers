package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/bmp388"
)

func main() {

	time.Sleep(time.Second)
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
	err := sensor.Configure(bmp388.BMP388Config{
		Pressure:         bmp388.SAMPLING_8X,
		Temperature:      bmp388.SAMPLING_2X,
		ODR:              bmp388.ODR_25,
		IIR:              bmp388.COEF_0,
		Mode:             bmp388.NORMAL,
		SeaLevelPressure: 10186257, // in centipascals
	})

	// This is also fine
	// err := sensor.Configure(bmp388.BMP388Config{})

	if err != nil {
		println(err)
	}

	for {
		temp, err := sensor.ReadTemperature() // returns the temperature in centicelsius
		press, err := sensor.ReadPressure()   // returns the pressure in centipascals
		alt, err := sensor.ReadAltitude()     // estimates the altitude in centimeters given the local sea level pressure

		if err != nil {
			println(err)
		}

		fmt.Printf("Temperature: %d cC\r\n", temp)
		fmt.Printf("Pressure:    %d cPa\r\n", press)
		fmt.Printf("Altitude:    %d cm\r\n\n", alt)
		time.Sleep(time.Second)
	}
}
