package keypad4x4

import (
	"machine"
	"time"
)

// NoKeyPressed is used, when no key was pressed
const NoKeyPressed = 255

// settleTime is the wait after a row changes, before the columns are read.
// The row line does not fall immediately, because the wiring has some
// capacitance. A read that is too early can still show the previous row.
const settleTime = 50 * time.Microsecond

// defaultMapping gives the position of a key, 0 at the upper left end and 15
// at the lower right end.
var defaultMapping = [4][4]uint8{
	{0, 1, 2, 3},
	{4, 5, 6, 7},
	{8, 9, 10, 11},
	{12, 13, 14, 15},
}

// Config holds the options of a keypad.
type Config struct {
	// Mapping is the value that GetKey returns for each key, in reading
	// order. Use it to give the value printed on the key, such as '7' or 'A',
	// instead of its position. The zero value asks for the default mapping,
	// which is the position of the key.
	Mapping [4][4]uint8

	// Inverted swaps the polarity of the scan.
	//
	// The default is false. A column rests high with a pull-up, each row is
	// driven low in turn, and a pressed key pulls its column low.
	//
	// Set it to true when a column cannot rest high, because something on the
	// board pulls that pin down harder than the internal pull-up can hold it
	// up. A strapping pin such as GPIO12 on the ESP32 often has a pull-down
	// fitted, and some keypad modules bring their own resistors. A column
	// then rests low, each row is driven high, and a pressed key pulls its
	// column high.
	Inverted bool

	// ColumnConfig overrides how the column pins are configured.
	//
	// Leave it nil for the usual case. The driver then uses PinInputPullup,
	// or PinInput when Inverted is set.
	//
	// Inverted wants an internal pull-down, but not every chip has one. AVR
	// has pull-ups only, so the driver cannot name PinInputPulldown without
	// breaking those targets. Pass it here on a chip that has it:
	//
	//	pulldown := machine.PinConfig{Mode: machine.PinInputPulldown}
	//	cfg := keypad4x4.Config{Inverted: true, ColumnConfig: &pulldown}
	//
	// PinInput on its own is enough when the board already holds the columns
	// at their resting level.
	ColumnConfig *machine.PinConfig
}

// Device is used as 4x4 keypad driver
type Device interface {
	Configure()
	GetKey() uint8
	GetIndices() (int, int)
}

// device is a driver for 4x4 keypads
type device struct {
	inputEnabled bool
	lastColumn   int
	lastRow      int
	columns      [4]machine.Pin
	rows         [4]machine.Pin
	mapping      [4][4]uint8
	inverted     bool
	columnConfig *machine.PinConfig
}

// takes r4 -r1 pins and c4 - c1 pins
func NewDevice(r4, r3, r2, r1, c4, c3, c2, c1 machine.Pin) Device {
	return NewDeviceWithConfig(Config{}, r4, r3, r2, r1, c4, c3, c2, c1)
}

// NewDeviceWithConfig is NewDevice with the options in config.
func NewDeviceWithConfig(config Config, r4, r3, r2, r1, c4, c3, c2, c1 machine.Pin) Device {
	result := &device{}
	result.columns = [4]machine.Pin{c4, c3, c2, c1}
	result.rows = [4]machine.Pin{r4, r3, r2, r1}
	result.inverted = config.Inverted
	result.columnConfig = config.ColumnConfig

	// The mapping belongs to the device, not to Configure. Configure used to
	// write it, which would throw away a mapping given here.
	result.mapping = config.Mapping
	if result.mapping == ([4][4]uint8{}) {
		result.mapping = defaultMapping
	}

	return result
}

// Configure sets the column pins as input and the row pins as output
func (keypad *device) Configure() {
	// PinInputPullup exists on every chip. A pull-down does not, so the
	// inverted default is PinInput and the caller gives a pull-down through
	// ColumnConfig when the chip has one.
	inputConfig := machine.PinConfig{Mode: machine.PinInputPullup}
	if keypad.inverted {
		inputConfig = machine.PinConfig{Mode: machine.PinInput}
	}
	if keypad.columnConfig != nil {
		inputConfig = *keypad.columnConfig
	}
	for i := range keypad.columns {
		keypad.columns[i].Configure(inputConfig)
	}

	outputConfig := machine.PinConfig{Mode: machine.PinOutput}
	for i := range keypad.rows {
		keypad.rows[i].Configure(outputConfig)
		keypad.releaseRow(keypad.rows[i])
	}

	keypad.inputEnabled = true
	keypad.lastColumn = -1
	keypad.lastRow = -1
}

// GetKey returns the code for the given key.
// The codes start with 0 at the upper left end of the keypad and end with 15 at the lower right end of the keypad
// Example:
// 0	1	2	3
// 4	5	6	7
// 8	9	10	11
// 12	13	14	15
// returns 255 for no keyPressed
func (keypad *device) GetKey() uint8 {
	row, column := keypad.GetIndices()
	if row == -1 && column == -1 {
		return NoKeyPressed
	}

	return keypad.mapping[row][column]
}

// selectRow drives a row to the level that tests it.
func (keypad *device) selectRow(pin machine.Pin) {
	pin.Set(keypad.inverted)
}

// releaseRow puts a row back to its resting level.
func (keypad *device) releaseRow(pin machine.Pin) {
	pin.Set(!keypad.inverted)
}

// pressed reports whether a column reads as a key held down.
func (keypad *device) pressed(pin machine.Pin) bool {
	return pin.Get() == keypad.inverted
}

// GetIndices returns the position of the pressed key
func (keypad *device) GetIndices() (int, int) {
	for rowIndex, rowPin := range keypad.rows {
		keypad.selectRow(rowPin)
		time.Sleep(settleTime)

		for columnIndex := range keypad.columns {
			columnPin := keypad.columns[columnIndex]

			if keypad.pressed(columnPin) && keypad.inputEnabled {
				keypad.inputEnabled = false

				keypad.lastColumn = columnIndex
				keypad.lastRow = rowIndex

				// Stop driving this row before leaving. A row left selected
				// makes two rows selected on the next scan, and a column then
				// does not say which row it came from.
				keypad.releaseRow(rowPin)

				return keypad.lastRow, keypad.lastColumn
			}

			if !keypad.pressed(columnPin) &&
				columnIndex == keypad.lastColumn &&
				rowIndex == keypad.lastRow &&
				!keypad.inputEnabled {
				keypad.inputEnabled = true
			}
		}

		keypad.releaseRow(rowPin)
	}

	return -1, -1
}
