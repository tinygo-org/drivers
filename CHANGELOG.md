0.28.0
---
- **new devices**
    - **epd2in66b**
        - Waveshare 2.66inch E-Paper Display Module (B) for Raspberry Pi Pico (#673)
    - **mcp9808**
        - Add driver for MCP9808 i2c temperature sensor (#676)

- **enhancements**
    - **encoders**
        - add atsamd21, atsamd51, atsame5x
    - **pixel**
        - add support for Monochrome types such as the SSD1306 display
    - **rtl8720dn**
        - implement ConnectModeAP
    - **servo**
        - add function SetAngle() to simplify API for most common use case
    - **ssd1306**
        - add DrawBitmap() function to complete Displayer interface
        - add rotation functions for Displayer interface
        - add Sleep() function for Displayer interface
    - **uc8151**
        - improvements to speed and also add flicker-free mode based on @antirez code example
        - update to support all functions needed by tinygl and board package Displayer interface
    - **wifinina**
        - implement ConnectModeAP

- **bugfixes**
    - **ft6336**
        - ignore bogus touch events
    - **pixel**
        - fix Image[Monochrome].Set for larger images
    - **uc8151**
        - correct DrawBitmap() also refactor SendCommand() and SendData() for clarity
    - **ws2812**
        - Fix typo and move initialization of neo to init()

- **examples**
    - **ws2812**
        - Simplify examples/ws2812


0.27.0
---
- **core**
    - prepare for CGo changes in TinyGo

- **new devices**
    - **adafruit4650**
        - support for Adafruit 4650 feather OLED
    - **net**
        - new networking support based on tinygo net package
    - **pixel**
        - add package for efficiently working with raw pixel buffers
    - **rotary**
        - Adding driver for rotary encoder support
    - **seesaw**
        - Adding support for Adafruit Seesaw platform
    - **sgp30**
        - add SGP30 air quality sensor
    - **sk6812**
        - added support for SK6812 to WS2812 device (#610)

- **enhancements**
    - **epd2in13**
        - add Sleep method like other displays
        - unify rotation configuration with other displays
        - use better black/white approximation
    - **ili9341**
        - add DrawBitmap method
    - **lora/lorawan**
        - LoRa WAN US915 Support
        - LoRa WAN add setter functions
        - refactor shared functionality for channels/regions
    - **mcp2515**
        - Add more line speeds to mcp2515.go (#626)
    - **rtl8720dn**
        - use drivers package version as the driver version
    - **ssd1306**
        - improvements needed for Thumby SPI display
    - **st7735**
        - make the display generic over RGB565 and RGB444
    - **st7789**
        - add DrawBitmap method
        - make the display generic over RGB565 and RGB444
    - **wifinina**
        - add ResetIsHigh cfg switch for MKR 1010 (copied from #561)
        - maintenence. Also see PR #4085 in the main TinyGo repo
        - use drivers package version as the driver version

- **bugfixes**
    - **adxl345**
        - Use int16 for ADXL345 readings (#656)
    - **at24cx**
        - fixed the description of the device struct
    - **rtl8720dn**
        - allow connecting to open wifi access points
        - fix check for bad Wifi connect
    - **sh1106**
        - fix I2C interface and add smoketest
        - fixed the description of the device struct
    - **wifinina**
        - add 'unknown failure' reason code for AP connect
        - fix concurrency issues with multiple sockets
        - fix wifinina UDP send
 
- **examples**
    - **ds3231**
        - fix the description in the example
    - **lorawan**
        - add missing functions for simulated interface
        - modify atcmd and basic demo to support choosing any one of the supported regions at compile time by using ldflags
    - **net**
        - all networking examples now using netdev and netlink.

- **build**
    - **all**
        - fix broken testrunner
        - migrated legacy I2C
        - add natiu package for tests
    - **smoketest**
        - add stack-size param for net tests.
        - allow stack-size flag as it is needed for net examples


0.26.0
---
- **core**
    - i2c iface refactor: Resolve #559
    - fix uses of legacy i2c WriteRegister calls
    - add correct Tx implementation for mock I2C interfaces
    - bump golang.org/x/net version

- **new devices**
    - **bma42x**
        - add new BMA421/BMA425 driver
    - **ndir**
        - add Sandbox Electronics NDIR CO2 sensor driver (#580)
    - **mpu9150**
        - implement driver for Mpu9150 (#596)
    - **sht4x**
        - implement driver for sht4x (#597)
    - **pcf8523**
        - implement driver for pcf8523 (#599)

- **enhancements**
    - **ssd1306**
        - improve bus error handling

- **bugfixes**
    - **st7789**
        - fix scrolling when rotated by 180°
    - **st7789**
        - fix incorrect Rotation configuration
        - fix SetScrollArea
    - **ili9341**
        - fix SetScrollArea

- **build**
    - use latest tag of tinygo-dev container for running tests


0.25.0
---

- **core**
    - add Sensor interface and Measurement type
    - **delay**
        - add new package for cycle-accurate delays

- **new devices**
    - **AS560x**
        - Add support for ams AS560x on-axis magnetic rotary position sensors
    - **onewire**
        - first implementation of 1-wire protocol (#505)
    - **mpu6886**
        - initial implementation
    - **ttp229**
        - initial support for ttp229 (BSF)

- **enhancements**
    - **gps**
        - make the date available in addition to the time (#532)
    - **i2csoft**
        - use cycle counting for delays
    - **ili9341**
        - add EnableTEOutput to be able to sync drawing with VSYNC
        - add sleep mode
        - unify rotation support
    - **st7735**
        - add DrawRGBBitmap8 method to draw raw RGB565 buffers
        - add sleep mode
        - unify rotation support
    - **st7789**
        - added DrawRGBBitmap8 (same as ili9341 & st7735)
        - allow changing the color format using COLMOD
        - make it possible to configure gamma
        - support the chip select pin
        - update saved rotation in SetRotation
        - add sleep mode
        - unify rotation support
    - **sx126x/sx127x**
        - Reduce spi buffer size, add missing select when using channels
        - Remove heap alloc in interrupt, add non blocking channel send/receive, and other cleanups
    - **wifinina**
        - add generated strings, improved debugging system and messages
        - add ResetIsHigh to control the behavior of the RESET pin for boards like the Arduino MKR 1010
        - only add generated strings when using wifidebug tag

- **bugfixes**
    - **ds3231**
        - Document incorrect leap year 2100
        - Fix negative temperature conversion
    - **ili9341**
        - fix Size() for mirrored rotation
    - **st7789**
        - avoid heap allocations after the driver is created
    - **net**
        - Revert "(#501) make IP.String() method return something sensible"
    - **wifinina**
        - small timing adjustments in Configure() to better ensure device reset

- **examples**
    - **sdcard**
        - remove tinyfs example and replace with link to tinyfs repo in docs
    - **wifinina**
        - improve connectToAP() and other needed minor corrections

- **build**
    - switch to ghcr.io for docker container
    - run smoke tests in parallel
    - **Makefile**
        - add XTENSA=0 flag to skip Xtensa tests
        - remove AVR=0 flag

- **docs**
    - remove full list of devices from README, better to keep it on the tinygo.org site
    - update LICENSE year


0.24.0
---
- **new devices**
    - **lora**
        - created shared RadioEvent
        - move shared config for sx126x/sx127x to single package
    - **lorawan**
        - add initial LoRaWAN stack support
        - Basic implementation of Lorawan Regional Settings and EU868/AU915 regions
    - **qmi8658c**
        - Add support for the QMI8658C sensor (#467)
    - **sh1106**
        - add support for SH1106 display driver
    - **sx127x**
        - Driver for Semtech sx127x radio modules

- **enhancements**
    - **bme280**
        - improve config support
        - add ReadAltitude() function copied from BMP280 driver
    - **buzzer**
        - make all note durations float64
        - no tone during rest
    - **dht22**
        - update DHT22 receive to use runtime/interrupt
    - **gps**
        - add support for GLL sentence type, add original sentence to gps errors
        - improve error handling
        - improve parsing and add tests to verify
    - **microbitmatrix**
        - add link to schema for microbit V2
        - add smoke test for microbitmatrix with microbit-v2
        - add support for brightness of led pixels
        - harmonize v1 and v2 implementation
        - move Size() to version agnostic part
    - **mpu6050**
        - add functions to configure clock, and scaling for accelerometer and gyroscope
    - **net/http**
        - add PostForm()
    - **sx126x**
        - add Reset() and needed pin
        - move RadioController into separate file for clarity
        - pre-define all errors to avoid heap allocations
        - refactor to RadioController interface to more easily handle non-STM32WL boards and remove duplicated code

    - **vl53l1x**
        - Add getter for the effective SPAD count
    - **wifinina**
        - add support for http server (#480)

- **bugfixes**
    - **lsm303agr**
        - fix I2C address auto increment for multi data read
    - **net**
        - (#501) make IP.String() method return something sensible
    - **mpu6050**
        - return I2C error when configuring fails
    - **sx126x**
        - fix in SetBandwidth function
        - actually set the frequency when calling SetFrequency()
        - correct RX/TX pin mapping for TheThingsIndustries GNSE board

- **examples**
    - **LoRaWAN**
        - example with LoRaWAN AT command set implementation
        - basic example
        - update all remaining examples for refactored API
    - **sx126x**
        - fix bandwidth,tx power in lora//lora_continuous example
    - **sx127x**
        - rx/tx example

- **build**
    - remove older format build tags
    - update to actions/checkout@v3
    - work around for CVE-2022-24765


0.23.0
---
- **new devices**
    - Add GC9A01 Driver (#452)
    - Initial support for VL6180x sensor

- **enhancements**
    - **rtl8720dn**
        - refactor by bringing this driver more in line with wifinina and espat
    - **ssd1306**
        - add getter method to the current buffer
    - **makeybutton**
        - revise to better match the algorithm defined by the original
    - **espat,wifinina,rtl8720dn**
        - change ssid/pass from const to var

- **bugfixes**
    - **microbitmatrix**
        - fix inverted axis
    - **espat**
        - Trim quotes from IP returned by espat's GetDNS()
    - **all**
        - correct go fmt

- **examples**
    - **rtl8720dn**
        - remove wifi setting
        - add call to optional debug setting
        - update all remaining examples for refactored API


0.22.0
---
- **new devices**
    - epd: add waveshare 2.9in (v1)
    - makeybutton: add driver for MakeyMakey-like button

- **enhancements**
    - **rtl8720dn**
        - add UDP close function
        - improve error handling
    - **net/http**
        - improve header parsing
        - add last-will-and-testament to MQTT
    - **net/mqtt**
        - adds keepalive pinging, disconnect, and graceful goroutine cleanup
        - support for cookies when https
        - add support for retained messsages

- **bugfixes**
    - irremote: Fix irremote reporting incorrect NEC addresses and command codes (#422)
    - net/http:  Fix http.Get() with port specification

- **build**
    - Makefile recursively finds unit-tests
    - switching to GHA

- **updates**
    - update tinyfont to v0.3.0
    - update tinyfs to v0.2.0

- **examples**
    - rtl8720dn: add ./examples/rtl8720dn/version


0.21.0
---
- **new devices**
    - lsm6ds3tr: initial implementation
    - UC8151: used in Pimoroni's badger2040 e-paper (#416)
    - scd4x: implement driver for CO2 sensor

- **enhancements**
    - easystepper: Add support for '8-step mode'
    - vl53l1x: Add functions for setting the device address
    - sdcard: support thingplus-rp2040
    - wifinina: add mutex to prevent communication race problems
    - **ws2812**
        - support thingplus-rp2040 board
        - Added 125 MHz rp2040 timing
        - Added unsafe.Pointer for pointer conversion

- **bugfixes**
    - ssd1351: Fix mirrored text on OLED display


0.20.0
---
- **new devices**
    - irremote: Add basic infra-red driver
    - IS31FL3731: add driver for IS31FL3731 matrix LED driver (#370)
    - l3gd20: add gyro driver
    - SSD1289: Driver for SSD1289 LCD

- **enhancements**
    - **ili9341**
        - add support for atsame5x
        - added Feather board support to InitDisplay()
        - avoid heap allocations
    - **lps22hb**
        - pin rename, sync with main repo
    - **lsmXXX**
        - unified, error handling, memory management
    - **max7xx**
        - Add a SetIntensity() function to max7xx driver and example
    - **vl53l1x**
        - Add functions for setting 'region of interest'
        - Fix switch-case semantics
    - **ws2812**
        - add support for m5stamp-c3
        - convert AVR assembly to C inline assembly
        - support high-MHz ARMv6M chips like the RP2040
        - write inline assembly using C instead of Go

- **bugfixes**
    - **dht**
        - fix error check in example
        - fix humidity and temperature extraction for DHT22 (#358)
    - **esp8266**
        - fix ConnectToAccessPoint timeout args
    - **image**
        - fix interface
    - **pca9685**
        - add buffered one shot write
        - fix on=0 bug
    - **wifinina**
        - correct sendParamStr to handle empty strings, such as when connecting to an unsecured access point

0.19.0
---
- **new devices**
    - ft6336: add support for ft6336
    - pca9685: PCA9685 driver
    - shtc3: Sensirion SHTC3 Relative Humidity / Temperature i2c sensor
    - sx126x: Driver for Semtech sx126x radio modules
    - xpt2046: XPT2046 Touch driver (#350)
- **enhancements**
    - **hd44780i2c** 
        - clean up for go fmt
        - Needed fixes and update hd44780i2c.go
    - **ili9341, ili9342**
        - add support for m5stack
        - add support for m5stack-core2
    - **wifi**
        - modify to use shared net.Adapter interface for all supported wifi devices
    - wifinina: remove busy wait
- **bugfixes**
    - **hd44780** 
        - fix 4-bit data length flag
        - Reset data pins to output mode after reading
    - Nano 33 BLE drivers (#351)
- **docs**
    - examples/wifi: add unified example for tcpclient that compiles for all supported wifi adaptors

0.18.0
---
- **new devices**
    - apds9960: add support for APDS-9960 Digital Proximity sensor
    - axp192: add support for AXP192 single Cell Li-Battery and power system management IC
    - hts221: add support for HTS221 capacitive digital sensor for relative humidity and temperature
    - i2csoft: add support for software I2C
    - image: add support for image/jpeg and image/png
    - lps22hb: add support for LPS22HB MEMS nano pressure sensor
    - lsm6dox: add support for lsm6dox accelerometer
    - lsm9ds1: add support for lsm9ds1 accelerometer
- **enhancements**
    - ili9341: change to use drivers.SPI interface
    - **ws2812**
        - generate assembly instead of handwriting it
        - improve timings to be compatible with the WS2811
        - add support for 168MHz (e.g. Adafruit Feather STM32F405)
        - add support for RISC-V
    - wifinina: control nina pins, for example leds
- **docs**
    - rtl8720dn: examples for tcpclient, udpstation, mqtt, and webserver
    - **wifinina** 
        - nina-fw update docs
        - examples/wifinina/http-get
    - ili9341: refactor examples
    - Fix broken link for SHT3x datasheet
- **core**
    - all: use build directives for both Go1.17 and earlier versions
- **bugfixes**
    - net: fix raddr of tcp conn
    - mcp3008: fix bitshift bug

0.17.1
---
- To correct an error in the release process. Same as 0.17.0.

0.17.0
---
- **new devices**
    - rtl8720dn: add support for rtl8720dn
    - sdcard: add support for spi sdcard driver, along with fatfs
- **enhancements**
    - apa102: use 4-byte buffer to improve speed
    - bmi160: avoid heap allocations
    - ili9341: add standard SPI driver
    - wifinina
        - avoid fmt package
        - Fix RSSI command for WiFiNINA + Print current SSID + Wait for correct time before printing it out + Cleanup
    - ws2812
        - rename the pin to ws2812
        - add tag for nrf52833
        - Disable interrupts before sending ws2812 data
        - add support for qtpy and atsame5x
- **core**
    - modules: switch to use tinygo-org version of tinyfs package
    - all: use machine.Serial as the default output

0.16.0
---
- **new devices**
    - aht20: add device
    - ina260: add new i2c device
    - keypad: add 4x4 keypad driver (#226)
    - max7219: add driver support
    - mcp2515: add support for mcp2515 CAN device
    - p1am: support the P1AM-100 hardware watchdog
    - pcf8563: add support for pcf8563 real time clock
    - servo: add driver using PWM
    - tm1637: add support for tm1637 7-segment LED
    - tone: add package for producing tones using the PWM interface
- **enhancements**
    - pwm: update drivers with PWM to use new interface
    - wifinina: Make TLS work over WiFiNINA Verified on Arduino Nano33 IoT and nina fw v1.4.5
    - ssd1306: Enable reset screen for SSD1306 via I2C
    - st7789: add scrolling functions to match st7735
- **bugfixes**
    - wifinina:
        - fix getMACAddress and getTime
        - fix println + cleanup
        - remove debug flag and remove unnecessary padding call
        - fix padding and implement missing functions
    - flash: fix EraseBlocks method which is erasing sectors instead
- **core**
    - all: use interfaces for UART objects
    - all: do not take the pointer of an I2C object
    - adc: update drivers with ADC to use new config struct
- **testing**
    - tester:
        - add a mock for command-oriented i2c devices
        - add 16-bit register mock device

- **docs**
    - ssd1306: example of ssd1306 with 128x64 display over I2C
    - wifinina:
        - add information about Adafruit boards with ESP32 wifi coprocessors, and modify examples to remove code that was both not being used, and also prevented many Adafruit boards from being able to be targeted by the examples
        - update docs to simplify the nina-fw update process
        - example that connects to AP and prints ip addresses, time and mac
    - p1am: documentation and example program
    - add missing new drivers added since last release

0.15.0
---
- **new devices**
    - dht: add DHTXX thermometer
    - mcp23017: new driver for MCP23017 (I2C port expander)
    - bmp388: Add bmp388 support (#219)
- **enhancements**
    - hd44780: add a mode to work with boards where the RW pin is grounded
    - st7789: add scrolling functions to match st7735
    - microbitmatrix: matrix now working on microbit v2
    - ds1307: Better interface "ReadTime" instead of "Time"
    - ws2812: make AVR support more robust
- **bugfixes**
    - all: fix main package in examples
- **core**
    - adc: update all drivers with ADC to use new config struct
    - spi: remove machine.SPI and replace with drivers.SPI interface for almost all SPI drivers
- **testing**
    - test: run unit tests against i2c drivers and any spi drivers without direct gpio
- **docs**
    - st7789: correct errors on various godoc comments

0.14.0
---
- **new devices**
    - lis2mdl: add LIS2MDL magnetometer (#187)
    - waveshare: add Waveshare 4.2in B/W e-paper driver (#183)
- **enhancements**
    - adt7410: add connection test and for that matter connection method
    - gps
        - add speed and heading to fix, as parsed from RMC NMEA sentence
        - improvements and bugfixes (#186)
    - ili9341
        - add support for setting framerate, vsync pause, and reading scanline data.
        - renamed NewSpi() to NewSPI() in accordance with Go naming conventions
    - ws2812
        - add support for ESP8266
        - add support for ESP32
- **bugfixes**
    - ili9341
        - rix setWindow bug, add CS pin for Clue compatibility. (#180)
        - bugfix for RAMWR bug
    - lis2mdl: turn on read mode on every read, to ensure that magnetometer data is updated
- **core**
    - i2c
        - switch all i2c drivers definitions to use i2c bus interface type instead of machine package concrete type
        - correct interface definition for I2C Tx function
- **testing**
    - fix smoke-test unless avr-gcc installed
    - add very basic mock structs for testing i2c devices, based on work done by @rogpeppe
    - improve API surface and implement one more test function in lis2mdl driver
- **docs**
    - replace README badge for godocs with pkgdocs

0.13.0
---
- **new devices**
    - bmi160: add initial support
    - bmp280: added support for the Bosch BMP280 temperature and pressure sensor. (#158)
    - lsm303agr: add lsm303agr (#162)
    - ssd1351: add SSD1351 OLED display driver (#146)
- **enhancements**
    - hd44780: add Hd44780i2c driver (#173)
    - ili9341
        - add ILI9341 TFT driver (SPI) for ATSAMD2x (#174)
        - cache address window to prevent sending unnecessary commands (#171)
        - ILI9341 TFT driver (SPI) (#153)
        - improve performance of ILI9341 on ATSAMD5X
    - ST77xx: fix DrawFastHLine for ST77xx, SSD1331 and SSD1351 DrawFastHLine uses FillRectangle(x,y,width,height,c), so height must be 1 to draw a horizontal line
    - tmp102: add Connected func to check for device
    - wifinina: added UDP support
    - ws2812: update ws2812_avr_16m.go
- **bugfixes**
    - apa102: avoid creating garbage
    - bmp180: fix temperature type conversion
- **core**
    - all
        - added custom import path (#161)
        - changeover to eliminate all direct use of master/slave terminology
    - build: try vendor in working directory to match expected module path
    - ci: support Go modules
    - modules: update go version and dependency
- **docs**
    - docs: reorder to correct alpha and adjust count of supported drivers

0.12.0
---
- **new devices**
    - hcsr04: Added HC-SR04 ultrasonic distance sensor. (#143)
    - spi/qspi: Low-level IO driver for serial flash memory via SPI and QSPI (#124)
    - tmp102: TMP102 low-power digital temperature sensor (#141)
    - amg88xx: AMG88xx thermal camera module
- **bugfixes**
    - mqtt: reduce use of goroutines in router to not start a new goroutine for each invocation of each callback

0.11.0
---
- **new devices**
    - shiftregister: Support for various shift register chips (#135)
- **enhancements**
    - shifter: simplify API surface for PyBadge (#137)
    - shifter: new API for shifter driver
    - mqtt: use buffered channels for incoming messages to handle bursts
    - ili9341: Adding scroll functionality (#121)
- **bugfixes**
    - wifinina: fix typo on StartScanNetworks
    - ili9341: various bugfixes for display
- **examples**
    - semihosting: add example
- **docs**
    - readme: Use degree sign instead of ordinal
    - all: fix celsius symbol in all code comments

0.10.0
---
- **new devices**
    - adt7410: Support for ADT7410 temperature sensor (#109)
    - ili9341: ILI9341 TFT driver (#115)
    - l293x: added support for h-bridge motor controller
    - l9110x: add support for L9110x h-bridge motor driver
    - resistive: Adding driver for four-wire resistive touchscreen (#118)
- **enhancements**
    - st7735: added scroll functionality to st7735
    - st7735: remove default offsets
    - st7789: remove default offsets
    - ws2812: Added nrf52840 tag to ws2812
    - ws2812: work-arounds to allow Digispark to control WS2812 LEDs
- **docs**
    - readme: update README to include list of all 44 drivers
    - wifinina: update docs and add Dockerfile to build firmware
    - wifinina: update docs and info on how to install WiFiNINA driver

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
