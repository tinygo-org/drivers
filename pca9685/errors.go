package pca9685

import "errors"

var (
	ErrInvalidMode1 = errors.New("pca9685: data read from MODE1 not valid")
	ErrBadPeriod    = errors.New("pca9685: period must be in range 1..25ms")
)
