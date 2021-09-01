# WifiNINA Driver

This package provides a driver to use a separate connected WiFi processor ESP32 for TCP/UDP communication.

The way this driver works is by using the SPI interface of your microcontroller to communicate with the WiFi chip using the Arduino SPI command set.

## Using the WiFiNINA Driver

For information on how to use this driver, please take a look at the examples located in the [examples/wifinina](../examples/wifinina) directory.

## Firmware

**PLEASE NOTE: New Adafruit Boards with WiFi and Arduino Nano33 IoT and Nano RP2040 Connect boards most likely already have a recent version of the nina-fw firmware pre-installed. You should not need to install the firmware yourself.**

In order to use this driver, you must have the [nina-fw firmware](https://github.com/arduino/nina-fw/) installed on the ESP32 chip. If it is already installed, you can just use it. You do not need to flash the firmware again unless it has known bugs or lacks functionality you plan to use.

The following instructions are only for those who want or need to update the firmware on your board.

### Update Arduino Boards with Arduino IDE

Probably, the easiest way to update nina-fw on Arduino boards is to use their IDE.
Please see [tutorial page](https://www.arduino.cc/en/Tutorial/WiFiNINA-FirmwareUpdater) for detailed instructions.

Sometimes you may want to flash a version of nina-fw Arduino IDE does not know about yet.  
There is no way to "refresh" firmware versions list, but you can "fool" Arduino IDE by substituting, say, 1.4.5 binary with 1.4.8 binary.  
For that you need to locate the binary Arduino IDE going to flash and replace it with a new file.  
Location differs from OS to OS, on macOS it's in "/Applications/Arduino.app/Contents/Java/tools/WiFi101/tool/firmwares/NINA/".  
Replace respective binary ".bin" file with new file and follow tutorial steps above flashing "1.4.5" version while in fact it is going to flash "1.4.8".

Latest firmware binary file for your board can be downloaded from their [releases page](https://github.com/arduino/nina-fw/releases).

Verify correct version flashed with "Examples/WiFiNINA/Tools/CheckFirmwareVersion" sketch in Arduino IDE.

### Update Arduino Boards without Arduino IDE
#### Install esptool to flash nina-fw firmware

In order to flash the firmware, you need to use Python to install the `esptool` package.

```shell
pip install esptool
```

On macOS you can also use brew

```shell
brew install esptool
```

Once you have installed `esptool` you can follow the correct procedure for flashing your board.

#### Update nina-fw on the Arduino Nano33 IoT

In the `updater` directory we have a precompiled binary of the "passthrough" code you will need to flash first, in order to update the ESP32 co-processor on your board.

This is what needs to be done. There is also a bash script that performs the same steps also located in the `updater` directory.

```shell
mkdir -p ../build

# reset board into bootloader mode using 1200 baud
stty -F /dev/ttyACM0 ispeed 1200 ospeed 1200

# flash the passthru binary to the SAMD21 using bossac
# code from https://github.com/arduino-libraries/WiFiNINA/blob/master/examples/Tools/SerialNINAPassthrough/SerialNINAPassthrough.ino
bossac -d -i -e -w -v -R --port=/dev/ttyACM0 --offset=0x2000 ./SerialNINAPassthrough.ino.nano_33_iot.bin

# download the nina-fw binary
wget -P ../build/ https://github.com/arduino/nina-fw/releases/download/1.4.8/NINA_W102-v1.4.8.bin 

# flash the nina-fw binary to the ESP32 using esptool
esptool --port /dev/ttyACM0 --before default_reset --baud 115200 write_flash 0 ../build/NINA_W102-v1.4.8.bin
```

You only need to do this one time, and then the correct nina-fw firmware will be on the NINA ESP32 chip, and you can just flash the Arduino Nano33 IoT board using TinyGo.

### Update Adafruit ESP32 WiFi Boards

Adafruit provides very good instructions for updating their boards that provide a ESP32 WiFi-BLE co-processor. For more information, please see:

https://learn.adafruit.com/upgrading-esp32-firmware
