package unor4matrix

import (
	"errors"
	"image/color"
	"reflect"
	"testing"
	"time"
)

type event struct {
	line   int
	action string
}

func traceDevice() (Device, *[]event) {
	var pins [numPins]CharlieplexPin
	events := new([]event)
	for i := range pins {
		pins[i] = CharlieplexPin{
			Set: func(level bool) {
				action := "low"
				if level {
					action = "high"
				}
				*events = append(*events, event{i, action})
			},
			Float: func() { *events = append(*events, event{i, "float"}) },
		}
	}
	return New(pins), events
}

func TestDisplayBeforeConfigure(t *testing.T) {
	d, events := traceDevice()
	d.LoadFrame([3]uint32{0x80000000, 0, 1})
	if err := d.Display(); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("unconfigured Display = %v", err)
	}
	if len(*events) != 0 {
		t.Fatal("unconfigured display touched GPIO")
	}
}

func TestDisplayScansCurrentPixelsAndBlanks(t *testing.T) {
	d, events := traceDevice()
	d.LoadFrame([3]uint32{0x80000000, 0, 1})
	if err := d.Configure(); err != nil {
		t.Fatal(err)
	}
	var floats []event
	for i := 0; i < 11; i++ {
		floats = append(floats, event{i, "float"})
	}
	if !reflect.DeepEqual(*events, floats) {
		t.Fatalf("configure = %v", *events)
	}
	*events = nil
	if err := d.Display(); err != nil {
		t.Fatal(err)
	}
	want := []event{{3, "low"}, {7, "high"}, {7, "float"}, {3, "float"}, {4, "low"}, {9, "high"}, {9, "float"}, {4, "float"}}
	if !reflect.DeepEqual(*events, want) {
		t.Fatalf("display = %v, want %v", *events, want)
	}
	d.ClearDisplay()
	d.SetPixel(1, 0, color.RGBA{R: 255})
	if err := d.Configure(); err != nil {
		t.Fatal(err)
	}
	*events = nil
	if err := d.Display(); err != nil {
		t.Fatal(err)
	}
	want = []event{{7, "low"}, {3, "high"}, {3, "float"}, {7, "float"}}
	if !reflect.DeepEqual(*events, want) {
		t.Fatalf("next frame = %v, want %v", *events, want)
	}
}

func TestDarkFrameKeepsAllSlotDelays(t *testing.T) {
	d, events := traceDevice()
	if err := d.Configure(); err != nil {
		t.Fatal(err)
	}
	*events = nil
	start := time.Now()
	if err := d.Display(); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed < 96*100*time.Microsecond {
		t.Fatalf("dark frame returned too early: %v", elapsed)
	}
	if len(*events) != 0 {
		t.Fatalf("dark frame drove GPIO: %v", *events)
	}
}

func TestIncompletePinsDoNotTouchHardware(t *testing.T) {
	for _, missingSet := range []bool{true, false} {
		var pins [numPins]CharlieplexPin
		touched := false
		for i := range pins {
			pins[i] = CharlieplexPin{Set: func(bool) { touched = true }, Float: func() { touched = true }}
		}
		if missingSet {
			pins[10].Set = nil
		} else {
			pins[10].Float = nil
		}
		d := New(pins)
		if err := d.Configure(); !errors.Is(err, ErrMissingPinFunc) {
			t.Fatalf("configure = %v", err)
		}
		if touched {
			t.Fatal("invalid device touched GPIO")
		}
	}
}
