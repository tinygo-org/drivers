package pcf8523

import (
	"encoding/hex"
	"testing"
	"time"
	"tinygo.org/x/drivers/tester"
)

func TestDecToBcd_RoundTrip(t *testing.T) {

	for i := 0; i < 60; i++ {
		a := bcd2bin(bin2bcd(i))
		if a != i {
			t.Logf("not equal: %d != %d", a, i)
			t.FailNow()
		}
	}
}

func TestDevice_Reset(t *testing.T) {
	bus := tester.NewI2CBus(t)
	fake := bus.NewDevice(DefaultAddress)

	dev := New(bus)

	err := dev.Reset()
	assertNoError(t, err)

	assertEquals(t, fake.Registers[rControl1], 0x58)
}

func TestDevice_SetPowerManagement(t *testing.T) {
	bus := tester.NewI2CBus(t)
	fake := bus.NewDevice(DefaultAddress)

	dev := New(bus)

	err := dev.SetPowerManagement(PowerManagement_SwitchOver_ModeStandard)
	assertNoError(t, err)

	assertEquals(t, fake.Registers[rControl3], 0b100<<5)
}

func TestDevice_SetTime(t *testing.T) {
	bus := tester.NewI2CBus(t)
	fake := bus.NewDevice(DefaultAddress)

	dev := New(bus)

	pointInTime, _ := time.Parse(time.RFC3339, "2023-09-12T22:35:50Z")
	err := dev.SetTime(pointInTime)
	assertNoError(t, err)

	actual := hex.EncodeToString(fake.Registers[rSeconds : rSeconds+7])
	expected := "50352212020923"
	assertEquals(t, actual, expected)
}

func TestDevice_ReadTime(t *testing.T) {
	bus := tester.NewI2CBus(t)
	fake := bus.NewDevice(DefaultAddress)

	expectedPointInTime := time.Date(2023, 9, 12, 17, 55, 42, 0, time.UTC)
	fake.Registers[rSeconds] = 0x42
	fake.Registers[rMinutes] = 0x55
	fake.Registers[rHours] = 0x17
	fake.Registers[rDays] = 0x12
	fake.Registers[rMonths] = 0x9
	fake.Registers[rYears] = 0x23

	dev := New(bus)

	//when
	actualPointInTime, err := dev.ReadTime()

	//then
	assertNoError(t, err)
	assertEquals(t, actualPointInTime, expectedPointInTime)
}

func assertNoError(t testing.TB, e error) {
	if e != nil {
		t.Fatalf("unexpected error: %v", e)
	}
}
func assertEquals[T comparable](t testing.TB, a, b T) {
	if a != b {
		t.Fatalf("%v != %v", a, b)
	}
}
