package main

import (
	"machine"
	"machine/usb/hid/mouse"
	"math"
	"time"

	"tinygo.org/x/drivers/as560x"
)

func main() {
	// Let's use the AS5600 to make the world's most useless mouse with just a single X-axis & no buttons (!)
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_400KHZ,
		SDA:       machine.GPIO4,
		SCL:       machine.GPIO5,
	})
	as5600 := as560x.NewAS5600(machine.I2C0)
	as5600.Configure(as560x.Config{})
	mouse := mouse.New()

	lastAngle := -1
	for {
		time.Sleep(time.Millisecond * 10)
		// Get the magnet status of the AS5600
		magnetDetected, magnetStrength, err := as5600.MagnetStatus()
		if err != nil {
			continue
		}
		// Get the raw angle from the AS5600
		angle, _, err := as5600.RawAngle(as560x.ANGLE_NATIVE)
		if err != nil {
			continue
		}
		str := ""
		if !magnetDetected {
			str += "NOT "
		}
		str += "detected. Strength is "
		switch magnetStrength {
		case as560x.MagnetTooWeak:
			str += "too weak"
		case as560x.MagnetTooStrong:
			str += "too strong"
		default:
			str += "ok"
		}
		println("Raw angle:", angle, "Magnet was", str)
		if lastAngle != -1 {
			diff := int(angle) - lastAngle
			// correct the zero crossover glitch
			if diff < -0xc00 {
				diff += 0xfff
			} else if diff > 0xc00 {
				diff -= 0xfff
			}
			// debounce the noise (could use the sensor's filters/hysteresis instead?)
			if math.Abs(float64(diff)) > 2 {
				// move the mouse x-axis in response to the AS5600
				mouse.Move(diff, 0)
			}
		}
		lastAngle = int(angle)
	}
}
