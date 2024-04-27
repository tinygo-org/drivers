
clean:
	@rm -rf build

FMT_PATHS = ./

fmt-check:
	@unformatted=$$(gofmt -l $(FMT_PATHS)); [ -z "$$unformatted" ] && exit 0; echo "Unformatted:"; for fn in $$unformatted; do echo "  $$fn"; done; exit 1

XTENSA ?= 1
smoke-test:
	@mkdir -p build
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
