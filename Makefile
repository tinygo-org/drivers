
clean:
	@rm -rf build

FMT_PATHS = ./*.go ./examples/**/*.go

fmt-check:
	@unformatted=$$(gofmt -l $(FMT_PATHS)); [ -z "$$unformatted" ] && exit 0; echo "Unformatted:"; for fn in $$unformatted; do echo "  $$fn"; done; exit 1

smoke-test:
	@mkdir -p build
	tinygo build -size short -o ./build/test.elf -target=itsybitsy-m0 ./examples/adxl345/main.go
	tinygo build -size short -o ./build/test.elf -target=itsybitsy-m0 ./examples/apa102/main.go
	tinygo build -size short -o ./build/test.elf -target=microbit ./examples/at24cx/main.go
	tinygo build -size short -o ./build/test.elf -target=itsybitsy-m0 ./examples/bh1750/main.go
	tinygo build -size short -o ./build/test.elf -target=itsybitsy-m0 ./examples/blinkm/main.go
	tinygo build -size short -o ./build/test.elf -target=itsybitsy-m0 ./examples/bmp180/main.go
	tinygo build -size short -o ./build/test.elf -target=bluepill ./examples/ds1307/sram/main.go
	tinygo build -size short -o ./build/test.elf -target=bluepill ./examples/ds1307/time/main.go
	tinygo build -size short -o ./build/test.elf -target=itsybitsy-m0 ./examples/ds3231/main.go
	tinygo build -size short -o ./build/test.elf -target=microbit ./examples/easystepper/main.go
	tinygo build -size short -o ./build/test.elf -target=itsybitsy-m0 ./examples/espat/espconsole/main.go
	tinygo build -size short -o ./build/test.elf -target=itsybitsy-m0 ./examples/espat/esphub/main.go
	tinygo build -size short -o ./build/test.elf -target=itsybitsy-m0 ./examples/espat/espstation/main.go
	tinygo build -size short -o ./build/test.elf -target=feather-m0 ./examples/gps/i2c/main.go
	tinygo build -size short -o ./build/test.elf -target=feather-m0 ./examples/gps/uart/main.go
	tinygo build -size short -o ./build/test.elf -target=microbit ./examples/hd44780/customchar/main.go
	tinygo build -size short -o ./build/test.elf -target=microbit ./examples/hd44780/text/main.go
	tinygo build -size short -o ./build/test.elf -target=microbit ./examples/hub75/main.go
	tinygo build -size short -o ./build/test.elf -target=circuitplay-express ./examples/lis3dh/main.go
	tinygo build -size short -o ./build/test.elf -target=itsybitsy-m0 ./examples/mag3110/main.go
	tinygo build -size short -o ./build/test.elf -target=microbit ./examples/microbitmatrix/main.go
	tinygo build -size short -o ./build/test.elf -target=itsybitsy-m0 ./examples/mma8653/main.go
	tinygo build -size short -o ./build/test.elf -target=itsybitsy-m0 ./examples/mpu6050/main.go
	tinygo build -size short -o ./build/test.elf -target=microbit ./examples/pcd8544/setbuffer/main.go
	tinygo build -size short -o ./build/test.elf -target=microbit ./examples/pcd8544/setpixel/main.go
	tinygo build -size short -o ./build/test.elf -target=microbit ./examples/sht3x/main.go
	tinygo build -size short -o ./build/test.elf -target=microbit ./examples/ssd1306/i2c_128x32/main.go
	tinygo build -size short -o ./build/test.elf -target=microbit ./examples/ssd1306/spi_128x64/main.go
	tinygo build -size short -o ./build/test.elf -target=circuitplay-express ./examples/thermistor/main.go
	tinygo build -size short -o ./build/test.elf -target=itsybitsy-m0 ./examples/vl53l1x/main.go
	tinygo build -size short -o ./build/test.elf -target=microbit ./examples/waveshare-epd/epd2in13/main.go
	tinygo build -size short -o ./build/test.elf -target=microbit ./examples/waveshare-epd/epd2in13x/main.go
	tinygo build -size short -o ./build/test.elf -target=circuitplay-express ./examples/ws2812/main.go
	tinygo build -size short -o ./build/test.elf -target=trinket-m0 ./examples/bme280/main.go

test: clean fmt-check smoke-test
