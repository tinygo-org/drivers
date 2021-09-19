// Connects to an LSM6DSOX I2C a 6 axis Inertial Measurement Unit (IMU)
package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/lsm6dsox"
)

const (
	PLOTTER           = true
	SHOW_ACCELERATION = false
	SHOW_ROTATION     = true
	SHOW_TEMPERATURE  = false

	// calibration is naive, good enough for a demo: do not shake a second after flashing
	GYRO_CALIBRATION = true
)

var cal = [...]float32{0, 0, 0}

func main() {

	machine.I2C0.Configure(machine.I2CConfig{})

	device := lsm6dsox.New(machine.I2C0)
	device.Configure(lsm6dsox.Configuration{
		AccelRange:      lsm6dsox.ACCEL_2G,
		AccelSampleRate: lsm6dsox.ACCEL_SR_104,
		GyroRange:       lsm6dsox.GYRO_250DPS,
		GyroSampleRate:  lsm6dsox.GYRO_SR_104,
	})

	for {

		if !device.Connected() {
			println("LSM6DSOX not connected")
			time.Sleep(time.Second)
			continue
		}

		// heuristic: after successful calibration the value can't be 0
		if GYRO_CALIBRATION && cal[0] == 0 {
			calibrateGyro(device)
		}

		ax, ay, az := device.ReadAcceleration()
		gx, gy, gz := device.ReadRotation()
		t, _ := device.ReadTemperature()

		if PLOTTER {
			printPlotter(ax, ay, az, gx, gy, gz, t)
			time.Sleep(time.Millisecond * 100)
		} else {
			printMonitor(ax, ay, az, gx, gy, gz, t)
			time.Sleep(time.Millisecond * 1000)
		}

	}

}

func calibrateGyro(device *lsm6dsox.Device) {
	for i := 0; i < 100; i++ {
		gx, gy, gz := device.ReadRotation()
		cal[0] += float32(gx) / 1000000
		cal[1] += float32(gy) / 1000000
		cal[2] += float32(gz) / 1000000
		time.Sleep(time.Millisecond * 10)
	}
	cal[0] /= 100
	cal[1] /= 100
	cal[2] /= 100
}

// Arduino IDE's Serial Plotter
func printPlotter(ax, ay, az, gx, gy, gz, t int32) {
	if SHOW_ACCELERATION {
		fmt.Printf("AX:%f, AY:%f, AZ:%f,", axis(ax, 0), axis(ay, 0), axis(az, 0))
	}
	if SHOW_ROTATION {
		fmt.Printf("GX:%f, GY:%f, GZ:%f,", axis(gx, cal[0]), axis(gy, cal[1]), axis(gz, cal[2]))
	}
	if SHOW_TEMPERATURE {
		fmt.Printf("T:%f", float32(t)/1000)
	}
	println()
}

// Any Serial Monitor
func printMonitor(ax, ay, az, gx, gy, gz, t int32) {
	if SHOW_ACCELERATION {
		fmt.Printf("Acceleration: %f, %f, %f\r\n", axis(ax, 0), axis(ay, 0), axis(az, 0))
	}
	if SHOW_ROTATION {
		fmt.Printf("Rotation: %f, %f, %f\r\n", axis(gx, cal[0]), axis(gy, cal[1]), axis(gz, cal[2]))
	}
	if SHOW_TEMPERATURE {
		fmt.Printf("Temperature C: %f\r\n", float32(t)/1000)
	}
	println()
}

func axis(raw int32, cal float32) float32 {
	return float32(raw)/1000000 - cal
}
