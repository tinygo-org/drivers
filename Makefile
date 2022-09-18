
clean:
	@rm -rf build

FMT_PATHS = ./

fmt-check:
	@unformatted=$$(gofmt -l $(FMT_PATHS)); [ -z "$$unformatted" ] && exit 0; echo "Unformatted:"; for fn in $$unformatted; do echo "  $$fn"; done; exit 1

smoke-test:
	@mkdir -p build
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/adt7410/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/adxl345/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=pybadge ./examples/amg88xx
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/apa102/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=nano-33-ble ./examples/apds9960/proximity/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/apa102/itsybitsy-m0/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/at24cx/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/bh1750/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/blinkm/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/bmi160/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/bmp180/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/bmp280/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=trinket-m0 ./examples/bmp388/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=bluepill ./examples/ds1307/sram/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=bluepill ./examples/ds1307/time/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/ds3231/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/easystepper/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=arduino-nano33 ./examples/espat/espconsole/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=arduino-nano33 ./examples/espat/esphub/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=arduino-nano33 ./examples/espat/espstation/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/flash/console/spi
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=pyportal ./examples/flash/console/qspi
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/gc9a01/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=feather-m0 ./examples/gps/i2c/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=feather-m0 ./examples/gps/uart/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/hcsr04/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/hd44780/customchar/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/hd44780/text/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=arduino-nano33 ./examples/hd44780i2c/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=nano-33-ble ./examples/hts221/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/hub75/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=pyportal ./examples/ili9341/basic
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=xiao ./examples/ili9341/basic
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=pyportal ./examples/ili9341/pyportal_boing
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=pyportal ./examples/ili9341/scroll
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=xiao ./examples/ili9341/scroll
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=pyportal ./examples/ili9341/slideshow
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=circuitplay-express ./examples/lis3dh/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=nano-33-ble ./examples/lps22hb/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/lsm303agr/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=arduino-nano33 ./examples/lsm6ds3/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/mag3110/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/mcp23017/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/mcp23017-multiple/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/mcp3008/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/mcp2515/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/microbitmatrix/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/mma8653/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/mpu6050/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=p1am-100 ./examples/p1am/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=pico ./examples/pca9685/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/pcd8544/setbuffer/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/pcd8544/setpixel/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=arduino ./examples/servo
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=pybadge ./examples/shifter/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/sht3x/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/shtc3/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/ssd1306/i2c_128x32/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/ssd1306/spi_128x64/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/ssd1331/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/st7735/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/st7789/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=circuitplay-express ./examples/thermistor/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=circuitplay-bluefruit ./examples/tone
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=arduino-nano33 ./examples/tm1637/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=pyportal ./examples/touch/resistive/fourwire/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=pyportal ./examples/touch/resistive/pyportal_touchpaint/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/vl53l1x/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/vl6180x/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/waveshare-epd/epd2in13/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/waveshare-epd/epd2in13x/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/waveshare-epd/epd4in2/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=arduino-nano33 ./examples/wifinina/ntpclient/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=arduino-nano33 ./examples/wifinina/udpstation/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=arduino-nano33 ./examples/wifinina/tcpclient/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=arduino-nano33 ./examples/wifinina/webclient/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=circuitplay-express ./examples/ws2812
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.bin -target=m5stamp-c3          ./examples/ws2812
	@md5sum ./build/test.bin
	tinygo build -size short -o ./build/test.hex -target=feather-nrf52840 ./examples/is31fl3731/main.go
	@md5sum ./build/test.hex
ifneq ($(AVR), 0)
	tinygo build -size short -o ./build/test.hex -target=arduino   ./examples/ws2812
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=digispark ./examples/ws2812
	@md5sum ./build/test.hex
endif
	tinygo build -size short -o ./build/test.hex -target=trinket-m0 ./examples/bme280/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=circuitplay-express ./examples/microphone/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=circuitplay-express ./examples/buzzer/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=trinket-m0 ./examples/veml6070/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=arduino-nano33 ./examples/l293x/simple/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=arduino-nano33 ./examples/l293x/speed/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=arduino-nano33 ./examples/l9110x/simple/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=arduino-nano33 ./examples/l9110x/speed/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=nucleo-f103rb ./examples/shiftregister/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=hifive1b ./examples/ssd1351/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=circuitplay-express ./examples/lis2mdl/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=arduino-nano33 ./examples/max72xx/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=feather-m0 ./examples/dht/main.go
	@md5sum ./build/test.hex
	# tinygo build -size short -o ./build/test.hex -target=arduino ./examples/keypad4x4/main.go
	# @md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=xiao ./examples/pcf8563/alarm/
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=xiao ./examples/pcf8563/clkout/
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=xiao ./examples/pcf8563/time/
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=xiao ./examples/pcf8563/timer/
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=feather-m0 ./examples/ina260/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=nucleo-l432kc ./examples/aht20/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=feather-m4 ./examples/sdcard/console/
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=feather-m4 ./examples/sdcard/tinyfs/
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=wioterminal ./examples/rtl8720dn/webclient/
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=wioterminal ./examples/rtl8720dn/webserver/
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=wioterminal ./examples/rtl8720dn/mqttsub/
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=feather-m4 ./examples/i2csoft/adt7410/
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.elf -target=wioterminal ./examples/axp192/m5stack-core2-blinky/
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.uf2 -target=pico ./examples/xpt2046/main.go
	@md5sum ./build/test.uf2
	tinygo build -size short -o ./build/test.elf -target=m5stack-core2 ./examples/ft6336/basic/
	@md5sum ./build/test.elf
	tinygo build -size short -o ./build/test.elf -target=m5stack-core2 ./examples/ft6336/touchpaint/
	@md5sum ./build/test.elf
	tinygo build -size short -o ./build/test.hex -target=nucleo-wl55jc ./examples/sx126x/lora_rxtx/
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.uf2 -target=pico ./examples/ssd1289/main.go
	@md5sum ./build/test.uf2
	tinygo build -size short -o ./build/test.hex -target=pico ./examples/irremote/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=badger2040 ./examples/uc8151/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.uf2 -target=pico ./examples/scd4x/main.go
	@md5sum ./build/test.uf2
	tinygo build -size short -o ./build/test.uf2 -target=circuitplay-express ./examples/makeybutton/main.go
	@md5sum ./build/test.uf2

# rwildcard is a recursive version of $(wildcard) 
# https://blog.jgc.org/2011/07/gnu-make-recursive-wildcard-function.html
rwildcard=$(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) $(filter $(subst *,%,$2),$d))
# Recursively find all *_test.go files from cwd & reduce to unique dir names
HAS_TESTS = $(sort $(dir $(call rwildcard,,*_test.go)))
# Exclude anything we explicitly don't want to test for whatever reason
EXCLUDE_TESTS = image
TESTS = $(filter-out $(addsuffix /%,$(EXCLUDE_TESTS)),$(HAS_TESTS))

unit-test:
	@go test -v $(addprefix ./,$(TESTS))

test: clean fmt-check unit-test smoke-test
