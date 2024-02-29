package espradio

/*
#include "include.h"
*/
import "C"

// Error is an error from the radio stack.
type Error C.esp_err_t

func (e Error) Error() string {
	switch {
	case e >= C.ESP_ERR_MEMPROT_BASE:
		return "espradio: unknown memprot error"
	case e >= C.ESP_ERR_HW_CRYPTO_BASE:
		return "espradio: unknown hw crypto error"
	case e >= C.ESP_ERR_FLASH_BASE:
		return "espradio: unknown flash error"
	case e >= C.ESP_ERR_MESH_BASE:
		return "espradio: unknown mesh error"
	case e >= C.ESP_ERR_WIFI_BASE:
		return "espradio: unknown wifi error"
	default:
		switch e {
		case C.ESP_OK:
			return "espradio: no error" // invalid usage of the Error type
		case C.ESP_ERR_NO_MEM:
			return "espradio: no memory"
		case C.ESP_ERR_INVALID_ARG:
			return "espradio: invalid argument"
		default:
			return "espradio: unknown error"
		}
	}
}

// makeError returns an error (using the Error type) if the error code is
// non-zero, otherwise it returns nil.
func makeError(errCode C.esp_err_t) error {
	if errCode != 0 {
		return Error(errCode)
	}
	return nil
}
