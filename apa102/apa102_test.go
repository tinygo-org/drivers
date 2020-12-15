package apa102

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/tester"
)

func TestDefaultAPA102(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewSPIBus(c)
	dev := New(bus)
	c.Assert(dev, qt.Not(qt.IsNil))
}
