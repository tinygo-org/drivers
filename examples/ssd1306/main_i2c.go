//go:build !thumby

package main

import (
	"machine"

	"tinygo.org/x/drivers/ssd1306"
)

func makeSSD1306(width, height int16) (*ssd1306.Device, error) {
	err := machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
	})
	if err != nil {
		return nil, err
	}
	address := uint16(ssd1306.Address)
	if width == 128 && (height == 32 || height == 64) {
		address = ssd1306.Address_128_32
	}
	display := ssd1306.NewI2C(machine.I2C0)
	display.Configure(ssd1306.Config{
		Address: address,
		Width:   width,
		Height:  height,
	})
	return &display, nil
}
