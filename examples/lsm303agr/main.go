/*
Example code for package lsm303agr, which implements a driver for the LSM303AGR, 
 a 3 axis accelerometer/magnetic sensor included onblard BBC micro:bits v1.5.

Datasheet: https://www.st.com/resource/en/datasheet/lsm303agr.pdf
micro:bit versions: https://tech.microbit.org/hardware/i2c/
*/

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
        
        accel_x, accel_y, accel_z := accel_mag.ReadAcceleration() // acceleration of all axis (1000+ = 1g)
        pitch, roll := accel_mag.ReadPitchRoll() // pitch and roll degrees
        mag_x, mag_y, mag_z := accel_mag.ReadMagneticField() // magnetic field level of all axis
        heading := accel_mag.ReadCompassHeading() // compass heading (-180~180, may not be accurate)
        temp := accel_mag.ReadTemperature() // temperature in Celsius
        
        println("ACCEL_X:", accel_x, " ACCEL_Y:", accel_y, " ACCEL_Z:", accel_z)
        println("MAG_X:", mag_x, " MAG_Y:", mag_y, " MAG_Z:", mag_z)
        println("Pitch:", pitch, " Roll:", roll)
        println("Heading:", heading)
        println("Temperature:", temp)
        println("\n")
        
        time.Sleep(time.Millisecond * 500)
    }
       
}
