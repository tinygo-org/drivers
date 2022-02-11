package is31fl3731

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
