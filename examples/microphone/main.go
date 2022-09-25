// Example using the i2s hardware interface on the Adafruit Circuit Playground Express
// to read data from the onboard MEMS microphone.
//
// Uses ideas from the https://github.com/adafruit/Adafruit_CircuitPlayground repo.
package main

import (
	"machine"

	"tinygo.org/x/drivers/microphone"
)

const (
	defaultSampleRate        = 22000
	quantizeSteps            = 64
	msForSPLSample           = 50
	defaultSampleCountForSPL = (defaultSampleRate / 1000) * msForSPLSample
)

func main() {
	machine.I2S0.Configure(machine.I2SConfig{
		Mode:           machine.I2SModePDM,
		AudioFrequency: defaultSampleRate * quantizeSteps / 16,
		ClockSource:    machine.I2SClockSourceExternal,
		Stereo:         true,
	})

	mic := microphone.New(machine.I2S0)
	mic.SampleCountForSPL = defaultSampleCountForSPL
	mic.Configure()

	for {
		spl, maxval := mic.GetSoundPressure()
		println("C", spl, "max", maxval)
	}
}
