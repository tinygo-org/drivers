// Package unoqmatrix provides a driver for the UnoQMatrix LED matrix display.
//
// The UnoQMatrix is an 8x13 LED matrix display that can be controlled using a single pin.
// It uses a multiplexing technique to control the LEDs, which allows for a large number of LEDs to be controlled with fewer pins.
//
// This driver provides basic functionality to set individual pixels, clear the display, and refresh the display.
//
// Note: The UnoQMatrix does not support brightness control or color depth. Each pixel can only be turned on or off.
// Could it suppport brightness control by using PWM on the pin? To be investigated.
package unoqmatrix

import (
	"image/color"
	"time"

	pin "tinygo.org/x/drivers/internal/pin"
)

type Config struct {
	// Rotation of the LED matrix.
	Rotation uint8
}

// Valid values:
//
//	0: regular orientation, (0 degree rotation)
//	1: 90 degree rotation clockwise
//	2: 180 degree rotation clockwise
//	3: 270 degree rotation clockwise
const (
	RotationNormal = 0
	Rotation90     = 1
	Rotation180    = 2
	Rotation270    = 3
)

const (
	ledRows = 8
	ledCols = 13

	pixelRefreshDelay = 10 * time.Microsecond
)

// CharlieplexPin represents a pin used for charlieplexing.
// It must be able to drive high/low (output mode) and float (high-impedance/input mode).
//
// Example construction from a machine.Pin using the pin HAL pattern:
//
//	var isOutput bool
//	cp := unoqmatrix.CharlieplexPin{
//		Set: pin.OutputFunc(func(level bool) {
//			if !isOutput {
//				p.Configure(machine.PinConfig{Mode: machine.PinOutput})
//				isOutput = true
//			}
//			p.Set(level)
//		}),
//		Float: func() {
//			if isOutput {
//				p.Configure(machine.PinConfig{Mode: machine.PinInput})
//				isOutput = false
//			}
//		},
//	}
type CharlieplexPin struct {
	Set   pin.OutputFunc // Drive pin high (true) or low (false); auto-configures to output mode.
	Float func()         // Put pin into high-impedance (input) mode.
}

const numPins = 11

// Device represents the UnoQMatrix LED matrix display.
type Device struct {
	pins     [numPins]CharlieplexPin
	buffer   [ledRows][ledCols]color.RGBA
	rotation uint8
}

// New returns a new unoqmatrix driver.
// The provided pins are the 11 charlieplex pins used to control the LED matrix.
func New(pins [numPins]CharlieplexPin) Device {
	return Device{pins: pins}
}

// Configure sets up the device.
func (d *Device) Configure(cfg Config) {
	d.SetRotation(cfg.Rotation)
}

// SetRotation changes the rotation of the LED matrix.
//
// Valid values for rotation:
//
//	0: regular orientation, (0 degree rotation)
//	1: 90 degree rotation clockwise
//	2: 180 degree rotation clockwise
//	3: 270 degree rotation clockwise
func (d *Device) SetRotation(rotation uint8) {
	d.rotation = rotation % 4
}

// SetPixel sets the color of a specific pixel.
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	d.buffer[y][x] = c
}

// GetPixel returns the color of a specific pixel.
func (d *Device) GetPixel(x int16, y int16) color.RGBA {
	return d.buffer[y][x]
}

// Display sends the buffer (if any) to the screen.
// Only lights active (non-black) pixels, and resets only the 2 previously
// driven pins between LEDs instead of all 11, making each refresh cycle
// proportional to the number of lit LEDs.
func (d *Device) Display() error {
	d.clearDisplay()

	var lastIdx0, lastIdx1 uint8
	hasLast := false

	for row := 0; row < ledRows; row++ {
		for col := 0; col < ledCols; col++ {
			c := d.buffer[row][col]
			if c.R == 0 && c.G == 0 && c.B == 0 {
				continue
			}

			idx := row*ledCols + col
			if idx < 0 || idx >= len(pinMapping) {
				continue
			}

			// Float only the two pins that were driving the previous LED.
			if hasLast {
				d.pins[lastIdx0].Float()
				d.pins[lastIdx1].Float()
			}
			hasLast = true

			idx0 := pinMapping[idx][0]
			idx1 := pinMapping[idx][1]
			d.pins[idx0].Set.High()
			d.pins[idx1].Set.Low()
			lastIdx0 = idx0
			lastIdx1 = idx1

			time.Sleep(pixelRefreshDelay)
		}
	}

	// Float the last driven LED.
	if hasLast {
		d.pins[lastIdx0].Float()
		d.pins[lastIdx1].Float()
	}

	return nil
}

// ClearDisplay turns off all the LEDs on the display.
func (d *Device) ClearDisplay() {
	for row := 0; row < ledRows; row++ {
		for col := 0; col < ledCols; col++ {
			d.buffer[row][col] = color.RGBA{0, 0, 0, 255}
		}
	}
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	return ledCols, ledRows
}

// pinMapping defines the mapping of LED indices to pin pairs. Each entry corresponds
// to an LED index (0-104) and contains the two pin numbers that need to be set to turn on that LED.
// based on https://github.com/arduino/ArduinoCore-zephyr/blob/main/loader/matrix.inc#L13
var pinMapping = [][2]uint8{
	{0, 1}, // 0
	{1, 0},
	{0, 2},
	{2, 0},
	{1, 2},
	{2, 1},
	{0, 3},
	{3, 0},
	{1, 3},
	{3, 1},
	{2, 3}, // 10
	{3, 2},
	{0, 4},
	{4, 0},
	{1, 4},
	{4, 1},
	{2, 4},
	{4, 2},
	{3, 4},
	{4, 3},
	{0, 5}, // 20
	{5, 0},
	{1, 5},
	{5, 1},
	{2, 5},
	{5, 2},
	{3, 5},
	{5, 3},
	{4, 5},
	{5, 4},
	{0, 6}, // 30
	{6, 0},
	{1, 6},
	{6, 1},
	{2, 6},
	{6, 2},
	{3, 6},
	{6, 3},
	{4, 6},
	{6, 4},
	{5, 6}, // 40
	{6, 5},
	{0, 7},
	{7, 0},
	{1, 7},
	{7, 1},
	{2, 7},
	{7, 2},
	{3, 7},
	{7, 3},
	{4, 7}, // 50
	{7, 4},
	{5, 7},
	{7, 5},
	{6, 7},
	{7, 6},
	{0, 8},
	{8, 0},
	{1, 8},
	{8, 1},
	{2, 8}, // 60
	{8, 2},
	{3, 8},
	{8, 3},
	{4, 8},
	{8, 4},
	{5, 8},
	{8, 5},
	{6, 8},
	{8, 6},
	{7, 8}, // 70
	{8, 7},
	{0, 9},
	{9, 0},
	{1, 9},
	{9, 1},
	{2, 9},
	{9, 2},
	{3, 9},
	{9, 3},
	{4, 9}, // 80
	{9, 4},
	{5, 9},
	{9, 5},
	{6, 9},
	{9, 6},
	{7, 9},
	{9, 7},
	{8, 9},
	{9, 8},
	{0, 10}, // 90
	{10, 0},
	{1, 10},
	{10, 1},
	{2, 10},
	{10, 2},
	{3, 10},
	{10, 3},
	{4, 10},
	{10, 4},
	{5, 10}, // 100
	{10, 5},
	{6, 10},
	{10, 6},
}

// clearDisplay turns off all the LEDs on the display by floating all pins.
func (d *Device) clearDisplay() {
	for i := range d.pins {
		d.pins[i].Float()
	}
}
