
clean:
	@rm -rf build

FMT_PATHS = ./*.go ./examples/**/*.go

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
	tinygo build -size short -o ./build/test.hex -target=circuitplay-express ./examples/lis3dh/main.go
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
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/microbitmatrix/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/mma8653/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/mpu6050/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/pcd8544/setbuffer/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/pcd8544/setpixel/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=pybadge ./examples/shifter/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=microbit ./examples/sht3x/main.go
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
	tinygo build -size short -o ./build/test.hex -target=pyportal ./examples/touch/resistive/fourwire/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=pyportal ./examples/touch/resistive/pyportal_touchpaint/main.go
	@md5sum ./build/test.hex
	tinygo build -size short -o ./build/test.hex -target=itsybitsy-m0 ./examples/vl53l1x/main.go
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
	tinygo build -size short -o ./build/test.hex -target=feather-m0 ./examples/dht/main.go
	@md5sum ./build/test.hex

DRIVERS = $(wildcard */)
NOTESTS = build examples flash semihosting pcd8544 shiftregister st7789 microphone mcp3008 gps microbitmatrix \
		hcsr04 ssd1331 ws2812 thermistor apa102 easystepper ssd1351 ili9341 wifinina shifter hub75 \
		hd44780 buzzer ssd1306 espat l9110x st7735 bmi160 l293x dht
TESTS = $(filter-out $(addsuffix /%,$(NOTESTS)),$(DRIVERS))

unit-test:
	@go test -v $(addprefix ./,$(TESTS)) 

test: clean fmt-check unit-test smoke-test
