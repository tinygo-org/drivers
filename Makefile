
clean:
	@rm -rf build

FMT_PATHS = ./*.go ./examples/**/*.go

fmt-check:
	@unformatted=$$(gofmt -l $(FMT_PATHS)); [ -z "$$unformatted" ] && exit 0; echo "Unformatted:"; for fn in $$unformatted; do echo "  $$fn"; done; exit 1

./build/%.hex:
	@mkdir -p "$(@D)"
	tinygo build -size short -o $@ -target=$(notdir $(basename $@)) $(subst build,./examples,$(dir $@))
	@md5sum $@

EXAMPLES = $(dir $(shell find ./examples -type f -name 'main.go'))
EXAMPLE_TARGETS = $(shell cat $(example)targets.txt)
EXAMPLE_HEX_FILE = $(subst examples,build,$(example)$(target).hex)
EXAMPLE_HEX_FILES = $(foreach example,$(EXAMPLES),$(foreach target,$(EXAMPLE_TARGETS),$(EXAMPLE_HEX_FILE)))

smoke-test: $(EXAMPLE_HEX_FILES)

DRIVERS = $(wildcard */)
NOTESTS = build examples flash semihosting pcd8544 shiftregister st7789 microphone mcp3008 gps microbitmatrix \
		hcsr04 ssd1331 ws2812 thermistor apa102 easystepper ssd1351 ili9341 wifinina shifter hub75 \
		hd44780 buzzer ssd1306 espat l9110x st7735 bmi160 l293x dht keypad4x4 max72xx p1am tone tm1637 \
		pcf8563
TESTS = $(filter-out $(addsuffix /%,$(NOTESTS)),$(DRIVERS))

unit-test:
	@go test -v $(addprefix ./,$(TESTS)) 

test: clean fmt-check unit-test smoke-test
