package espradio

/*
// TODO: vary these by chip
#cgo CFLAGS: -Iblobs/headers
#cgo CFLAGS: -Iblobs/headers/esp32c3
#cgo CFLAGS: -Iblobs/include
#cgo LDFLAGS: -Lblobs/libs/esp32c3 -lcore -lmesh -lnet80211 -lphy -lpp -lwpa_supplicant

#include "include.h"
extern wifi_init_config_t wifi_config;
*/
import "C"
import (
	"runtime/interrupt"
	"time"
)

type LogLevel uint8

// Various log levels to use inside the espradio. Higher log levels will produce
// more output over the serial console.
const (
	LogLevelNone    = C.WIFI_LOG_NONE
	LogLevelError   = C.WIFI_LOG_ERROR
	LogLevelWarning = C.WIFI_LOG_WARNING
	LogLevelInfo    = C.WIFI_LOG_INFO
	LogLevelDebug   = C.WIFI_LOG_DEBUG
	LogLevelVerbose = C.WIFI_LOG_VERBOSE
)

type Config struct {
	Logging LogLevel
}

// Enable and configure the radio.
func Enable(config Config) error {
	initHardware()

	// TODO: run timers in separate goroutine

	errCode := C.esp_wifi_internal_set_log_level(C.wifi_log_level_t(config.Logging))
	if errCode != 0 {
		return makeError(errCode)
	}

	// TODO: BLE needs the interrupts RWBT, RWBLE, BT_BB

	mask := interrupt.Disable()
	// TODO: setup 200Hz tick rate timer
	// TODO: init_clocks
	interrupt.Restore(mask)

	// Initialize the wireless stack.
	errCode = C.esp_wifi_init_internal(&C.wifi_config)
	if errCode != 0 {
		return makeError(errCode)
	}

	return nil
}

//export espradio_panic
func espradio_panic(msg *C.char) {
	panic("espradio: " + C.GoString(msg))
}

//export espradio_log_timestamp
func espradio_log_timestamp() uint32 {
	return uint32(time.Now().UnixMilli())
}
