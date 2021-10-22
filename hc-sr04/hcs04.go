package main

import (
	"machine"
	"runtime/interrupt"
	"time"
)

// Dev represents a handle to HC-SR04 ultrasonic distance sensor.
type Dev struct {
	out     machine.Pin
	echo    machine.Pin
	timeout time.Duration
}

// New instantiates a handle to HC-SR04 ultrasonic distance sensor.
// Calls to pin.Configure are called as well. Timeout is the max
// time-of-flight permissible. Ideally no more than a few milliseconds.
// Here's a table of max measurable distance and timeout required:
//  10 centimeter: 600 micosecond timeout
//  1 meter: approximately 6millisecond timeout
// These come from the calculation:
//  t = 2*d*c
// where d is the desired measurable distance and c is the speed of sound (343 meters per second).
// Keep in mind the hardware is limited to around 2 meters max distance.
func New(out, echo machine.Pin, timeout time.Duration) *Dev {
	out.Configure(machine.PinConfig{Mode: machine.PinOutput})
	echo.Configure(machine.PinConfig{Mode: machine.PinInput})
	return &Dev{
		out:     out,
		echo:    echo,
		timeout: timeout,
	}
}

// ReadDistance returns measured distance in millimeters.
func (d *Dev) ReadDistance() uint32 {
	d.out.Low()
	time.Sleep(2 * time.Microsecond)
	d.out.High()
	time.Sleep(10 * time.Microsecond)
	d.out.Low()
	si := interrupt.Disable()
	pulseWidth := pulseIn(d.echo, true, d.timeout)
	interrupt.Restore(si)
	return uint32(pulseWidth.Nanoseconds() / 58e2)
}

// pulseIn does more or less what Arduino's pulseIn does.
func pulseIn(p machine.Pin, val bool, timeout time.Duration) time.Duration {
	start := time.Now()
	for p.Get() == val { // let the current value pass if present.
		time.Sleep(time.Microsecond)
		if time.Since(start) > timeout {
			return 0
		}
	}
	for p.Get() != val { // wait for desired pulse
		time.Sleep(time.Microsecond)
		if time.Since(start) > timeout {
			return 0
		}
	}
	start = time.Now() // we got the desired pulese
	for p.Get() == val {
		if time.Since(start) > timeout {
			return 0
		}
		time.Sleep(time.Microsecond)
	}
	return time.Since(start)
}
