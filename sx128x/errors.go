package sx128x

import "errors"

var (
	ErrBusyPinTimeout        = errors.New("busy pin timeout")
	ErrDataTooLong           = errors.New("data over 256 bytes")
	ErrInvalidSleepConfig    = errors.New("invalid sleep config")
	ErrInvalidStandbyConfig  = errors.New("invalid standby config")
	ErrFrequencyTooLow       = errors.New("frequency below 2.4Ghz")
	ErrFrequencyTooHigh      = errors.New("frequency above 2.5Ghz")
	ErrPowerTooLow           = errors.New("power level below -18dBm")
	ErrPowerTooHigh          = errors.New("power level above 13dBm")
	ErrInvalidPeriodBase     = errors.New("invalid period base")
	ErrInvalidPacketType     = errors.New("invalid packet type")
	ErrInvalidRegulatorMode  = errors.New("invalid regulator mode")
	ErrPayloadLengthTooShort = errors.New("payload length too short")
	ErrPayloadLengthTooLong  = errors.New("payload length too long")
)
