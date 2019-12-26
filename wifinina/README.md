# WifiNINA Driver

This package provides a driver to use a separate connected WiFi processor ESP32 for TCP/UDP communication.

The way this driver works is by using the SPI interface of your microcontroller to communicate with the WiFi chip using the Arduino SPI command set.

## WiFiNINA Firmware Installation

In order to use this driver, you must have the WiFiNINA firmware installed on the ESP32 chip.

### Installing on Arduino Nano33 IoT

In order to install the needed firmware on the Arduino Nano33 IoT board's built-in NINA W102 chip, you will need to use this firmware:

https://github.com/arduino/nina-fw

To flash this firmware on the Arduino Nano33 IoT you will need to follow the following procedure:

- Install _Arduino SAMD Boards_ from the Boards Manager.
- Install _WiFiNINA_ from the Library Manager.
- Using the normal Arduino software, load the `SerialNINAPassthrough` sketch on to the board (in File -> Examples -> WiFiNINA-> Tools).
- Build the NINA 102 firmware using the `make firmware` command in the https://github.com/arduino/nina-fwt repo.
- Flash the WifiNINA firmware using the `esptool` script:

    ```shell
    python esptool.py --chip esp32 --port /dev/ttyACM0 --baud 115200 --before no_reset --after hard_reset write_flash -z --flash_mode dio --flash_freq 40m --flash_size detect 0x1000 /home/ron/Development/nina-fw/build/bootloader/bootloader.bin 0xf000 /home/ron/Development/nina-fw/build/phy_init_data.bin 0x30000 /home/ron/Development/nina-fw/build/nina-fw.bin 0x8000 /home/ron/Development/nina-fw/build/partitions.bin
    ```

You only need to do this one time, and then the correct WiFiNINA firmware will be on the NINA chip, and you can just flash the Arduino Nano33 IoT board using TinyGo. We should be able to remove some of these steps in a future release of this software.
