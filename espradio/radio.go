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
	"unsafe"
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

func millisecondsToTicks(ms uint32) uint32 {
	return ms * (ticksPerSecond / 1000)
}

func ticksToMilliseconds(ticks uint32) uint32 {
	return ticks / (ticksPerSecond / 1000)
}

//export espradio_panic
func espradio_panic(msg *C.char) {
	panic("espradio: " + C.GoString(msg))
}

//export espradio_log_timestamp
func espradio_log_timestamp() uint32 {
	return uint32(time.Now().UnixMilli())
}

//export espradio_run_task
func espradio_run_task(task_func, param unsafe.Pointer)

//export espradio_task_create_pinned_to_core
func espradio_task_create_pinned_to_core(task_func unsafe.Pointer, name *C.char, stack_depth uint32, param unsafe.Pointer, prio uint32, task_handle *unsafe.Pointer, core_id uint32) int32 {
	ch := make(chan struct{}, 1)
	go func() {
		*task_handle = tinygo_task_current()
		close(ch)
		espradio_run_task(task_func, unsafe.Pointer(task_handle))
	}()
	<-ch
	return 1
}

//export espradio_task_delete
func espradio_task_delete(task_handle unsafe.Pointer) {
	println("espradio TODO: delete task", task_handle)
}

//export tinygo_task_current
func tinygo_task_current() unsafe.Pointer

//export espradio_task_get_current_task
func espradio_task_get_current_task() unsafe.Pointer {
	return tinygo_task_current()
}

//export espradio_task_delay
func espradio_task_delay(ticks uint32) {
	const ticksPerMillisecond = ticksPerSecond / 1000
	// Round milliseconds up.
	ms := (ticks + ticksPerMillisecond - 1) / ticksPerMillisecond
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

//export espradio_task_ms_to_tick
func espradio_task_ms_to_tick(ms uint32) int32 {
	return int32(millisecondsToTicks(ms))
}

//export espradio_wifi_int_disable
func espradio_wifi_int_disable(wifi_int_mux unsafe.Pointer) uint32 {
	// This is portENTER_CRITICAL (or portENTER_CRITICAL_ISR).
	return uint32(interrupt.Disable())
}

//export espradio_wifi_int_restore
func espradio_wifi_int_restore(wifi_int_mux unsafe.Pointer, tmp uint32) {
	// This is portEXIT_CRITICAL (or portEXIT_CRITICAL_ISR).
	interrupt.Restore(interrupt.State(tmp))
}
