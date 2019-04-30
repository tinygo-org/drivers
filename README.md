# TinyGo Drivers

[![GoDoc](https://godoc.org/github.com/tinygo-org/drivers?status.svg)](https://godoc.org/github.com/tinygo-org/drivers) [![CircleCI](https://circleci.com/gh/tinygo-org/drivers/tree/master.svg?style=svg)](https://circleci.com/gh/tinygo-org/drivers/tree/master)


This package provides a collection of hardware drivers for devices that can be used together with [TinyGo](https://tinygo.org).

## Installing

```shell
go get github.com/tinygo-org/drivers
```

## How to use

Here is an example in TinyGo that uses the BMP180 digital barometer:

```go
package main

import (
    "time"

    "machine"

    "github.com/tinygo-org/drivers/bmp180"
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
        println("Temperature:", float32(temp)/1000, "ºC")

        pressure, _ := sensor.ReadPressure()
        println("Pressure", float32(pressure)/100000, "hPa")

        time.Sleep(2 * time.Second)
    }
}
```

## Currently supported devices

| Device Name | Interface Type |
|----------|-------------|
| [ADXL345 accelerometer](http://www.analog.com/media/en/technical-documentation/data-sheets/ADXL345.pdf) | I2C |
| [APA102 RGB LED](https://cdn-shop.adafruit.com/product-files/2343/APA102C.pdf) | SPI |
| [BH1750 ambient light sensor](https://www.mouser.com/ds/2/348/bh1750fvi-e-186247.pdf) | I2C |
| [BlinkM RGB LED](http://thingm.com/fileadmin/thingm/downloads/BlinkM_datasheet.pdf) | I2C |
| [BMP180 barometer](https://cdn-shop.adafruit.com/datasheets/BST-BMP180-DS000-09.pdf) | I2C |
| [DS3231 real time clock](https://datasheets.maximintegrated.com/en/ds/DS3231.pdf) | I2C |
| ["Easystepper" stepper motor controller](https://en.wikipedia.org/wiki/Stepper_motor) | GPIO |
| [ESP8266/ESP32 AT Command set for WiFi/TCP/UDP](https://github.com/espressif/esp32-at) | UART |
| [LIS3DH accelerometer](https://www.st.com/resource/en/datasheet/lis3dh.pdf) | I2C |
| [MAG3110 magnetometer](https://www.nxp.com/docs/en/data-sheet/MAG3110.pdf) | I2C |
| [MMA8653 accelerometer](https://www.nxp.com/docs/en/data-sheet/MMA8653FC.pdf) | I2C |
| [MPU6050 accelerometer/gyroscope](https://store.invensense.com/datasheets/invensense/MPU-6050_DataSheet_V3%204.pdf) | I2C |
| [SSD1306 OLED display](https://cdn-shop.adafruit.com/datasheets/SSD1306.pdf) | I2C / SPI |
| [Thermistor](https://www.farnell.com/datasheets/33552.pdf) | ADC |
| [VL53L1X time-of-flight distance sensor](https://www.st.com/resource/en/datasheet/vl53l1x.pdf) | I2C |
| [WS2812 RGB LED](https://cdn-shop.adafruit.com/datasheets/WS2812.pdf) | GPIO |

## Contributing

This collection of drivers is part of the [TinyGo](https://github.com/tinygo-org/tinygo) project. Patches are welcome but new drivers should follow the patterns established by similar existing drivers.

## License

This project is licensed under the BSD 3-clause license, just like the [Go project](https://golang.org/LICENSE) itself.
