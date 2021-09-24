package espnet

// #include <esp_err.h>
// #include <esp_wifi.h>
import "C"

// Wrapper for C.esp_err_t. Don't convert a C.esp_err_t to an Error type,
// instead use makeError to handle ESP_OK.
type Error C.esp_err_t

// makeError converts a C.esp_err_t into an error or nil depending on whether
// errCode indicates an error or not.
func makeError(errCode C.esp_err_t) error {
	if errCode == C.ESP_OK {
		return nil
	}
	return Error(errCode)
}

func (e Error) Error() string {
	switch {
	case e < C.ESP_ERR_WIFI_BASE:
		// esp-idf/components/esp_common/include/esp_err.h
		switch e {
		case C.ESP_OK:
			return "OK" // not an error
		case C.ESP_FAIL:
			return "ESP FAIL"
		case C.ESP_ERR_NO_MEM:
			return "Out of memory"
		case C.ESP_ERR_INVALID_ARG:
			return "Invalid argument"
		case C.ESP_ERR_INVALID_STATE:
			return "Invalid state"
		case C.ESP_ERR_INVALID_SIZE:
			return "Invalid size"
		case C.ESP_ERR_NOT_FOUND:
			return "Requested resource not found"
		case C.ESP_ERR_NOT_SUPPORTED:
			return "Operation or feature not supported"
		case C.ESP_ERR_TIMEOUT:
			return "Operation timed out"
		case C.ESP_ERR_INVALID_RESPONSE:
			return "Received response was invalid"
		case C.ESP_ERR_INVALID_CRC:
			return "CRC or checksum was invalid"
		case C.ESP_ERR_INVALID_VERSION:
			return "Version was invalid"
		case C.ESP_ERR_INVALID_MAC:
			return "MAC address was invalid"
		default:
			return "Unknown error"
		}
	case e >= C.ESP_ERR_WIFI_BASE && e < C.ESP_ERR_MESH_BASE:
		// esp-idf/components/esp_wifi/include/esp_wifi.h
		switch e {
		case C.ESP_ERR_WIFI_NOT_INIT:
			return "WiFi driver was not installed by esp_wifi_init"
		case C.ESP_ERR_WIFI_NOT_STARTED:
			return "WiFi driver was not started by esp_wifi_start"
		case C.ESP_ERR_WIFI_NOT_STOPPED:
			return "WiFi driver was not stopped by esp_wifi_stop"
		case C.ESP_ERR_WIFI_IF:
			return "WiFi interface error"
		case C.ESP_ERR_WIFI_MODE:
			return "WiFi mode error"
		case C.ESP_ERR_WIFI_STATE:
			return "WiFi internal state error"
		case C.ESP_ERR_WIFI_CONN:
			return "WiFi internal control block of station or soft-AP error"
		case C.ESP_ERR_WIFI_NVS:
			return "WiFi internal NVS module error"
		case C.ESP_ERR_WIFI_MAC:
			return "MAC address is invalid"
		case C.ESP_ERR_WIFI_SSID:
			return " SSID is invalid"
		case C.ESP_ERR_WIFI_PASSWORD:
			return "Password is invalid"
		case C.ESP_ERR_WIFI_TIMEOUT:
			return "Timeout error"
		case C.ESP_ERR_WIFI_WAKE_FAIL:
			return "WiFi is in sleep state(RF closed) and wakeup fail"
		case C.ESP_ERR_WIFI_WOULD_BLOCK:
			return "The caller would block"
		case C.ESP_ERR_WIFI_NOT_CONNECT:
			return "Station still in disconnect status"
		case C.ESP_ERR_WIFI_POST:
			return "Failed to post the event to WiFi task"
		case C.ESP_ERR_WIFI_INIT_STATE:
			return "Invalid WiFi state when init/deinit is called"
		case C.ESP_ERR_WIFI_STOP_STATE:
			return "Returned when WiFi is stopping"
		case C.ESP_ERR_WIFI_NOT_ASSOC:
			return "The WiFi connection is not associated"
		case C.ESP_ERR_WIFI_TX_DISALLOW:
			return "The WiFi TX is disallowed"
		default:
			return "Other WiFi error"
		}
	default:
		return "Other error"
	}
}
