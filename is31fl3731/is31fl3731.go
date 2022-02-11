// Package is31fl3731 provides a driver for the Lumissil IS31FL3731 matrix LED
// driver.
//
// Driver supports following layouts:
//   - any custom LED matrix layout
//   - Adafruit 15x7 CharliePlex LED Matrix FeatherWing (CharlieWing)
//     https://www.adafruit.com/product/3163
//
// Datasheet:
//    https://www.lumissil.com/assets/pdf/core/IS31FL3731_DS.pdf
//
// This driver inspired by Adafruit Python driver:
//    https://github.com/adafruit/Adafruit_CircuitPython_IS31FL3731
//
package is31fl3731

import (
	"fmt"
	"time"

	"tinygo.org/x/drivers"
)

// board keeps reference to all implemented LED matrix boards
type board int

const (
	// Raw LEDs layout assumed to be 16x9 matrix, but can be used with any custom
	// board that has IS31FL3731 driver
	boardRaw board = iota

	// Adafruit 15x7 CharliePlex LED Matrix FeatherWing (CharlieWing):
	// https://www.adafruit.com/product/3163
	//
	// The LED bits order:
	//
	//   "o" - connected (soldered) LEDs
	//   "x" - not connected LEDs
	//
	//     + - - - - - - - - - - - - - - +
	//     | + - - - - - - - - - - - - + |
	//     | |                         | |
	//     | |                         v v
	//   +---------------------------------+
	//   | o o o o o o o o o o o o o o o x |
	//   | o o o o o o o o o o o o o o o x |
	//   | o o o o o o o o o o o o o o o x |
	//   | o o o o o o o o o o o o o o o x |
	//   | o o o o o o o o o o o o o o o x |
	//   | o o o o o o o o o o o o o o o x |
	//   | o o o o o o o o o o o o o o o x |
	//   | x x x x x x x x x x x x x x x x |
	//   +---------------------------------+
	//     ^ ^                         | |
	//     | |                 ... - - + |
	//     | + - - - - - - - - - - - - - +
	//     |
	//     start (address 0x00)
	//
	boardAdafruitCharlieWing15x7
)

// Device implements TinyGo driver for Lumissil IS31FL3731 matrix LED driver
type Device struct {
	Address uint8
	bus     drivers.I2C
	board   board

	// Currently selected command register (one of the frame registers or the
	// function register)
	selectedCommand uint8
}

// Configure chip for operating as a LED matrix display
func (d *Device) Configure() (err error) {
	// Shutdown software
	err = d.writeFunctionRegister(SET_SHUTDOWN, []byte{SOFTWARE_OFF})
	if err != nil {
		return fmt.Errorf("failed to shutdown: %w", err)
	}

	time.Sleep(time.Millisecond * 10)

	// Wake up software
	err = d.writeFunctionRegister(SET_SHUTDOWN, []byte{SOFTWARE_ON})
	if err != nil {
		return fmt.Errorf("failed to wake up: %w", err)
	}

	// Set display to a picture mode ("auto frame play mode" and "audio frame play
	// mode" are not supported in this version of the driver)
	err = d.writeFunctionRegister(SET_DISPLAY_MODE, []byte{DISPLAY_MODE_PICTURE})
	if err != nil {
		return fmt.Errorf("failed to switch to a picture move: %w", err)
	}

	// Enable LEDs that are present (soldered) on the board. From the datasheet:
	// LEDs which are no connected must be off by LED Control Register (Frame
	// Registers) or it will affect other LEDs
	err = d.enableLEDs()
	if err != nil {
		return fmt.Errorf("failed to enable LEDs: %w", err)
	}

	// Disable audiosync
	err = d.writeFunctionRegister(SET_AUDIOSYNC, []byte{AUDIOSYNC_OFF})
	if err != nil {
		return fmt.Errorf("failed to disable audiosync: %w", err)
	}

	// Clear all frames
	for frame := FRAME_0; frame <= FRAME_7; frame++ {
		err = d.Clear(frame)
		if err != nil {
			return fmt.Errorf("failed to clear frame %d: %w", frame, err)
		}
	}

	// 1st frame is displayed by default
	err = d.SetActiveFrame(FRAME_0)
	if err != nil {
		return fmt.Errorf("failed to set active frame: %w", err)
	}

	return nil
}

// selectCommand selects command register, can be:
// - frame registers 0-7
// - function register
func (d *Device) selectCommand(command uint8) (err error) {
	if command != d.selectedCommand {
		d.selectedCommand = command
		return d.bus.WriteRegister(d.Address, COMMAND, []byte{command})
	}

	return nil
}

// writeFunctionRegister selects the function register and writes data into it
func (d *Device) writeFunctionRegister(operation uint8, data []byte) (err error) {
	err = d.selectCommand(FUNCTION)
	if err != nil {
		return err
	}

	return d.bus.WriteRegister(d.Address, operation, data)
}

// enableLEDs enables only LEDs that are soldered on the set board:
// - 15x7 -- matrix for Adafruit CharlieWing
// - 16x9 -- enable all leds
func (d *Device) enableLEDs() (err error) {
	for frame := FRAME_0; frame <= FRAME_7; frame++ {
		err = d.selectCommand(frame)
		if err != nil {
			return err
		}

		if d.board == boardAdafruitCharlieWing15x7 {
			// Enable left half
			for i := uint8(0); i < 16; i += 2 {
				err = d.bus.WriteRegister(d.Address, i, []byte{0b11111110})
				if err != nil {
					return err
				}
			}
			// Enable right half
			for i := uint8(3); i < 16; i += 2 {
				err = d.bus.WriteRegister(d.Address, i, []byte{0b01111111})
				if err != nil {
					return err
				}
			}
			// Disable invisible column on the right side
			err = d.bus.WriteRegister(d.Address, 1, []byte{0b00000000})
			if err != nil {
				return err
			}
		} else {
			// Enable every LED (16 columns x 9 rows)
			for i := uint8(0); i < 16; i++ {
				err = d.bus.WriteRegister(d.Address, i, []byte{0xFF})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// setPixelPWD sets individual pixel's PWM value [0-255] on the selected frame
func (d *Device) setPixelPWD(frame, n, value uint8) (err error) {
	err = d.selectCommand(frame)
	if err != nil {
		return err
	}

	return d.bus.WriteRegister(d.Address, LED_PWM_OFFSET+n, []byte{value})
}

// SetActiveFrame sets frame to display with LEDs
func (d *Device) SetActiveFrame(frame uint8) (err error) {
	if frame > FRAME_7 {
		return fmt.Errorf("frame %d is out of valid range [0-7]", frame)
	}

	return d.writeFunctionRegister(SET_ACTIVE_FRAME, []byte{frame})
}

// Fill the whole frame with provided PWM value [0-255]
func (d *Device) Fill(frame, value uint8) (err error) {
	if frame > FRAME_7 {
		return fmt.Errorf("frame %d is out of valid range [0-7]", frame)
	}

	err = d.selectCommand(frame)
	if err != nil {
		return err
	}

	data := make([]byte, 24)
	for i := range data {
		data[i] = value
	}

	for i := uint8(0); i < 6; i++ {
		err = d.bus.WriteRegister(d.Address, LED_PWM_OFFSET+i*24, data)
		if err != nil {
			return err
		}
	}

	return nil
}

// Clear the whole frame
func (d *Device) Clear(frame uint8) (err error) {
	return d.Fill(frame, 0x00)
}

// DrawPixelIndex draws a single pixel on the selected frame by its index with
// provided PWM value [0-255]
func (d *Device) DrawPixelIndex(frame, index, value uint8) (err error) {
	if frame > FRAME_7 {
		return fmt.Errorf("frame %d is out of valid range [0-7]", frame)
	}

	return d.setPixelPWD(frame, index, value)
}

// DrawPixelXY draws a single pixel on the selected frame by its XY coordinates
// with provided PWM value [0-255]
func (d *Device) DrawPixelXY(frame, x, y, value uint8) (err error) {
	var index uint8

	if d.board == boardAdafruitCharlieWing15x7 {
		if x >= 15 {
			return fmt.Errorf("invalid value: X is out of range [0, 15]")
		} else if y >= 7 {
			return fmt.Errorf("invalid value: Y is out of range [0, 7]")
		}

		if x < 8 {
			index = 16*x + y + 1
		} else { //          ^-- board is one pixel shorter (7 vs 8 pixels)
			index = 16*(16-x) - y - 1 - 1
		} //                      ^-- board is one pixel shorter (7 vs 8 pixels)
	} else {
		index = 16*x + y
	}

	return d.setPixelPWD(frame, index, value)
}

// New creates a raw driver w/o any preset board layout
func New(bus drivers.I2C, address uint8) (d Device, err error) {
	d = Device{
		Address: address,
		bus:     bus,
		board:   boardRaw,
	}

	err = d.configure()

	return d, err
}

// NewAdafruitCharlieWing15x7 creates a new driver with Adafruit 15x7
// CharliePlex LED Matrix FeatherWing (CharlieWing) layout
func NewAdafruitCharlieWing15x7(bus drivers.I2C, address uint8) (d Device, err error) {
	d = Device{
		Address: address,
		bus:     bus,
		board:   boardAdafruitCharlieWing15x7,
	}

	err = d.configure()

	return d, err
}
