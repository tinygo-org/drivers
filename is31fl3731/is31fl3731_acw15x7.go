package is31fl3731

import (
	"fmt"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// DeviceAdafruitCharlieWing15x7 implements TinyGo driver for Lumissil
// IS31FL3731 matrix LED driver on Adafruit 15x7 CharliePlex LED Matrix
// FeatherWing (CharlieWing) board: https://www.adafruit.com/product/3163
type DeviceAdafruitCharlieWing15x7 struct {
	Device
}

// enableLEDs enables only LEDs that are soldered on the Adafruit CharlieWing
// board. The board has following LEDs matrix layout:
//
//	"o" - connected (soldered) LEDs
//	"x" - not connected LEDs
//
//	  + - - - - - - - - - - - - - - +
//	  | + - - - - - - - - - - - - + |
//	  | |                         | |
//	  | |                         v v
//	+---------------------------------+
//	| o o o o o o o o o o o o o o o x |
//	| o o o o o o o o o o o o o o o x |
//	| o o o o o o o o o o o o o o o x |
//	| o o o o o o o o o o o o o o o x |
//	| o o o o o o o o o o o o o o o x |
//	| o o o o o o o o o o o o o o o x |
//	| o o o o o o o o o o o o o o o x |
//	| x x x x x x x x x x x x x x x x |
//	+---------------------------------+
//	  ^ ^                         | |
//	  | |                 ... - - + |
//	  | + - - - - - - - - - - - - - +
//	  |
//	  start (address 0x00)
func (d *DeviceAdafruitCharlieWing15x7) enableLEDs() (err error) {
	for frame := FRAME_0; frame <= FRAME_7; frame++ {
		err = d.selectCommand(frame)
		if err != nil {
			return err
		}

		// Enable left half
		for i := uint8(0); i < 16; i += 2 {
			err = legacy.WriteRegister(d.bus, d.Address, i, []byte{0b11111110})
			if err != nil {
				return err
			}
		}
		// Enable right half
		for i := uint8(3); i < 16; i += 2 {
			err = legacy.WriteRegister(d.bus, d.Address, i, []byte{0b01111111})
			if err != nil {
				return err
			}
		}
		// Disable invisible column on the right side
		err = legacy.WriteRegister(d.bus, d.Address, 1, []byte{0b00000000})
		if err != nil {
			return err
		}
	}

	return nil
}

// DrawPixelXY draws a single pixel on the selected frame by its XY coordinates
// with provided PWM value [0-255]
func (d *DeviceAdafruitCharlieWing15x7) DrawPixelXY(frame, x, y, value uint8) (err error) {
	var index uint8

	if x >= 15 {
		return fmt.Errorf("invalid value: X is out of range [0, 15]")
	} else if y >= 7 {
		return fmt.Errorf("invalid value: Y is out of range [0, 7]")
	}

	// Board is one pixel shorter (7 vs 8 supported pixels)
	if x < 8 {
		index = 16*x + y + 1
	} else {
		index = 16*(16-x) - y - 1 - 1
	}

	return d.setPixelPWD(frame, index, value)
}

// NewAdafruitCharlieWing15x7 creates a new driver with Adafruit 15x7
// CharliePlex LED Matrix FeatherWing (CharlieWing) layout.
// Available addresses:
// - 0x74 (default)
// - 0x77 (when the address jumper soldered)
func NewAdafruitCharlieWing15x7(bus drivers.I2C, address uint8) DeviceAdafruitCharlieWing15x7 {
	return DeviceAdafruitCharlieWing15x7{
		Device: Device{
			Address: address,
			bus:     bus,
		},
	}
}
