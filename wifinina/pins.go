package wifinina

import "errors"

// Mimics machine package's pin control
//
// NB! These are NINA chip pins, not main unit pins.
//
// Digital pin values and modes taken from
// https://github.com/arduino/nina-fw/blob/master/arduino/cores/esp32/wiring_digital.h

type Pin uint8

const (
	PinLow uint8 = iota
	PinHigh
)

type PinMode uint8

const (
	PinInput PinMode = iota
	PinOutput
	PinInputPullup
)

type PinConfig struct {
	Mode PinMode
}

var (
	ErrPinNoDevice = errors.New("wifinina pin: device not set")
)

var pinDevice *Device

func pinUseDevice(d *Device) {
	pinDevice = d
}

func (p Pin) Configure(config PinConfig) error {
	if pinDevice == nil {
		return ErrPinNoDevice
	}
	return pinDevice.PinMode(uint8(p), uint8(config.Mode))
}

func (p Pin) Set(high bool) error {
	if pinDevice == nil {
		return ErrPinNoDevice
	}
	value := PinLow
	if high {
		value = PinHigh
	}
	return pinDevice.DigitalWrite(uint8(p), value)
}

func (p Pin) High() error {
	return p.Set(true)
}

func (p Pin) Low() error {
	return p.Set(false)
}
