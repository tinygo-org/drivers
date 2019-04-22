// Package drivers provides a collection of hardware drivers for devices that
// can be used together with TinyGo (https://tinygo.org).
//
// Here is an example in TinyGo that uses the BMP180 digital barometer:
//
// 	package main
//
// 	import (
//		"time"
//		"machine"
//
// 		"github.com/tinygo-org/drivers/bmp180"
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
//			temp, _ := sensor.Temperature()
//			println("Temperature:", float32(temp)/1000, "ºC")
//
//			pressure, _ := sensor.Pressure()
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
package drivers
