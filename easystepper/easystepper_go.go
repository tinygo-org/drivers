package easystepper

import (
	"errors"
	"time"

	"tinygo.org/x/drivers/internal/pin"
)

func NewCrossPlatform(stepcount, rpm uint, mode StepMode, pins [4]pin.OutputFunc) (*Device, error) {
	if stepcount == 0 || rpm == 0 {
		return nil, errors.New("zero rpm and/or stepcount")
	}
	for i := range pins {
		if pins[i] == nil {
			return nil, errors.New("nil pin")
		}
	}
	d := &Device{
		pins:      pins,
		stepDelay: time.Second * 60 / time.Duration((stepcount * rpm)),
		stepMode:  mode,
		config:    func() {},
	}
	return d, nil
}
