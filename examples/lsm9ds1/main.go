// LSM9DS1, 9 axis Inertial Measurement Unit (IMU)
package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/lsm9ds1"
)

const (
	PLOTTER             = false
	SHOW_ACCELERATION   = true
	SHOW_ROTATION       = true
	SHOW_MAGNETIC_FIELD = true
	SHOW_TEMPERATURE    = true
)

func main() {

	// I2C configure
	machine.I2C0.Configure(machine.I2CConfig{})

	// LSM9DS1 setup
	device := lsm9ds1.New(machine.I2C0)
	err := device.Configure(lsm9ds1.Configuration{
		AccelRange:      lsm9ds1.ACCEL_2G,
		AccelSampleRate: lsm9ds1.ACCEL_SR_119,
		GyroRange:       lsm9ds1.GYRO_250DPS,
		GyroSampleRate:  lsm9ds1.GYRO_SR_119,
		MagRange:        lsm9ds1.MAG_4G,
		MagSampleRate:   lsm9ds1.MAG_SR_40,
	})
	if err != nil {
		for {
			println("Failed to configure", err.Error())
			time.Sleep(time.Second)
		}
	}

	for {

		if con, err := device.Connected(); !con || err != nil {
			println("LSM9DS1 not connected")
			time.Sleep(time.Second)
			continue
		}

		ax, ay, az, _ := device.ReadAcceleration()
		gx, gy, gz, _ := device.ReadRotation()
		mx, my, mz, _ := device.ReadMagneticField()
		t, _ := device.ReadTemperature()

		if PLOTTER {
			printPlotter(ax, ay, az, gx, gy, gz, mx, my, mz, t)
			time.Sleep(time.Millisecond * 100)
		} else {
			printMonitor(ax, ay, az, gx, gy, gz, mx, my, mz, t)
			time.Sleep(time.Millisecond * 1000)
		}

	}

}

// Arduino IDE's Serial Plotter
func printPlotter(ax, ay, az, gx, gy, gz, mx, my, mz, t int32) {
	if SHOW_ACCELERATION {
		fmt.Printf("AX:%f, AY:%f, AZ:%f,", axis(ax), axis(ay), axis(az))
	}
	if SHOW_ROTATION {
		fmt.Printf("GX:%f, GY:%f, GZ:%f,", axis(gx), axis(gy), axis(gz))
	}
	if SHOW_MAGNETIC_FIELD {
		fmt.Printf("MX:%d, MY:%d, MZ:%d,", mx, my, mz)
	}
	if SHOW_TEMPERATURE {
		fmt.Printf("T:%f", float32(t)/1000)
	}
	println()
}

// Any Serial Monitor
func printMonitor(ax, ay, az, gx, gy, gz, mx, my, mz, t int32) {
	if SHOW_ACCELERATION {
		fmt.Printf("Acceleration (g): %f, %f, %f\r\n", axis(ax), axis(ay), axis(az))
	}
	if SHOW_ROTATION {
		fmt.Printf("Rotation (dps): %f, %f, %f\r\n", axis(gx), axis(gy), axis(gz))
	}
	if SHOW_MAGNETIC_FIELD {
		fmt.Printf("Magnetic field (nT): %d, %d, %d\r\n", mx, my, mz)
	}
	if SHOW_TEMPERATURE {
		fmt.Printf("Temperature C: %f\r\n", float32(t)/1000)
	}
	println()
}

func axis(raw int32) float32 {
	return float32(raw) / 1000000
}
