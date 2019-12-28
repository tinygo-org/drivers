#!/bin/bash
cd /src/esp/nina-fw
export PATH=/src/esp/xtensa-esp32-elf/bin:$PATH
export IDF_PATH=/src/esp/esp-idf
make firmware
cp /src/esp/nina-fw/build/bootloader/bootloader.bin /src/build/
cp /src/esp/nina-fw/build/phy_init_data.bin /src/build/
cp /src/esp/nina-fw/build/nina-fw.bin /src/build/
cp /src/esp/nina-fw/build/partitions.bin /src/build/
cd -
