package aht20

import "errors"

const (
	Address = 0x38

	CMD_INITIALIZE = 0xBE
	CMD_STATUS     = 0x71
	CMD_TRIGGER    = 0xAC
	CMD_SOFTRESET  = 0xBA

	STATUS_BUSY       = 0x80
	STATUS_CALIBRATED = 0x08
)

var (
	ErrBusy    = errors.New("device busy")
	ErrTimeout = errors.New("timeout")
)
