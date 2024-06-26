package backlight

import "machine"

type PWM interface {
	Configure(config machine.PWMConfig) error
	Channel(pin machine.Pin) (channel uint8, err error)
	Top() uint32
	Set(channel uint8, value uint32)
}

type PWMDriver struct {
	pwm PWM
	ch  uint8
}

func NewPWMDriver(pwm PWM, pin machine.Pin) (PWMDriver, error) {
	err := pwm.Configure(machine.PWMConfig{})
	if err != nil {
		return PWMDriver{}, err
	}
	ch, err := pwm.Channel(pin)
	if err != nil {
		return PWMDriver{}, err
	}
	return PWMDriver{pwm, ch}, nil
}

func (b PWMDriver) SetBrightness(brightness uint8) {
	b.pwm.Set(b.ch, b.pwm.Top()*uint32(brightness)/255)
}
