# ESP-AT Driver

This package provides a driver to use a separate connected WiFi processor either the ESP8266 or the ESP32 from Espressif. 

The way this driver works is by using the UART interface to communicate with the WiFi chip using the Espressif AT command set.

## ESP-AT Firmware Installation

In order to use this driver, you must have the ESP-AT firmware installed on the ESP8266/ESP32 chip.

### Installing on Arduino Nano33 IoT

In order to install the needed firmware on the Arduino Nano33 IoT board's built-in NINA W102 chip, you will need to use the `arduino-nano33-iot` branch of this fork of the firmware:

https://github.com/hybridgroup/esp32-at

To flash this firmware on the Arduino Nano33 IoT you will need to follow the following procedure:

- Install _Arduino SAMD Boards_ from the Boards Manager.
- Install _WiFiNANO_ from the Library Manager.
- Using the normal Arduino software, load the `SerialNINAPassthrough` sketch on to the board (in File -> Examples -> WiFiNINA-> Tools).
- Flash the NINA 102 firmware using the `make flash` command in the https://github.com/hybridgroup/esp32-at repo.

You only need to do this one time, and then the correct ESP-AT firmware will be on the NINA chip, and you can just flash the Arduino Nano33 IoT board using TinyGo. We should be able to remove some of these step in a future release of this software.

### Installing on ESP32

The official repository for the ESP-AT for the ESP32 processor is located here:

https://github.com/espressif/esp32-at

Your best option is to follow the instructions in the official repo.

### Installing on ESP8266

The official repository for the AT command set firmware for the ESP8266 processor is located here:

https://github.com/espressif/ESP8266_NONOS_SDK

First clone the repo:

```shell
git clone https://github.com/espressif/ESP8266_NONOS_SDK.git
```

You will also need to install the Espressif `esptool` to flash this firmware on your ESP8266:

https://github.com/espressif/esptool

Once you have obtained the binary code, and installed `esptool`, you can flash the ESP8266.

Here is an example shell script that flashes a Wemos D1 Mini board:


```python
#!/bin/sh
SPToolDir="$HOME/.local/lib/python2.7/site-packages"
FirmwareDir="$HOME/Development/ESP8266_NONOS_SDK"
cd "$SPToolDir"
port=/dev/ttyUSB0
if [ ! -c $port ]; then
   port=/dev/ttyUSB0
fi
if [ ! -c $port ]; then
   echo "No device appears to be plugged in.  Stopping."
fi
printf "Writing AT firmware to the Wemos D1 Mini in 3..."
sleep 1; printf "2..."
sleep 1; printf "1..."
sleep 1; echo "done."
echo "Erasing the flash first"
esptool.py --port $port erase_flash
esptool.py --port /dev/ttyUSB0 --baud 115200 \
   write_flash -fm dio -ff 20m -fs detect \
   0x0000 "$FirmwareDir/bin/boot_v1.7.bin" \
   0x01000 "$FirmwareDir/bin/at/512+512/user1.1024.new.2.bin" \
   0x3fc000 "$FirmwareDir/bin/esp_init_data_default_v05.bin"  \
   0x7e000 "$FirmwareDir/bin/blank.bin" \
   0x3fe000 "$FirmwareDir/bin/blank.bin"

echo "Check the boot by typing: miniterm $port 74800"
echo " and then resetting.  Use Ctrl-] to quit miniterm."

```
