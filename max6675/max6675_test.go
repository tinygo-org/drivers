package max6675_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/max6675"
	"tinygo.org/x/drivers/tester"
)

func Test_MAX6675_Read(t *testing.T) {
	c := qt.New(t)
	spi := tester.NewSPIBus(c)
	cs := tester.NewPin(c)

	var expectedCelsius float32 = 20.25

	temp := uint16(expectedCelsius/0.25) << 3

	cs.ExpectSet(false)
	spi.Expect(
		[]byte{0, 0},
		[]byte{byte(temp >> 8), byte(temp)},
	)
	cs.ExpectSet(true)

	dev := max6675.NewDevice(spi, cs)

	actual, err := dev.Read()
	c.Assert(err, qt.Equals, nil)
	c.Assert(actual, qt.Equals, expectedCelsius)
}

func Test_MAX6675_Read_ErrThermocoupleOpen(t *testing.T) {
	c := qt.New(t)
	spi := tester.NewSPIBus(c)

	spi.Expect(
		[]byte{0, 0},
		[]byte{0, 0x04},
	)

	dev := max6675.NewDevice(spi, tester.NewNoopPin())

	actual, err := dev.Read()
	c.Assert(err, qt.Equals, max6675.ErrThermocoupleOpen)
	c.Assert(actual, qt.Equals, float32(0))
}
