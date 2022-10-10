package microbitmatrix // import "tinygo.org/x/drivers/microbitmatrix"

import (
	"image/color"
	"time"
)

type Config struct {
	// Rotation of the LED matrix.
	//
	// Valid values:
	//
	//     0: regular orientation, (0 degree rotation)
	//     1: 90 degree rotation clock wise
	//     2: 180 degree rotation clock wise
	//     3: 270 degree rotation clock wise
	Rotation uint8
}

const (
	RotationNormal = 0
	Rotation90     = 1
	Rotation180    = 2
	Rotation270    = 3
)

// New returns a new microbitmatrix driver.
func New() Device {
	return Device{}
}

// Configure sets up the device.
func (d *Device) Configure(cfg Config) {
	d.SetRotation(cfg.Rotation)

	d.assignPins()

	d.ClearDisplay()
	d.DisableAll()
}

// SetRotation changes the rotation of the LED matrix.
//
// Valid values for rotation:
//
//	0: regular orientation, (0 degree rotation)
//	1: 90 degree rotation clock wise
//	2: 180 degree rotation clock wise
//	3: 270 degree rotation clock wise
func (d *Device) SetRotation(rotation uint8) {
	d.rotation = rotation % 4
}

// Source:
// https://github.com/bbcmicrobit/micropython/blob/1252f887ddc790676bf9314a136bd17650b9c36c/source/microbit/microbitdisplay.cpp#L282
var renderTimings = []time.Duration{
	0,  // Bright, Ticks Duration, Relative power
	2,  //   1,   2,     32µs,     inf
	2,  //   2,   4,     64µs,     200%
	4,  //   3,   8,     128µs,    200%
	7,  //   4,   15,    240µs,    187%
	13, //   5,   28,    448µs,    187%
	25, //   6,   53,    848µs,    189%
	49, //   7,   102,   1632µs,   192%
	97, //   8,   199,   3184µs,   195%
}

// Source:
// https://github.com/bbcmicrobit/micropython/blob/1252f887ddc790676bf9314a136bd17650b9c36c/source/microbit/microbitdisplay.cpp#L368
const tickDuration = 16 * time.Microsecond

const (
	rowIdx = 0
	colIdx = 1
)

// SetPixel modifies the internal buffer in a single pixel.
//
// The alpha channel of the RGBA is used to control the brightness of the LED
// in 9 different levels.
//
//	alpha channel, brightness level
//	  0 -  27, 9 (no transparency = highest brightness)
//	 28 -  55, 8
//	 56 -  83, 7
//	 84 - 111, 6
//	112 - 139, 5
//	140 - 167, 4
//	168 - 195, 3
//	196 - 223, 2
//	224 - 251, 1 (very high transparency = lowest brightness)
//	252 - 255, 0 (full transparency = off)
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	if x < 0 || x >= 5 || y < 0 || y >= 5 {
		return
	}
	col := x
	row := y
	if c.R != 0 || c.G != 0 || c.B != 0 {
		d.buffer[matrixRotations[d.rotation][row][col][rowIdx]][matrixRotations[d.rotation][row][col][colIdx]] = brightness(c.A)
	} else {
		d.buffer[matrixRotations[d.rotation][row][col][rowIdx]][matrixRotations[d.rotation][row][col][colIdx]] = 0
	}
}

const (
	brightnessLevels  = 9
	brightnessDivider = int8(255 / brightnessLevels)
)

var (
	Brightness0 = color.RGBA{R: 0, G: 0, B: 0, A: 0}
	Brightness1 = color.RGBA{R: 255, G: 255, B: 255, A: 255 - uint8(brightnessDivider)*1}
	Brightness2 = color.RGBA{R: 255, G: 255, B: 255, A: 255 - uint8(brightnessDivider)*2}
	Brightness3 = color.RGBA{R: 255, G: 255, B: 255, A: 255 - uint8(brightnessDivider)*3}
	Brightness4 = color.RGBA{R: 255, G: 255, B: 255, A: 255 - uint8(brightnessDivider)*4}
	Brightness5 = color.RGBA{R: 255, G: 255, B: 255, A: 255 - uint8(brightnessDivider)*5}
	Brightness6 = color.RGBA{R: 255, G: 255, B: 255, A: 255 - uint8(brightnessDivider)*6}
	Brightness7 = color.RGBA{R: 255, G: 255, B: 255, A: 255 - uint8(brightnessDivider)*7}
	Brightness8 = color.RGBA{R: 255, G: 255, B: 255, A: 255 - uint8(brightnessDivider)*8}
	Brightness9 = color.RGBA{R: 255, G: 255, B: 255, A: 255 - uint8(brightnessDivider)*9}

	BrightnessOff  = Brightness0
	BrightnessFull = Brightness9
)

func brightness(alpha uint8) int8 {
	return brightnessLevels - int8(alpha/uint8(brightnessDivider))
}

// GetPixel returns if the specific pixels is enabled.
func (d *Device) GetPixel(x int16, y int16) bool {
	if x < 0 || x >= 5 || y < 0 || y >= 5 {
		return false
	}
	col := x
	row := y
	return d.buffer[matrixRotations[d.rotation][row][col][rowIdx]][matrixRotations[d.rotation][row][col][colIdx]] > 0
}

const displayRefreshDelay = 8 * time.Millisecond

// Display sends the buffer (if any) to the screen.
func (d *Device) Display() error {
	var displayBuffer [ledRows][ledCols]int8
	for row := 0; row < ledRows; row++ {
		for col := 0; col < ledCols; col++ {
			displayBuffer[row][col] = d.buffer[row][col]
		}
	}

	for row := 0; row < ledRows; row++ {
		d.DisableAll()
		d.pin[ledCols+row].High()

		for col := 0; col < ledCols; col++ {
			if displayBuffer[row][col] > 0 {
				d.pin[col].Low()
			}
		}

		then := time.Now()
		var offset time.Duration = 0
		for _, ticks := range renderTimings {
			for time.Since(then).Nanoseconds() < int64(ticks*tickDuration+offset) {
				time.Sleep(offset / 10)
			}
			offset += ticks + tickDuration
			for col := 0; col < ledCols; col++ {
				displayBuffer[row][col]--
				if displayBuffer[row][col] <= 0 {
					d.pin[col].High()
				}
			}
		}
	}
	time.Sleep(displayRefreshDelay)
	return nil
}

// ClearDisplay erases the internal buffer.
func (d *Device) ClearDisplay() {
	for row := 0; row < ledRows; row++ {
		for col := 0; col < ledCols; col++ {
			d.buffer[row][col] = 0
		}
	}
}

// DisableAll disables all the LEDs without modifying the buffer.
func (d *Device) DisableAll() {
	for i := 0; i < ledCols; i++ {
		d.pin[i].High()
	}
	for i := 0; i < ledRows; i++ {
		d.pin[ledCols+i].Low()
	}
}

// EnableAll enables all the LEDs without modifying the buffer.
func (d *Device) EnableAll() {
	for i := 0; i < ledCols; i++ {
		d.pin[i].Low()
	}
	for i := 0; i < ledRows; i++ {
		d.pin[ledCols+i].High()
	}
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	return 5, 5
}
