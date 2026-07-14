package ds3231

import (
	"testing"
)

func TestPositiveMilliCelsius(t *testing.T) {
	t1000 := milliCelsius(0)
	if t1000 != 0 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(0b0000000001000000)
	if t1000 != 250 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(0b0000000010000000)
	if t1000 != 500 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(0b0000000011000000)
	if t1000 != 750 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(0b0000000100000000)
	if t1000 != 1000 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(0b0000001000000000)
	if t1000 != 2000 {
		t.Fatal(t1000)
	}

	// highest temperature is 127.750C
	t1000 = milliCelsius(0b0111111111000000)
	if t1000 != 127750 {
		t.Fatal(t1000)
	}
}

func TestNegativeMilliCelsius(t *testing.T) {
	t1000 := milliCelsius(0b1111111111000000)
	if t1000 != -250 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(0b1111111110000000)
	if t1000 != -500 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(0b1111111101000000)
	if t1000 != -750 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(0b1111111100000000)
	if t1000 != -1000 {
		t.Fatal(t1000)
	}

	t1000 = milliCelsius(0b1111111000000000)
	if t1000 != -2000 {
		t.Fatal(t1000)
	}

	// lowest temperature is -128.000C
	t1000 = milliCelsius(0b1000000000000000)
	if t1000 != -128000 {
		t.Fatal(t1000)
	}
}
