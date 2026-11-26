package main

import (
	"device/sam"
	"encoding/binary"
	"machine"
	"main/imu"
	"strconv"
	"time"
)

var (
	// declare a I2C port - configure from generic sercomSERIAL
	I2C_MPU6050 = &machine.I2C{Bus: sam.SERCOM3_I2CM, SERCOM: 3}
)

func main() {
	// wait for terminal to connect
	time.Sleep(time.Millisecond * 500)
	println("Initialize Accelerometer on MPU6050 ...")

	// setup a sercomSERIAL for I2C use
	err := I2C_MPU6050.Configure(machine.I2CConfig{
		//Frequency: uint32(400e3), // 400kHz
		Frequency: uint32(100e3), // 100kHz
	})

	if err != nil {
		println("Initiating I2C failed. Aborting")
		for {
			// failure
		}
	}

	// Create and Initialise MPU6050 for DMP mode
	imu, err := imu.New(I2C_MPU6050)
	if err != nil {
		println("Unable to create a MPU6050 instance. Aborting.")
		for {
			// failure
		}
	}

	if true {
		err = imu.Init()
		if err != nil {
			println("Unable to initialize IMU. Aborting.")
			for {
				// failure
			}
		}
		imu.Calibrate()
		println("MPU calibrated!!")
		imu.PrintIMUOffsets()
		for {
			angles, err := imu.GetYawPitchRoll()
			if err == nil {
				println(
					strconv.FormatFloat(float64(angles.Yaw), 'f', 2, 32), "\t",
					strconv.FormatFloat(float64(angles.Pitch), 'f', 2, 32), "\t",
					strconv.FormatFloat(float64(angles.Roll), 'f', 2, 32))
			}
			time.Sleep(time.Millisecond * 100) // sleep for 100 milliseconds - allow Animation.py to keep up
		}
	}
}

func test_tinygo_internal_integer_endianess() {
	println("Testing endianess")
	// test endainess
	test := uint16(0xAABB)
	b := make([]byte, 2)
	// There are two ways to load the 16 bit integers
	// 1. Th e manual method
	b[0] = byte(test >> 8 & 0xFF)
	b[1] = byte(test & 0xFF)
	println("0x", strconv.FormatUint(uint64(b[0]), 16), strconv.FormatUint(uint64(b[1]), 16))
	if b[0] == 0xAA && b[1] == 0xBB {
		println("Tinygo integer is stored as BIG ENDIAN")
	}
	// 2. The GO library  way
	binary.BigEndian.PutUint16(b, test)
	println("0x", strconv.FormatUint(uint64(b[0]), 16), strconv.FormatUint(uint64(b[1]), 16))
	if b[0] == 0xAA && b[1] == 0xBB {
		println("Tinygo integer is stored as BIG ENDIAN")
	}
}
