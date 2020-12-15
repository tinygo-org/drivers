package buzzer

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/tester"
)

func TestDefaultBuzzer(t *testing.T) {
	c := qt.New(t)
	pin := tester.NewPin(c)
	dev := New(pin)
	c.Assert(dev, qt.Not(qt.IsNil))
}
