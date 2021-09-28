package espnet

/*
#cgo CFLAGS: -DCONFIG_IDF_TARGET_ESP32C3
#cgo CFLAGS: -Iinclude
#cgo CFLAGS: -Iesp-idf/components/esp_common/include
#cgo CFLAGS: -Iesp-idf/components/esp_event/include
#cgo CFLAGS: -Iesp-idf/components/esp_netif/include
#cgo CFLAGS: -Iesp-idf/components/esp_wifi/include
#cgo CFLAGS: -Iesp-idf/components/esp_timer/include
#cgo CFLAGS: -Iesp-idf/components/riscv/include
#cgo CFLAGS: -Iesp-idf/components/heap/include
#cgo CFLAGS: -Iesp-idf/components/hal/include
#cgo CFLAGS: -Iesp-idf/components/soc/include
#cgo CFLAGS: -Iesp-idf/components/soc/esp32c3/include
#cgo CFLAGS: -Iesp-idf/components/hal/esp32c3/include
#cgo CFLAGS: -Iesp-idf/components/freertos/port/riscv/include
#cgo CFLAGS: -Iesp-idf/components/esp_hw_support/include
#cgo CFLAGS: -Iesp-idf/components/esp_hw_support/include/soc
#cgo CFLAGS: -Iesp-idf/components/esp_rom/include
#cgo CFLAGS: -Iesp-idf/components/esp_system/include
#cgo CFLAGS: -Iesp-idf/components/newlib/platform_include

#cgo LDFLAGS: -Lesp-idf/components/esp_wifi/lib/esp32c3 -lnet80211 -lpp -lphy -lmesh -lcore
#cgo LDFLAGS: -Tesp-idf/components/esp_rom/esp32c3/ld/esp32c3.rom.ld

#include "esp_private/wifi.h"
#include "esp_wifi_types.h"
#include "espnet.h"
*/
import "C"

import (
	"unsafe"
	_ "compat/freertos"
)

type ESPWiFi struct {
}

var WiFi = &ESPWiFi{}

type Config struct {
}

func (wifi ESPWiFi) Configure(config Config) error {
	C.esp_wifi_internal_set_log_level(5)
	cfg := &C.wifi_init_config_t{}
	C.wifi_init_default(unsafe.Pointer(cfg))
	println("magic:", cfg.magic)
	println("config address:", cfg)
	return makeError(C.esp_wifi_init_internal(cfg))
}

func (wifi ESPWiFi) AccessPointMAC() ([6]byte, error) {
	var mac [6]byte
	errCode := C.esp_wifi_get_mac(C.ESP_IF_WIFI_AP, &mac[0])
	return mac, makeError(errCode)
}
