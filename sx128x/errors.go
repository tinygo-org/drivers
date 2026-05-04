package sx128x

import "errors"

var (
	ErrBusyPinTimeout        = errors.New("busy pin timeout")
	errDataTooLong           = errors.New("data over 256 bytes")
	errInvalidSleepConfig    = errors.New("invalid sleep config")
	errInvalidStandbyConfig  = errors.New("invalid standby config")
	errFrequencyTooLow       = errors.New("frequency below 2.4Ghz")
	errFrequencyTooHigh      = errors.New("frequency above 2.5Ghz")
	errPowerTooLow           = errors.New("power level below -18dBm")
	errPowerTooHigh          = errors.New("power level above 13dBm")
	errInvalidPeriodBase     = errors.New("invalid period base")
	errInvalidPacketType     = errors.New("invalid packet type")
	errInvalidRegulatorMode  = errors.New("invalid regulator mode")
	errPayloadLengthTooShort = errors.New("payload length too short")
	errPayloadLengthTooLong  = errors.New("payload length too long")
)
