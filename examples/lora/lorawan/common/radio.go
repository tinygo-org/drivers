package common

import (
	"errors"
)

var (
	errRadioNotFound = errors.New("radio not found")
	errRxTimeout     = errors.New("radio RX timeout")
)
