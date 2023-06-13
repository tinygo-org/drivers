package ds3231

import (
	"testing"
)

func TestPositiveMilliCelsius(t *testing.T) {
	t1000 := milliCelsius(0, 0)
	if t1000 != 0 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(0, 0b01000000)
	if t1000 != 250 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(0, 0b10000000)
	if t1000 != 500 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(0, 0b11000000)
	if t1000 != 750 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(1, 0b00000000)
	if t1000 != 1000 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(2, 0b00000000)
	if t1000 != 2000 {
		t.Fatal(t1000)
	}

	// highest temperature is 127.750C
	t1000 = milliCelsius(0x7f, 0b11000000)
	if t1000 != 127750 {
		t.Fatal(t1000)
	}
}

func TestNegativeMilliCelsius(t *testing.T) {
	t1000 := milliCelsius(0xff, 0b11000000)
	if t1000 != -250 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(0xff, 0b10000000)
	if t1000 != -500 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(0xff, 0b01000000)
	if t1000 != -750 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(0xff, 0b00000000)
	if t1000 != -1000 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(0xfe, 0b00000000)
	if t1000 != -2000 {
		t.Fatal(t1000)
	}

	// lowest temperature is -128.000C
	t1000 = milliCelsius(0x80, 0b00000000)
	if t1000 != -128000 {
		t.Fatal(t1000)
	}
}
