package hub75

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/tester"
)

func TestDefaultHub75(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewSPIBus(c)
	pin1 := tester.NewPin(c)
	pin2 := tester.NewPin(c)
	pin3 := tester.NewPin(c)
	pin4 := tester.NewPin(c)
	pin5 := tester.NewPin(c)
	pin6 := tester.NewPin(c)
	dev := New(bus, pin1, pin2, pin3, pin4, pin5, pin6)
	c.Assert(dev, qt.Not(qt.IsNil))
}
