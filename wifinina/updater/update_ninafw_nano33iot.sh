#!/bin/bash

# flash the passthru binary to the SAMD21 using bossac
# code from https://github.com/arduino-libraries/WiFiNINA/blob/master/examples/Tools/SerialNINAPassthrough/SerialNINAPassthrough.ino
stty -F /dev/ttyACM0 ispeed 1200 ospeed 1200
bossac -d -i -e -w -v -R --port=/dev/ttyACM0 --offset=0x2000 ./SerialNINAPassthrough.ino.nano_33_iot.bin

# download the nina-fw binary
wget https://github.com/arduino/nina-fw/releases/download/1.4.8/NINA_W102-v1.4.8.bin 

# flash the nina-fw binary to the ESP32 using esptool
esptool --port /dev/ttyACM0 --before default_reset --baud 115200 write_flash 0 ./NINA_W102-v1.4.8.bin

