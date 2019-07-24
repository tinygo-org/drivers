#!/bin/sh

# Exit on the first error.
set -e

# Flash the pass-through firmware.
# The binary has been created following the procedure over at
# https://github.com/tinygo-org/drivers/tree/master/espat.
# Note: this has to be an up-to-date bossa tool, you can get get it from here:
# https://github.com/shumatech/BOSSA.git
# Commit 8202074d53ba666a7bbe9def780a9a9f78a4b140 at 2019-06-03 is known to
# work.
echo "Flashing pass-through firmware..."
bossac -i -d --port=ttyACM0 -e -w -v --offset=0x2000 -R SerialNINAPassthrough.ino.bin

echo "Waiting for a bit..."
sleep 1

echo "Flashing firmware to ESP32..."
python $HOME/src/esp-idf/components/esptool_py/esptool/esptool.py --chip esp32 --port /dev/ttyACM0 --baud 921600 --before no_reset --after hard_reset write_flash -z --flash_mode dio --flash_freq 40m --flash_size detect 0x10000 ota_data_initial.bin 0x1000 bootloader.bin 0x20000 at_customize.bin 0x24000 server_cert.bin 0x26000 server_key.bin 0x28000 server_ca.bin 0x2a000 client_cert.bin 0x2c000 client_key.bin 0x2e000 client_ca.bin 0x30000 factory_param.bin 0xf000 phy_init_data.bin 0x100000 esp-at.bin 0x8000 partitions_at.bin
