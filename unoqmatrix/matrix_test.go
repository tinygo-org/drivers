package unoqmatrix

import (
	"image/color"
	"testing"

	pin "tinygo.org/x/drivers/internal/pin"
)

// pinState tracks the state of a mock charlieplex pin.
type pinState struct {
	level    bool // true=high, false=low
	isOutput bool // true=output mode, false=floating (high-Z)
}

// mockPins creates 11 mock CharlieplexPins and returns them along with their observable state.
func mockPins() ([numPins]CharlieplexPin, *[numPins]pinState) {
	var pins [numPins]CharlieplexPin
	var states [numPins]pinState
	for i := range pins {
		idx := i // capture
		pins[i] = CharlieplexPin{
			Set: pin.OutputFunc(func(level bool) {
				states[idx].isOutput = true
				states[idx].level = level
			}),
			Float: func() {
				states[idx].isOutput = false
				states[idx].level = false
			},
		}
	}
	return pins, &states
}

func newTestDevice() (Device, *[numPins]pinState) {
	pins, states := mockPins()
	d := New(pins)
	return d, states
}

func TestNew(t *testing.T) {
	d, _ := newTestDevice()
	w, h := d.Size()
	if w != ledCols || h != ledRows {
		t.Errorf("Size() = (%d, %d), want (%d, %d)", w, h, ledCols, ledRows)
	}
}

func TestSize(t *testing.T) {
	d, _ := newTestDevice()
	w, h := d.Size()
	if w != 13 {
		t.Errorf("width = %d, want 13", w)
	}
	if h != 8 {
		t.Errorf("height = %d, want 8", h)
	}
}

func TestSetGetPixel(t *testing.T) {
	d, _ := newTestDevice()
	c := color.RGBA{R: 255, G: 128, B: 64, A: 255}

	d.SetPixel(3, 2, c)
	got := d.GetPixel(3, 2)
	if got != c {
		t.Errorf("GetPixel(3,2) = %v, want %v", got, c)
	}

	// Unset pixel should be zero-value.
	got = d.GetPixel(0, 0)
	if got != (color.RGBA{}) {
		t.Errorf("GetPixel(0,0) = %v, want zero", got)
	}
}

func TestClearDisplay(t *testing.T) {
	d, _ := newTestDevice()
	on := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	off := color.RGBA{A: 255}

	d.SetPixel(0, 0, on)
	d.SetPixel(5, 3, on)
	d.ClearDisplay()

	for y := int16(0); y < ledRows; y++ {
		for x := int16(0); x < ledCols; x++ {
			got := d.GetPixel(x, y)
			if got != off {
				t.Errorf("after ClearDisplay, GetPixel(%d,%d) = %v, want %v", x, y, got, off)
			}
		}
	}
}

func TestSetRotation(t *testing.T) {
	d, _ := newTestDevice()

	tests := []struct {
		input uint8
		want  uint8
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 0}, // wraps
		{7, 3}, // wraps
	}
	for _, tt := range tests {
		d.SetRotation(tt.input)
		if d.rotation != tt.want {
			t.Errorf("SetRotation(%d): rotation = %d, want %d", tt.input, d.rotation, tt.want)
		}
	}
}

func TestConfigure(t *testing.T) {
	d, _ := newTestDevice()
	d.Configure(Config{Rotation: 2})
	if d.rotation != 2 {
		t.Errorf("Configure(Rotation:2): rotation = %d, want 2", d.rotation)
	}
}

func TestDisplayEmptyBuffer(t *testing.T) {
	d, states := newTestDevice()

	err := d.Display()
	if err != nil {
		t.Fatalf("Display() error: %v", err)
	}

	// All pins should be floating after displaying an empty buffer.
	for i, s := range states {
		if s.isOutput {
			t.Errorf("pin %d still in output mode after empty Display()", i)
		}
	}
}

func TestDisplaySinglePixel(t *testing.T) {
	d, states := newTestDevice()
	on := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// LED index 0 -> pinMapping[0] = {0, 1}: pin 0 high, pin 1 low.
	d.SetPixel(0, 0, on)
	err := d.Display()
	if err != nil {
		t.Fatalf("Display() error: %v", err)
	}

	// After Display completes, all pins should be floating (last LED turned off).
	for i, s := range states {
		if s.isOutput {
			t.Errorf("pin %d still in output mode after Display()", i)
		}
	}
}

func TestDisplayMultiplePixels(t *testing.T) {
	d, states := newTestDevice()
	on := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	d.SetPixel(0, 0, on) // idx 0 -> pins {0,1}
	d.SetPixel(1, 0, on) // idx 1 -> pins {1,0}
	d.SetPixel(2, 0, on) // idx 2 -> pins {0,2}

	err := d.Display()
	if err != nil {
		t.Fatalf("Display() error: %v", err)
	}

	// All pins floating after display completes.
	for i, s := range states {
		if s.isOutput {
			t.Errorf("pin %d still in output mode after Display()", i)
		}
	}
}

// pinEvent records a single pin action during Display().
type pinEvent struct {
	pinIdx int
	action string // "high", "low", or "float"
}

// traceDevice creates a device that records every pin event for verification.
func traceDevice() (Device, *[]pinEvent) {
	var pins [numPins]CharlieplexPin
	events := &[]pinEvent{}
	for i := range pins {
		idx := i
		pins[i] = CharlieplexPin{
			Set: pin.OutputFunc(func(level bool) {
				action := "low"
				if level {
					action = "high"
				}
				*events = append(*events, pinEvent{pinIdx: idx, action: action})
			}),
			Float: func() {
				*events = append(*events, pinEvent{pinIdx: idx, action: "float"})
			},
		}
	}
	d := New(pins)
	return d, events
}

func TestDisplayDrivesCorrectPins(t *testing.T) {
	d, events := traceDevice()
	on := color.RGBA{R: 255, A: 255}

	// Set pixel at (0,0) -> LED index 0 -> pinMapping[0] = {0, 1}.
	d.SetPixel(0, 0, on)
	d.Display()

	// Expected sequence:
	// 1. clearDisplay: float pins 0..10
	// 2. Drive LED 0: pin 0 high, pin 1 low
	// 3. Cleanup: float pin 0, float pin 1

	// Find the high/low events (skip initial floats from clearDisplay).
	var driveEvents []pinEvent
	for _, e := range *events {
		if e.action == "high" || e.action == "low" {
			driveEvents = append(driveEvents, e)
		}
	}

	if len(driveEvents) != 2 {
		t.Fatalf("expected 2 drive events, got %d: %v", len(driveEvents), driveEvents)
	}
	if driveEvents[0].pinIdx != 0 || driveEvents[0].action != "high" {
		t.Errorf("first drive event = %v, want pin 0 high", driveEvents[0])
	}
	if driveEvents[1].pinIdx != 1 || driveEvents[1].action != "low" {
		t.Errorf("second drive event = %v, want pin 1 low", driveEvents[1])
	}
}

func TestDisplaySkipsBlackPixels(t *testing.T) {
	d, events := traceDevice()
	on := color.RGBA{R: 255, A: 255}

	// Only set one pixel in the middle of the matrix.
	d.SetPixel(4, 1, on) // idx = 1*13+4 = 17 -> pinMapping[17] = {4,2}
	d.Display()

	var driveEvents []pinEvent
	for _, e := range *events {
		if e.action == "high" || e.action == "low" {
			driveEvents = append(driveEvents, e)
		}
	}

	// Should only drive one LED's worth of pin events.
	if len(driveEvents) != 2 {
		t.Fatalf("expected 2 drive events for 1 lit pixel, got %d", len(driveEvents))
	}
	if driveEvents[0].pinIdx != 4 || driveEvents[0].action != "high" {
		t.Errorf("expected pin 4 high, got %v", driveEvents[0])
	}
	if driveEvents[1].pinIdx != 2 || driveEvents[1].action != "low" {
		t.Errorf("expected pin 2 low, got %v", driveEvents[1])
	}
}

func TestDisplayFloatsBetweenLEDs(t *testing.T) {
	d, events := traceDevice()
	on := color.RGBA{R: 255, A: 255}

	d.SetPixel(0, 0, on) // idx 0 -> {0,1}
	d.SetPixel(1, 0, on) // idx 1 -> {1,0}
	d.Display()

	// After the initial clearDisplay floats, the sequence for two LEDs should be:
	// drive LED0 (pin0 high, pin1 low)
	// float pin0, float pin1  (between LEDs)
	// drive LED1 (pin1 high, pin0 low)
	// float pin1, float pin0  (cleanup)

	// Skip the initial 11 float events from clearDisplay.
	postClear := (*events)[numPins:]

	// Verify pin 0 and 1 are floated between the two LEDs.
	foundFloatBetween := false
	driveCount := 0
	for _, e := range postClear {
		if e.action == "high" || e.action == "low" {
			driveCount++
		}
		// After the first pair of drive events, we should see floats before the next pair.
		if driveCount == 2 && e.action == "float" {
			foundFloatBetween = true
			break
		}
	}
	if !foundFloatBetween {
		t.Error("expected float events between LED drives, found none")
	}
}

func TestPinMappingLength(t *testing.T) {
	expected := 104 // 8x13 matrix = 104 LEDs
	if len(pinMapping) != expected {
		t.Errorf("pinMapping has %d entries, want %d", len(pinMapping), expected)
	}
}

func TestPinMappingIndicesInRange(t *testing.T) {
	for i, pair := range pinMapping {
		if pair[0] >= numPins {
			t.Errorf("pinMapping[%d][0] = %d, exceeds numPins (%d)", i, pair[0], numPins)
		}
		if pair[1] >= numPins {
			t.Errorf("pinMapping[%d][1] = %d, exceeds numPins (%d)", i, pair[1], numPins)
		}
		if pair[0] == pair[1] {
			t.Errorf("pinMapping[%d] has same pin for both: %d", i, pair[0])
		}
	}
}
