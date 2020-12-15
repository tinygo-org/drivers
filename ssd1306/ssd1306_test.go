package ssd1306

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/tester"
)

func TestDefaultSSD1306I2C(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	dev := NewI2C(bus)
	c.Assert(dev, qt.Not(qt.IsNil))
}

func TestDefaultSSD1306SPI(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewSPIBus(c)
	pin1 := tester.NewPin(c)
	pin2 := tester.NewPin(c)
	pin3 := tester.NewPin(c)
	dev := NewSPI(bus, pin1, pin2, pin3)
	c.Assert(dev, qt.Not(qt.IsNil))
}
