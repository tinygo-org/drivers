package shiftregister

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/tester"
)

func TestDefaultShiftregister(t *testing.T) {
	c := qt.New(t)
	pin1 := tester.NewPin(c)
	pin2 := tester.NewPin(c)
	pin3 := tester.NewPin(c)
	dev := New(EIGHT_BITS, pin1, pin2, pin3)
	c.Assert(dev, qt.Not(qt.IsNil))
}
