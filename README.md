# TinyGo Drivers

[![PkgGoDev](https://pkg.go.dev/badge/tinygo.org/x/drivers)](https://pkg.go.dev/tinygo.org/x/drivers) [![Build](https://github.com/tinygo-org/drivers/actions/workflows/build.yml/badge.svg?branch=dev)](https://github.com/tinygo-org/drivers/actions/workflows/build.yml)


This package provides a collection of hardware drivers for devices such as sensors and displays that can be used together with [TinyGo](https://tinygo.org).

## Installing

```shell
go get tinygo.org/x/drivers
```

## How to use

Here is an example in TinyGo that uses the BMP180 digital barometer:

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

## Currently supported devices

The following 83 devices are supported.

| Device Name                                                                                                                                                                                         | Interface Type |
|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------|
| [ADT7410 I2C Temperature Sensor](https://www.analog.com/media/en/technical-documentation/data-sheets/ADT7410.pdf)                                                                                   | I2C |
| [ADXL345 accelerometer](http://www.analog.com/media/en/technical-documentation/data-sheets/ADXL345.pdf)                                                                                             | I2C |
| [AHT20 I2C Temperature and Humidity Sensor](http://www.aosong.com/userfiles/files/media/AHT20%20%E8%8B%B1%E6%96%87%E7%89%88%E8%AF%B4%E6%98%8E%E4%B9%A6%20A0%2020201222.pdf)                         | I2C |
| [AMG88xx 8x8 Thermal camera sensor](https://cdn-learn.adafruit.com/assets/assets/000/043/261/original/Grid-EYE_SPECIFICATIONS%28Reference%29.pdf)                                                   | I2C |
| [APA102 RGB LED](https://cdn-shop.adafruit.com/product-files/2343/APA102C.pdf)                                                                                                                      | SPI |
| [APDS9960 Digital proximity, ambient light, RGB and gesture sensor](https://cdn.sparkfun.com/assets/learn_tutorials/3/2/1/Avago-APDS-9960-datasheet.pdf)                                            | I2C |
| [AT24CX 2-wire serial EEPROM](https://www.openimpulse.com/blog/wp-content/uploads/wpsc/downloadables/24C32-Datasheet.pdf)                                                                           | I2C |
| [AXP192 single Cell Li-Battery and Power System Management](https://github.com/m5stack/M5-Schematic/blob/master/Core/AXP192%20Datasheet_v1.1_en_draft_2211.pdf)                                     | I2C |
| [BBC micro:bit LED matrix](https://github.com/bbcmicrobit/hardware/blob/master/SCH_BBC-Microbit_V1.3B.pdf)                                                                                          | GPIO |
| [BH1750 ambient light sensor](https://www.mouser.com/ds/2/348/bh1750fvi-e-186247.pdf)                                                                                                               | I2C |
| [BlinkM RGB LED](http://thingm.com/fileadmin/thingm/downloads/BlinkM_datasheet.pdf)                                                                                                                 | I2C |
| [BME280 humidity/pressure sensor](https://cdn-shop.adafruit.com/datasheets/BST-BME280_DS001-10.pdf)                                                                                                 | I2C |
| [BMI160 accelerometer/gyroscope](https://www.bosch-sensortec.com/media/boschsensortec/downloads/datasheets/bst-bmi160-ds000.pdf)                                                                    | SPI |
| [BMP180 barometer](https://cdn-shop.adafruit.com/datasheets/BST-BMP180-DS000-09.pdf)                                                                                                                | I2C |
| [BMP280 temperature/barometer](https://www.bosch-sensortec.com/media/boschsensortec/downloads/datasheets/bst-bmp280-ds001.pdf)                                                                      | I2C |
| [BMP388 pressure sensor](https://www.bosch-sensortec.com/media/boschsensortec/downloads/datasheets/bst-bmp388-ds001.pdf)                                                                            | I2C |
| [Buzzer](https://en.wikipedia.org/wiki/Buzzer#Piezoelectric)                                                                                                                                        | GPIO |
| [DHTXX thermometer and humidity sensor](https://cdn-shop.adafruit.com/datasheets/Digital+humidity+and+temperature+sensor+AM2302.pdf)                                                                | GPIO |
| [DS1307 real time clock](https://datasheets.maximintegrated.com/en/ds/DS1307.pdf)                                                                                                                   | I2C |
| [DS3231 real time clock](https://datasheets.maximintegrated.com/en/ds/DS3231.pdf)                                                                                                                   | I2C |
| [ESP32 as WiFi Coprocessor with Arduino nina-fw](https://github.com/arduino/nina-fw)                                                                                                                | SPI |
| [ESP8266/ESP32 AT Command set for WiFi/TCP/UDP](https://github.com/espressif/esp32-at)                                                                                                              | UART |
| [FT6336 touch controller](https://focuslcds.com/content/FT6236.pdf)                                                                                                                                 | I2C |
| [GPS module](https://www.u-blox.com/en/product/neo-6-series)                                                                                                                                        | I2C/UART |
| [HC-SR04 Ultrasonic distance sensor](https://cdn.sparkfun.com/datasheets/Sensors/Proximity/HCSR04.pdf)                                                                                              | GPIO |
| [HD44780 LCD controller](https://www.sparkfun.com/datasheets/LCD/HD44780.pdf)                                                                                                                       | GPIO/I2C |
| [HTS221 digital humidity and temperature sensor](https://www.st.com/resource/en/datasheet/hts221.pdf)                                                                                               | I2C |
| [HUB75 RGB led matrix](https://cdn-learn.adafruit.com/downloads/pdf/32x16-32x32-rgb-led-matrix.pdf)                                                                                                 | SPI |
| [software I2C driver](https://www.ti.com/lit/an/slva704/slva704.pdf)                                                                                                                                | GPIO |
| [ILI9341 TFT color display](https://cdn-shop.adafruit.com/datasheets/ILI9341.pdf)                                                                                                                   | SPI |
| [INA260 Volt/Amp/Power meter](https://www.ti.com/lit/ds/symlink/ina260.pdf)                                                                                                                         | I2C |
| [Infrared remote control](https://en.wikipedia.org/wiki/Consumer_IR)                                                                                                                                | GPIO |
| [IS31FL3731 matrix LED driver](https://www.lumissil.com/assets/pdf/core/IS31FL3731_DS.pdf)                                                                                                          | I2C |
| [4x4 Membrane Keypad](https://cdn.sparkfun.com/assets/f/f/a/5/0/DS-16038.pdf)                                                                                                                       | GPIO |
| [L293x motor driver](https://www.ti.com/lit/ds/symlink/l293d.pdf)                                                                                                                                   | GPIO/PWM |
| [L9110x motor driver](https://www.elecrow.com/download/datasheet-l9110.pdf)                                                                                                                         | GPIO/PWM |
| [LIS2MDL magnetometer](https://www.st.com/resource/en/datasheet/lis2mdl.pdf)                                                                                                                        | I2C |
| [LIS3DH accelerometer](https://www.st.com/resource/en/datasheet/lis3dh.pdf)                                                                                                                         | I2C |
| [LPS22HB MEMS nano pressure sensor](https://www.st.com/resource/en/datasheet/dm00140895.pdf)                                                                                                        | I2C |
| [LSM6DS3 accelerometer](https://www.st.com/resource/en/datasheet/lsm6ds3.pdf)                                                                                                                       | I2C |
| [LSM6DSOX accelerometer](https://www.st.com/resource/en/datasheet/lsm6dsox.pdf)                                                                                                                     | I2C |
| [LSM6DS3TR accelerometer](https://www.st.com/resource/en/datasheet/lsm6ds3tr.pdf)                                                                                                                   | I2C |
| [LSM303AGR accelerometer](https://www.st.com/resource/en/datasheet/lsm303agr.pdf)                                                                                                                   | I2C |
| [LSM9DS1 accelerometer](https://www.st.com/resource/en/datasheet/lsm9ds1.pdf)                                                                                                                       | I2C |
| [Makey Button](https://makeymakey.com/)                                                                                                                                                             | GPIO |
| [MAG3110 magnetometer](https://www.nxp.com/docs/en/data-sheet/MAG3110.pdf)                                                                                                                          | I2C |
| [MAX7219 & MAX7221 display driver](https://datasheets.maximintegrated.com/en/ds/MAX7219-MAX7221.pdf)                                                                                                | SPI |
| [MCP2515 Stand-Alone CAN Controller with SPI Interface](https://ww1.microchip.com/downloads/en/DeviceDoc/MCP2515-Family-Data-Sheet-DS20001801K.pdf)                                                 | SPI |
| [MCP3008 analog to digital converter (ADC)](http://ww1.microchip.com/downloads/en/DeviceDoc/21295d.pdf)                                                                                             | SPI |
| [MCP23017 port expander](https://ww1.microchip.com/downloads/en/DeviceDoc/20001952C.pdf)                                                                                                            | I2C |
| [Microphone - PDM](https://cdn-learn.adafruit.com/assets/assets/000/049/977/original/MP34DT01-M.pdf)                                                                                                | I2S/PDM |
| [MMA8653 accelerometer](https://www.nxp.com/docs/en/data-sheet/MMA8653FC.pdf)                                                                                                                       | I2C |
| [MPU6050 accelerometer/gyroscope](https://store.invensense.com/datasheets/invensense/MPU-6050_DataSheet_V3%204.pdf)                                                                                 | I2C |
| [P1AM-100 Base Controller](https://facts-engineering.github.io/modules/P1AM-100/P1AM-100.html)                                                                                                      | SPI |
| [PCD8544 display](http://eia.udg.edu/~forest/PCD8544_1.pdf)                                                                                                                                         | SPI |
| [PCF8563 real time clock](https://www.nxp.com/docs/en/data-sheet/PCF8563.pdf)                                                                                                                       | I2C |
| [Resistive Touchscreen (4-wire)](http://ww1.microchip.com/downloads/en/Appnotes/doc8091.pdf)                                                                                                        | GPIO |
| [RTL8720DN 2.4G/5G Dual Bands Wireless and BLE5.0](https://www.seeedstudio.com/Realtek8720DN-2-4G-5G-Dual-Bands-Wireless-and-BLE5-0-Combo-Module-p-4442.html)                                       | UART |
| [SCD4x CO2 Sensor](https://sensirion.com/media/documents/C4B87CE6/627C2DCD/CD_DS_SCD40_SCD41_Datasheet_D1.pdf)                                                                                      | I2C |
| [Semihosting](https://wiki.segger.com/Semihosting)                                                                                                                                                  | Debug |
| [Servo](https://learn.sparkfun.com/tutorials/hobby-servo-tutorial/all)                                                                                                                              | PWM |
| [Shift register (PISO)](https://en.wikipedia.org/wiki/Shift_register#Parallel-in_serial-out_\(PISO\))                                                                                               | GPIO |
| [Shift registers (SIPO)](https://en.wikipedia.org/wiki/Shift_register#Serial-in_parallel-out_(SIPO))                                                                                                | GPIO |
| [SHT3x Digital Humidity Sensor](https://www.sensirion.com/fileadmin/user_upload/customers/sensirion/Dokumente/2_Humidity_Sensors/Datasheets/Sensirion_Humidity_Sensors_SHT3x_Datasheet_digital.pdf) | I2C |
| [SHTC3 Digital Humidity Sensor (RH/T)](https://www.sensirion.com/fileadmin/user_upload/customers/sensirion/Dokumente/2_Humidity_Sensors/Datasheets/Sensirion_Humidity_Sensors_SHTC3_Datasheet.pdf)  | I2C |
| [SPI NOR Flash Memory](https://en.wikipedia.org/wiki/Flash_memory#NOR_flash)                                                                                                                        | SPI/QSPI |
| [SPI SDCARD/MMC](https://en.wikipedia.org/wiki/SD_card)                                                                                                                                             | SPI |
| [SSD1306 OLED display](https://cdn-shop.adafruit.com/datasheets/SSD1306.pdf)                                                                                                                        | I2C / SPI |
| [SSD1331 TFT color display](https://www.crystalfontz.com/controllers/SolomonSystech/SSD1331/381/)                                                                                                   | SPI |
| [SSD1351 OLED display](https://download.mikroe.com/documents/datasheets/ssd1351-revision-1.3.pdf)                                                                                                   | SPI |
| [ST7735 TFT color display](https://www.crystalfontz.com/controllers/Sitronix/ST7735R/319/)                                                                                                          | SPI |
| [ST7789 TFT color display](https://cdn-shop.adafruit.com/product-files/3787/3787_tft_QT154H2201__________20190228182902.pdf)                                                                        | SPI |
| [Stepper motor "Easystepper" controller](https://en.wikipedia.org/wiki/Stepper_motor)                                                                                                               | GPIO |
| [Thermistor](https://www.farnell.com/datasheets/33552.pdf)                                                                                                                                          | ADC |
| [TM1637 7-segment LED display](https://www.mcielectronics.cl/website_MCI/static/documents/Datasheet_TM1637.pdf)                                                                                     | I2C |
| [TMP102 I2C Temperature Sensor](https://download.mikroe.com/documents/datasheets/tmp102-data-sheet.pdf)                                                                                             | I2C |
| [UC8151 All-in-one driver IC for ESL](https://www.buydisplay.com/download/ic/UC8151C.pdf)                                                                                                           | I2C |
| [VEML6070 UV light sensor](https://www.vishay.com/docs/84277/veml6070.pdf)                                                                                                                          | I2C |
| [VL53L1X time-of-flight distance sensor](https://www.st.com/resource/en/datasheet/vl53l1x.pdf)                                                                                                      | I2C |
| [VL6180X time-of-flight distance sensor](https://www.st.com/resource/en/datasheet/vl6180x.pdf)                                                                                                      | I2C |
| [Waveshare 2.13" (B & C) e-paper display](https://www.waveshare.com/w/upload/d/d3/2.13inch-e-paper-b-Specification.pdf)                                                                             | SPI |
| [Waveshare 2.13" e-paper display](https://www.waveshare.com/w/upload/e/e6/2.13inch_e-Paper_Datasheet.pdf)                                                                                           | SPI |
| [Waveshare 2.9" e-paper display (V1)](https://www.waveshare.com/w/upload/e/e6/2.9inch_e-Paper_Datasheet.pdf)                                                                                        | SPI |
| [Waveshare 4.2" e-paper B/W display](https://www.waveshare.com/w/upload/6/6a/4.2inch-e-paper-specification.pdf)                                                                                     | SPI |
| [Waveshare GC9A01 TFT round display](https://www.waveshare.com/w/upload/5/5e/GC9A01A.pdf)                                                                             | SPI |
| [WS2812 RGB LED](https://cdn-shop.adafruit.com/datasheets/WS2812.pdf)                                                                                                                               | GPIO |
| [XPT2046 touch controller](http://grobotronics.com/images/datasheets/xpt2046-datasheet.pdf)                                                                                                         | GPIO |
| [Semtech SX126x Lora](https://www.semtech.com/products/wireless-rf/lora-transceiv-ers/sx1261)                                                                                                       | SPI |
| [SSD1289 TFT color display](http://aitendo3.sakura.ne.jp/aitendo_data/product_img/lcd/tft2/M032C1289TP/3.2-SSD1289.pdf)                                                                             | GPIO |

## Contributing

Your contributions are welcome!

Please take a look at our [CONTRIBUTING.md](./CONTRIBUTING.md) document for details.

## License

This project is licensed under the BSD 3-clause license, just like the [Go project](https://golang.org/LICENSE) itself.
