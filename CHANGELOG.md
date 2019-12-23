0.9.0
---
- **new devices**
    - net: shared implementation of net package for serial wifi devices
    - shifter: add support for bit Parallel In Serial Out (PISO) shifter
    - stepper: add support for dual stepper motor
    - wifinina: add implementation for WiFiNINA firmware
- **enhancements**
    - st7735: improvements in st7735 driver
    - st7789: improvements in st7789 driver
    - ws2812: add support for 120Mhz Cortex-M4
    - ws2812: added Feather M0 and Trinket M0 to build tags for WS2812
    - ws2812: add support for simulation
- **bugfixes**
    - ws2812: fix "invalid symbol redefinition" error
- **examples**
    - Add examples for wifinina drivers

0.8.0
---
- **new devices**
    - mcp3008: add implementation for MCP3008 ADC with SPI interface
    - semihosting: initial implementation of ARM semihosting
- **enhancements**
    - espat: refactor response processing for greater speed and efficiency
    - espat: implement mqtt subscribe functionality via blocking select/channels (experiemental)
- **bugfixes**
    - st7789: fix index out of bounds error
- **examples**
    - Add espat driver example for mqtt subscribe

0.7.0
---
- **new devices**
    - veml6070: add Vishay UV light sensor
- **enhancements**
    - lis3dh: example uses I2C1 so requires config to specify pins since they are not default
    - ssd1331: make SPI TX faster
    - st7735: make SPI Tx faster
- **docs**
    - complete missing GoDocs for main and sub-packages
- **core**
    - add Version string for support purposes
- **examples**
    - Change all espat driver examples to use Arduino Nano33 IoT by default

0.6.0
---
- **new devices**
    - Support software SPI for APA102 (Itsy Bitsy M0 on-board "Dotstar" LED as example)

0.5.0
---
- **new devices**
    - LSM6DS3 accelerometer
- **bugfixes**
    - ws2812: fix timings for the nrf51
- **enhancements**
    - ws2812: Add build tag for Arduino Nano33 IoT

0.4.0
---
- **new devices**
    - SSD1331 TFT color display
    - ST7735 TFT color display
    - ST7789 TFT color display
- **docs**
    - espat
        - complete list of dependencies for flashing NINA-W102 as used in Arduino Nano33 IoT board.

0.3.0
---
- **new devices**
    - Buzzer for piezo or small speaker
    - PDM MEMS microphone support using I2S interface
- **enhancements**
    - epd2in13: added rotation
    - espat
        - add built-in support for MQTT publish using the Paho library packets, alongside some modifications needed for the AT protocol.
        - add DialTLS and Dial methods, update MQTT example to allow both MQTT and MQTTS connections
        - add example that uses MQTT publish to open server
        - add README with information on how to flash ESP32 or ESP8266 with AT command set firmware.
        - add ResolveUDPAddr and ResolveTCPAddr implementations using AT command for DNS lookup
        - change Response() method to use a passed-in timeout value instead of fixed pauses.
        - implement TCPConn using AT command set
        - improve error handling for key TCP functions
        - refactor net and tls interface compatible code into separate sub-packages
        - update MQTT example for greater stability
        - use only AT commands that work on both ESP8266 and ESP32
        - add documentation on how to use Arduino Nano33 IoT built-in WiFi NINA-W102 chip.
- **bugfixes**
    - core: Error strings should not be capitalized (unless beginning with proper nouns or acronyms) or end with punctuation, since they are usually printed following other context.
    - docs: add note to current/future contributors to please start by opening a GH issue to avoid duplication of efforts
    - examples: typo in package name of examples
    - mpu6050: properly scale the outputs of the accel/gyro

0.2.0
---
- **new devices**
    - AT24C32/64 2-wire serial EEPROM
    - BME280 humidity/pressure sensor
- **bugfixes**
    - ws2812: better support for nrf52832

0.1.0
---
- **first release**
    - This is the first official release of the TinyGo drivers repo, matching TinyGo 0.6.0. The following devices are supported:
        - ADXL345
        - APA102
        - BH1750
        - BlinkM
        - BMP180
        - DS1307
        - DS3231
        - Easystepper
        - ESP8266/ESP32
        - GPS
        - HUB75
        - LIS3DH
        - MAG3110
        - microbit LED matrix
        - MMA8653
        - MPU6050
        - PCD8544
        - SHT3x
        - SSD1306
        - Thermistor
        - VL53L1X
        - Waveshare 2.13"
        - Waveshare 2.13" (B & C)
        - WS2812
