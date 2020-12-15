package hcsr04

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/tester"
)

func TestDefaultEasyStepper(t *testing.T) {
	c := qt.New(t)
	pin1 := tester.NewPin(c)
	pin2 := tester.NewPin(c)
	dev := New(pin1, pin2)
	c.Assert(dev, qt.Not(qt.IsNil))
}
