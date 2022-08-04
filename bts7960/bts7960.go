package bts7960

import (
	"machine"
	"time"
)

type Device struct {
	lEn            machine.Pin
	rEn            machine.Pin
	lPwm           machine.Pin
	rPwm           machine.Pin
	rPwmCh, lPwmCh uint8
	pwm            machine.PWM
}

// New returns a new motor driver.
func New(lEn, rEn, lPwm, rPwm machine.Pin, pwm machine.PWM) *Device {
	return &Device{lEn: lEn, rEn: rEn, lPwm: lPwm, rPwm: rPwm, pwm: pwm}
}

// Configure configures the Device.
func (d *Device) Configure() error {
	d.rPwm.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.lPwm.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.lEn.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.rEn.Configure(machine.PinConfig{Mode: machine.PinOutput})

	var err error
	d.lPwmCh, err = d.pwm.Channel(d.lPwm)
	if err != nil {
		println("failed to configure lpwn: " + err.Error())
		return err
	}

	d.rPwmCh, err = d.pwm.Channel(d.rPwm)
	if err != nil {
		println("failed to configure rpwn: " + err.Error())
		return err
	}

	d.Stop()

	return nil
}

// Enable enables the motor driver
func (d *Device) Enable() {
	d.lEn.High()
	d.rEn.High()
}

// Disable disabled the motor driver
func (d *Device) Disable() {
	d.lEn.Low()
	d.rEn.Low()
}

// Stop stops the motor
func (d *Device) Stop() {
	d.pwm.Set(d.lPwmCh, 0)
	d.pwm.Set(d.rPwmCh, 0)
}

// Left turns motor left.
func (d *Device) Left(speed uint32) {
	d.pwm.Set(d.rPwmCh, 0)
	time.Sleep(time.Microsecond * 100)
	d.pwm.Set(d.lPwmCh, d.pwm.Top()*speed/100)
}

// Right turns motor right.
func (d *Device) Right(speed uint32) {
	d.pwm.Set(d.lPwmCh, 0)
	time.Sleep(time.Microsecond * 100)
	d.pwm.Set(d.rPwmCh, d.pwm.Top()*speed/100)
}
