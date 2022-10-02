// Connects to an QMI8658C I2C accelerometer/gyroscope and print the read data.
// This example was made with the "WaveShare RP2040 Round LCD 1.28in" in mind.
// For more infor about this development board:
// https://www.waveshare.com/wiki/RP2040-LCD-1.28
package main

import (
	"machine"
	"time"

	imu "tinygo.org/x/drivers/qmi8658c"
)

func main() {
	i2c := machine.I2C1
	// This is the default pinout for the "WaveShare RP2040 Round LCD 1.28in"
	err := i2c.Configure(machine.I2CConfig{
		SDA:       machine.GP6,
		SCL:       machine.GP7,
		Frequency: 100000,
	})
	if err != nil {
		println("unable to configure I2C:", err)
		return
	}
	// Create a new device
	d := imu.New(i2c)

	// Check if the device is connected
	if !d.Connected() {
		println("unable to connect to sensor")
		return
	}

	// This IMU has multiple configurations like output data rate, multiple
	// measurements scales, low pass filters, low power modes, all the vailable
	// values can be found in the datasheet and were defined at registers file.
	// This is the default configuration which will be used if the `nil` value
	// is passed do the `Configure` method.
	config := imu.Config{
		SPIMode:     imu.SPI_4_WIRE,
		SPIEndian:   imu.SPI_BIG_ENDIAN,
		SPIAutoInc:  imu.SPI_AUTO_INC,
		AccEnable:   imu.ACC_ENABLE,
		AccScale:    imu.ACC_8G,
		AccRate:     imu.ACC_NORMAL_1000HZ,
		AccLowPass:  imu.ACC_LOW_PASS_2_62,
		GyroEnable:  imu.GYRO_FULL_ENABLE,
		GyroScale:   imu.GYRO_512DPS,
		GyroRate:    imu.GYRO_1000HZ,
		GyroLowPass: imu.GYRO_LOW_PASS_2_62,
	}
	d.Configure(config)

	// Read the accelation, rotation and temperature data and print them.
	for {
		acc_x, acc_y, acc_z := d.ReadAcceleration()
		gyro_x, gyro_y, gyro_z := d.ReadRotation()
		temp, _ := d.ReadTemperature()
		println("-------------------------------")
		println("acc:", acc_x, acc_y, acc_z)
		println("gyro:", gyro_x, gyro_y, gyro_z)
		println("temp:", temp)
		time.Sleep(time.Millisecond * 100)
	}
}
