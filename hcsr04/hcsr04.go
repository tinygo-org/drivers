// Package hcsr04 provides a driver for the HC-SR04 ultrasonic distance sensor
//
// Datasheet:
// https://cdn.sparkfun.com/datasheets/Sensors/Proximity/HCSR04.pdf
package hcsr04

import (
	"time"

	"tinygo.org/x/drivers/internal/pin"
)

const TIMEOUT = 23324 // max sensing distance (4m)

// Device holds the pins
type Device struct {
	trigger       pin.OutputFunc
	echo          pin.InputFunc
	configurePins func()
}

// New returns a new ultrasonic driver given 2 pins
func New(trigger pin.Output, echo pin.Input) Device {
	return Device{
		trigger: trigger.Set,
		echo:    echo.Get,
		configurePins: func() {
			pin.ConfigureOutput(trigger)
			pin.ConfigureInput(echo)
		},
	}
}

// Configure configures the pins of the Device
func (d *Device) Configure() {
	if d.configurePins == nil {
		panic(pin.ErrConfigBeforeInstantiated)
	}
	d.configurePins()
}

// ReadDistance returns the distance of the object in mm
func (d *Device) ReadDistance() int32 {
	pulse := d.ReadPulse()

	// sound speed is 343000 mm/s
	// pulse is roundtrip measured in microseconds
	// distance = velocity * time
	// 2 * distance = 343000 * (pulse/1000000)
	return (pulse * 1715) / 10000 //mm
}

// ReadPulse returns the time of the pulse (roundtrip) in microseconds
func (d *Device) ReadPulse() int32 {
	t := time.Now()
	d.trigger.Low()
	time.Sleep(2 * time.Microsecond)
	d.trigger.High()
	time.Sleep(10 * time.Microsecond)
	d.trigger.Low()
	i := uint8(0)
	for {
		if d.echo() {
			t = time.Now()
			break
		}
		i++
		if i > 10 {
			if time.Since(t).Microseconds() > TIMEOUT {
				return 0
			}
			i = 0
		}
	}
	i = 0
	for {
		if !d.echo() {
			return int32(time.Since(t).Microseconds())
		}
		i++
		if i > 10 {
			if time.Since(t).Microseconds() > TIMEOUT {
				return 0
			}
			i = 0
		}
	}
	return 0
}
