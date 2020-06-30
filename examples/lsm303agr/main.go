package main

import (
	"machine"
	"time"
	"tinygo.org/x/drivers/lsm303agr"
)

func main() {

	machine.I2C0.Configure(machine.I2CConfig{})
	accel_mag := lsm303agr.New(machine.I2C0)

	if !accel_mag.Connected() {
		println("LSM303AGR/MAG not connected!")
		return
	}

	accel_mag.Configure(lsm303agr.Configuration{}) //default settings

	for {

		accel_x, accel_y, accel_z := accel_mag.ReadAcceleration()
		pitch, roll := accel_mag.ReadPitchRoll()
		mag_x, mag_y, mag_z := accel_mag.ReadMagneticField()
		heading := accel_mag.ReadCompass()
		temp, _ := accel_mag.ReadTemperature()

		println("ACCEL_X:", accel_x, " ACCEL_Y:", accel_y, " ACCEL_Z:", accel_z)
		println("MAG_X:", mag_x, " MAG_Y:", mag_y, " MAG_Z:", mag_z)
		println("Pitch:", pitch, " Roll:", roll)
		println("Heading:", heading)
		println("Temperature:", temp/1000)
		println("\n")

		time.Sleep(time.Millisecond * 100)
	}

}
