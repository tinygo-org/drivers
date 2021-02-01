// Package thermistor is for temperature sensing using a thermistor
// such as the NTC 3950.
//
// Datasheet: https://www.farnell.com/datasheets/33552.pdf
//
// This code is an interpretation of Adafruit Thermistor module in Python:
// https://github.com/adafruit/Adafruit_CircuitPython_Thermistor
//
// It uses the Steinhart–Hart equation to calculate the temperature
// based on the resistance:
// https://en.wikipedia.org/wiki/Steinhart%E2%80%93Hart_equation
//
// To use with other thermistors adjust the BCoefficient and NominalTemperature
// values to match the specific thermistor you wish to use.
//
//			sensor.NominalTemperature = 25
//			sensor.BCoefficient = 3950
//
// Set the SeriesResistor and NominalResistance based on the microcontroller voltage and
// circuit that you have in use. Set HighSide based on if the thermistor is connected from
// the ADC pin to the powered side (true) or to ground (false).
//
//			sensor.SeriesResistor = 10000
//			sensor.NominalResistance = 10000
//			sensor.HighSide = true
//
package thermistor // import "tinygo.org/x/drivers/thermistor"

import (
	"machine"
	"math"
)

// Device holds the ADC pin and the needed settings for calculating the
// temperature based on the resistance.
type Device struct {
	adc                *machine.ADC
	SeriesResistor     uint32
	NominalResistance  uint32
	NominalTemperature uint32
	BCoefficient       uint32
	HighSide           bool
}

// New returns a new thermistor driver given an ADC pin.
func New(pin machine.Pin) Device {
	adc := machine.ADC{pin}
	return Device{
		adc:                &adc,
		SeriesResistor:     10000,
		NominalResistance:  10000,
		NominalTemperature: 25,
		BCoefficient:       3950,
		HighSide:           true,
	}
}

// Configure configures the ADC pin used for the thermistor.
func (d *Device) Configure() {
	d.adc.Configure(machine.ADCConfig{})
}

// ReadTemperature returns the temperature in celsius milli degrees (°C/1000)
func (d *Device) ReadTemperature() (temperature int32, err error) {
	var reading uint32
	if d.HighSide {
		// Thermistor connected from analog input to high logic level.
		val := d.adc.Get()
		reading = uint32(val) / 64
		reading = (1023 * d.SeriesResistor) / reading
		reading -= d.SeriesResistor
	} else {
		// Thermistor connected from analog input to ground.
		reading = d.SeriesResistor / uint32(65535/d.adc.Get()-1)
	}

	var steinhart float64
	steinhart = float64(reading) / float64(d.NominalResistance) // (R/Ro)
	steinhart = math.Log(steinhart)                             // ln(R/Ro)
	steinhart /= float64(d.BCoefficient)                        // 1/B * ln(R/Ro)
	steinhart += 1.0 / (float64(d.NominalTemperature) + 273.15) // + (1/To)
	steinhart = 1.0 / steinhart                                 // Invert
	steinhart -= 273.15                                         // convert to C

	return int32(steinhart * 1000), nil
}
