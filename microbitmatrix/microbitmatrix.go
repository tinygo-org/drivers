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
//     0: regular orientation, (0 degree rotation)
//     1: 90 degree rotation clock wise
//     2: 180 degree rotation clock wise
//     3: 270 degree rotation clock wise
func (d *Device) SetRotation(rotation uint8) {
	d.rotation = rotation % 4
}

const (
	rowIdx = 0
	colIdx = 1
)

// SetPixel modifies the internal buffer in a single pixel.
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	if x < 0 || x >= 5 || y < 0 || y >= 5 {
		return
	}
	col := x
	row := y
	if c.R != 0 || c.G != 0 || c.B != 0 {
		d.buffer[matrixRotations[d.rotation][row][col][rowIdx]][matrixRotations[d.rotation][row][col][colIdx]] = true
	} else {
		d.buffer[matrixRotations[d.rotation][row][col][rowIdx]][matrixRotations[d.rotation][row][col][colIdx]] = false
	}
}

// GetPixel returns if the specific pixels is enabled.
func (d *Device) GetPixel(x int16, y int16) bool {
	if x < 0 || x >= 5 || y < 0 || y >= 5 {
		return false
	}
	col := x
	row := y
	return d.buffer[matrixRotations[d.rotation][row][col][rowIdx]][matrixRotations[d.rotation][row][col][colIdx]]
}

const displayRefreshDelay = 8 * time.Millisecond

// Display sends the buffer (if any) to the screen.
func (d *Device) Display() error {
	for row := 0; row < ledRows; row++ {
		d.DisableAll()
		d.pin[ledCols+row].Low()

		for col := 0; col < ledCols; col++ {
			if d.buffer[row][col] {
				d.pin[col].High()
			} else {
				d.pin[col].Low()
			}
		}
		time.Sleep(time.Millisecond * 4)
	}
	return nil
}

// ClearDisplay erases the internal buffer.
func (d *Device) ClearDisplay() {
	for row := 0; row < ledRows; row++ {
		for col := 0; col < ledCols; col++ {
			d.buffer[row][col] = false
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
