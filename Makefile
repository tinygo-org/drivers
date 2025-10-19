# sub-directory used for build
BUILD_DIR := build
# find all packages without any dependency to underlying tiny-go packages, e.g. "machine" or "device/arm"
BUILD_TAGS_CHECK := m5stack_core2,microbit,xiao_ble
ALL_WITHOUT_MACHINE := $(shell go list -e -tags $(BUILD_TAGS_CHECK) -f '{{.Dir}},{{.Deps}}' ./... | awk -F, '$$2 !~ /machine/ && $$2 !~ /device\/arm/ {print $$1}')
# exclude anything found in build output and "image" directory, and exclude some further folders, which contains problematic dependencies in sub-folders
EXCLUDE_PACKAGES = $(CURDIR)/$(BUILD_DIR)/% $(CURDIR)/image/% $(CURDIR)/touch $(CURDIR)/waveshare-epd
ALL_TO_CHECK := $(filter-out $(EXCLUDE_PACKAGES),$(ALL_WITHOUT_MACHINE))

.PHONY: clean fmt-check smoke-test unit-test test check fmt_check fmt_fix $(ALL_TO_CHECK)

clean:
	@rm -rf $(BUILD_DIR)

FMT_PATHS = ./

fmt-check:
	@unformatted=$$(gofmt -l $(FMT_PATHS)); [ -z "$$unformatted" ] && exit 0; echo "Unformatted:"; for fn in $$unformatted; do echo "  $$fn"; done; exit 1

XTENSA ?= 1
smoke-test:
	@mkdir -p $(BUILD_DIR)
	@go run ./smoketest.go -xtensa=$(XTENSA) smoketest.sh


# rwildcard is a recursive version of $(wildcard) 
# https://blog.jgc.org/2011/07/gnu-make-recursive-wildcard-function.html
rwildcard=$(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) $(filter $(subst *,%,$2),$d))
# Recursively find all *_test.go files from cwd & reduce to unique dir names
HAS_TESTS = $(sort $(dir $(call rwildcard,,*_test.go)))
# Exclude anything we explicitly don't want to test for whatever reason
EXCLUDE_TESTS = image waveshare-epd/epd2in66b
TESTS = $(filter-out $(addsuffix /%,$(EXCLUDE_TESTS)),$(HAS_TESTS))

unit-test:
	@go test -v $(addprefix ./,$(TESTS))

test: clean fmt-check unit-test smoke-test

fmt_quick_check:
	@# a very fast check before build, but depends on accessibility of all imports
	@# switch off the "stdmethods" analyzer is needed due to finding:
	@# at24cx/at24cx.go:57:18: method WriteByte(eepromAddress uint16, value uint8) error should have signature WriteByte(byte) error
	@# at24cx/at24cx.go:67:18: method ReadByte(eepromAddress uint16) (uint8, error) should have signature ReadByte() (byte, error)
	@# switch off the "shift" analyzer is needed due to finding:
	@#tmc5160/registers.go:1939:16: m.CUR_A (16 bits) too small for shift of 16
	@#tmc5160/registers.go:1996:16: m.X3 (8 bits) too small for shift of 27
	@#tmc5160/registers.go:1996:27: m.X2 (8 bits) too small for shift of 24
	@#tmc5160/registers.go:1996:38: m.X1 (8 bits) too small for shift of 21
	@#tmc5160/registers.go:1996:49: m.W3 (8 bits) too small for shift of 18
	@#tmc5160/registers.go:1996:60: m.W2 (8 bits) too small for shift of 16
	@#tmc5160/registers.go:1996:71: m.W1 (8 bits) too small for shift of 14
	@#tmc5160/registers.go:1996:82: m.W0 (8 bits) too small for shift of 12
	go vet -tags $(BUILD_TAGS_CHECK) -stdmethods=false -shift=false $(ALL_TO_CHECK)

fmt_check:
	@# a complete format check, but depends on accessibility of all imports
	golangci-lint -v run $(ALL_TO_CHECK)

fmt_fix:
	@# an automatic reformat and complete format check, but depends on accessibility of all imports
	@#TODO: activate when ready 
	@#gofumpt -l -w $(ALL_TO_CHECK)
	golangci-lint -v run $(ALL_TO_CHECK) --fix

print_collected_packages:
	@#this target is used to unify mechanism in CI with the local one, see ".github/workflows/golangci-lint.yml"
	@#we need additional exclude the root folder here, because will be checked recursive when used as working directory
	@echo $(filter-out $(CURDIR),$(ALL_TO_CHECK))
