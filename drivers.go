// Package drivers provides a collection of hardware drivers for TinyGo (https://tinygo.org)
// for devices such as sensors and displays.
//
// Here is an example in TinyGo that uses the BMP180 digital barometer:
//
// 	package main
//
// 	import (
//		"time"
//		"machine"
//
// 		"tinygo.org/x/drivers/bmp180"
// 	)
//
// 	func main() {
//		machine.I2C0.Configure(machine.I2CConfig{})
//		sensor := bmp180.New(machine.I2C0)
//		sensor.Configure()
//
// 		connected := sensor.Connected()
// 		if !connected {
//			println("BMP180 not detected")
//			return
//		}
//		println("BMP180 detected")
//
//		for {
//			temp, _ := sensor.ReadTemperature()
//			println("Temperature:", float32(temp)/1000, "°C")
//
//			pressure, _ := sensor.ReadPressure()
//			println("Pressure", float32(pressure)/100000, "hPa")
//
//			time.Sleep(2 * time.Second)
//		}
//	}
//
// Each individual driver is contained within its own sub-package within this package and
// there are no interdependencies in order to minimize the final size of compiled code that
// uses any of these drivers.
//
package drivers // import "tinygo.org/x/drivers"
