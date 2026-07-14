package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/lsm303dlhc"
)

func main() {

	// LSM303DLHC is connected to the I2C0 bus on Adafruit Feather M4 via pins: 20(SDA) and 21(SCL).
	machine.I2C0.Configure(machine.I2CConfig{})

	sensor := lsm303dlhc.New(machine.I2C0)
	//default settings
	err := sensor.Configure(lsm303dlhc.Configuration{
		AccelPowerMode: lsm303dlhc.ACCEL_POWER_NORMAL,
		AccelRange:     lsm303dlhc.ACCEL_RANGE_2G,
		AccelDataRate:  lsm303dlhc.ACCEL_DATARATE_100HZ,
		MagPowerMode:   lsm303dlhc.MAG_POWER_NORMAL,
		MagSystemMode:  lsm303dlhc.MAG_SYSTEM_CONTINUOUS,
		MagDataRate:    lsm303dlhc.MAG_DATARATE_10HZ,
	})
	if err != nil {
		for {
			println("Failed to configure", err.Error())
			time.Sleep(time.Second)
		}
	}

	for {
		accel_x, accel_y, accel_z, err := sensor.ReadAcceleration()
		if err != nil {
			println("Failed to read accel", err.Error())
		}
		println("ACCEL_X:", accel_x, " ACCEL_Y:", accel_y, " ACCEL_Z:", accel_z)

		mag_x, mag_y, mag_z, err := sensor.ReadMagneticField()
		if err != nil {
			println("Failed to read mag", err.Error())
		}
		println("MAG_X:", mag_x, " MAG_Y:", mag_y, " MAG_Z:", mag_z)

		pitch, roll, _ := sensor.ReadPitchRoll()
		println("Pitch:", float32(pitch), " Roll:", float32(roll))

		heading, _ := sensor.ReadCompass()
		println("Heading:", float32(heading), "degrees")

		temp, _ := sensor.ReadTemperature()
		println("Temperature:", float32(temp)/1000, "*C")

		println("\n")
		time.Sleep(time.Millisecond * 250)
	}

}
