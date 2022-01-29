package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/lsm303agr"
)

func main() {

	// LSM303AGR/MAG is connected to the I2C0 bus on micro:bit v1 (the same as P19/P20) and v2 (internal)
	machine.I2C0.Configure(machine.I2CConfig{})

	sensor := lsm303agr.New(machine.I2C0)
	err := sensor.Configure(lsm303agr.Configuration{}) //default settings
	if err != nil {
		for {
			println("Failed to configure", err.Error())
			time.Sleep(time.Second)
		}
	}

	// you can specify the following options to adjust accuracy, sensor range or save power.
	// see https://github.com/tinygo-org/drivers/blob/release/lsm303agr/registers.go for details:
	/*
		sensor.Configure(lsm303agr.Configuration{
			AccelPowerMode: lsm303agr.ACCEL_POWER_NORMAL,
			AccelRange:     lsm303agr.ACCEL_RANGE_2G,
			AccelDataRate:  lsm303agr.ACCEL_DATARATE_100HZ,
			MagPowerMode:   lsm303agr.MAG_POWER_NORMAL,
			MagSystemMode:  lsm303agr.MAG_SYSTEM_CONTINUOUS,
			MagDataRate:    lsm303agr.MAG_DATARATE_10HZ,
		})
	*/

	for {

		if !sensor.Connected() {
			println("LSM303AGR/MAG not connected!")
			time.Sleep(time.Second)
			continue
		}

		// accel_x, accel_y, accel_z := sensor.ReadAcceleration()
		// println("ACCEL_X:", accel_x/100000, " ACCEL_Y:", accel_y/100000, " ACCEL_Z:", accel_z/100000)

		// mag_x, mag_y, mag_z := sensor.ReadMagneticField()
		// println("MAG_X:", mag_x/100000, " MAG_Y:", mag_y/100000, " MAG_Z:", mag_z/100000)

		pitch, roll, _ := sensor.ReadPitchRoll()
		println("Pitch:", float32(pitch)/100000, " Roll:", float32(roll)/100000)

		heading, _ := sensor.ReadCompass()
		println("Heading:", float32(heading)/100000, "degrees")

		temp, _ := sensor.ReadTemperature()
		println("Temperature:", float32(temp)/1000, "*C")

		println("\n")
		time.Sleep(time.Millisecond * 250)
	}

}
