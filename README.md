# TinyGo Drivers

[![PkgGoDev](https://pkg.go.dev/badge/tinygo.org/x/drivers)](https://pkg.go.dev/tinygo.org/x/drivers) [![Build](https://github.com/tinygo-org/drivers/actions/workflows/build.yml/badge.svg?branch=dev)](https://github.com/tinygo-org/drivers/actions/workflows/build.yml)


This package provides a collection of over 100 different hardware drivers for devices such as sensors, displays, wireless adaptors, and actuators, that can be used together with [TinyGo](https://tinygo.org).

For the complete list, please see:
https://tinygo.org/docs/reference/devices/

## Installing

```shell
go get tinygo.org/x/drivers
```

## How to use

Here is an example in TinyGo that uses the BMP180 digital barometer.  This example should work on any board that supports I2C:

```go
package main

import (
    "time"

    "machine"

    "tinygo.org/x/drivers/bmp180"
)

func main() {
    machine.I2C0.Configure(machine.I2CConfig{})
    sensor := bmp180.New(machine.I2C0)
    sensor.Configure()

    connected := sensor.Connected()
    if !connected {
        println("BMP180 not detected")
        return
    }
    println("BMP180 detected")

    for {
        temp, _ := sensor.ReadTemperature()
        println("Temperature:", float32(temp)/1000, "°C")

        pressure, _ := sensor.ReadPressure()
        println("Pressure", float32(pressure)/100000, "hPa")

        time.Sleep(2 * time.Second)
    }
}
```

## Examples Using GPIO or SPI 

If compiling these examples directly you are likely to need to make minor changes to the defined variables to map the pins for the board you are using.  For example, this block in main.go:

```golang
var (
        spi   = machine.SPI0
        csPin = machine.D5
)
```

It might not be obvious, but you need to change these to match how you wired your specific board.  Constants are [defined for each supported microcontroller](https://tinygo.org/docs/reference/microcontrollers/).  

For example, to change the definitions for use on a Raspberry Pi Pico using typical wiring, you might need to do this:

```golang
var (
        spi   = machine.SPI0
        csPin = machine.GP17
)
```

## Contributing

Your contributions are welcome!

Please take a look at our [CONTRIBUTING.md](./CONTRIBUTING.md) document for details.

## License

This project is licensed under the BSD 3-clause license, just like the [Go project](https://golang.org/LICENSE) itself.
