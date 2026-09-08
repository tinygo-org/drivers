// Package rgb75 implements a driver for the HUB75 LED matrix.
//
// Unlike package hub75 which takes advantage of a high-speed SPI peripheral for
// both clock and data, this package uses individually-addressable GPIO for the
// column-select (R,G,B) lines.
//
// Extra features:
//   - Matrix chaining (connecting multiple displays to form one large matrix)
//   - Double-buffering (requires more RAM!)
//   - Efficient GPIO access (not natively provided by TinyGo machine package)
//
// Designed for and implemented with an Adafruit MatrixPortal M4 (SAMD51). Care
// was taken to separate the HUB75 logic from target-specific code so that other
// devices and architectures can be added easily. This package contains only the
// portable HUB75 logic. The required hardware interface and all target-specific
// units implementing this interface can be found in rgb75/native.
//
// Note that you must carefully select which GPIO pins on your target device are
// connected to the HUB75 matrix interface. If using the onboard HUB75 connector
// on an Adafruit MatrixPortal M4, then these conditions are already satisfied.
//   REQUIRED:
//     - All six (6) HUB75 RGB data lines are on a single, common GPIO port.
//   OPTIONAL (improves performance):
//     - All HUB75 row-address select lines are on a single, common GPIO port.
//     - The HUB75 CLK line is on the same port as RGB data lines.
//
// Inspired by the Adafruit_Protomatter library for Arduino, written by:
//   - Phil "Paint Your Dragon" Burgess
//   - Jeff Epler
// See their original project:
// 		https://github.com/adafruit/Adafruit_Protomatter
//
// TODO:
//   To reduce the number of port accesses (slightly), we could use the write-
//   only GPIO toggle registers if the MCU supports it (e.g., OUTTGL of SAMD51
//   PORT_Type). This requires a small (as in memory) overhead to implement, but
//   it does make the logic more complicated, because we need to keep track of
//   the last 32-bit value written to these write-only port registers.
package rgb75 // import "tinygo.org/x/drivers/rgb75"

import (
	"errors"
	"image/color"
	"machine"

	"tinygo.org/x/drivers/rgb75/native"
)

var (
	ErrInvalidDataPins = errors.New("RGB data pins must be on a common GPIO port")
	ErrInvalidHeight   = errors.New("invalid matrix height for given number of row address pins")
)

var ClearColor = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00}

// Default configuration settings for a Device.
const (
	DefaultWidth      = 64 // (pixels) default total width of matrix chain
	DefaultHeight     = 32 // (pixels) default total height of matrix chain
	DefaultColorDepth = 4  // (bits) default color depth of each R,G,B component

	bitPeriod = 1800 // timer period at bitplane index 0
)

// Config holds the configuration settings for a Device.
type Config struct {
	Width      int   // (pixels) total width of matrix chain
	Height     int   // (pixels) total height of matrix chain
	ColorDepth uint8 // (bits) color depth of each R,G,B component
	DoubleBuf  bool  // use double-buffering to reduce flicker

	oneAddrPort bool // all address pins are on a single GPIO port
	clkDataPort bool // RGB and CLK pins are all on a single GPIO port
	numAddrRows int  // number of addressable rows
	maxHeight   int  // (pixels) maximum height given number of row address pins
}

// Device represents a connection to a chain of one or more RGB LED matrix
// panels (HUB75).
type Device struct {
	cfg Config           // configuration settings
	hub native.Hub75     // HUB75 connection
	oen machine.Pin      // output enable pin (active low)
	lat machine.Pin      // RGB data latch pin
	clk machine.Pin      // RGB clock pin
	rgb dataPins         // all (6) RGB data pins
	row []machine.Pin    // slice of all row address pins
	buf [][][]color.RGBA // panel framebuffers (indexed by: [buf][col][row])
	fbs bufferState      // double-buffering enabled
	pos rowPlane         // current row/bitplane of ISR
	val uint32           // current timer position
}

type (
	// rgbPins holds one set of GPIO pins (3) for the RGB data lines on a HUB75
	// connector (upper-half OR lower-half of matrix).
	rgbPins struct{ r, g, b machine.Pin }
	// dataPins holds all GPIO pins (6) for the two sets of RGB data lines on a
	// HUB75 connector (upper-half AND lower-half of matrix).
	dataPins struct{ up, lo rgbPins }
	// rowPlane holds the current rows and bitplane of the row-scan state machine.
	rowPlane struct {
		frame         int    // frame index
		yPr, yUp, yLo int    // previous-upper, upper, and lower row index
		bit           int    // bitplane index
		per           uint32 // timer period for current bitplane index
	}
)

// bufferState holds the current state of the framebuffer for double-buffering
// support.
type bufferState struct {
	// double is true if and only if double-buffering is enabled
	double bool
	// - The front-buffer (fg) is actively being displayed;
	// 		i.e., The front-buffer (fg) is latched by the RGB LED drivers in the
	// 					main row-scan timer's interrupt handler, (*Device).handleRow.
	// - The back-buffer (bg) is an off-screen canvas where updates to the screen
	//   are performed.
	fg, bg int // current front- and back-buffer indices
	// Once all updates are complete (i.e., an entire frame has been drawn), the
	// front- and back-buffers are swapped, so that the row-scan ISR mentioned
	// above receives a complete screen instantly and never draws a partially-
	// updated frame (which causes flickering).
	// The front- and back-buffers are swapped in method (*bufferState).swap.
}

// swap swaps the front- and back-buffer indices of the receiver bufferState s.
//
// This instantly changes which framebuffer content is latched by the RGB LED
// drivers in the main row-scan timer's interrupt handler, (*Device).handleRow.
func (s *bufferState) swap() { s.fg, s.bg = s.bg, s.fg }

// New returns a new HUB75 driver. The returned Device must be initialized with
// method Configure before it can be used.
//
// rgb is a 6-element Pin array, corresponding to the color bits for each pair
// of RGB pins (upper-half & lower-half), ordered as: upper-red, -green, -blue,
// lower-red, -green, -blue.
//
// ** Note that all 6 RGB data pins should be on the same GPIO port. **
//
// row refers to each of the address lines for selecting the active data row.
// The length of this slice N is determined by the total height of the matrix
// chain (in pixels): Height = 2^(N+1). For example, a matrix with height 16px
// must provide slice row of length 3 (= log2(16)-1); 32px = 4; 64px = 5; etc.
//
// There is no GPIO port restriction for the row address control pins (they can
// be spread among different GPIO ports), but performance is improved when they
// are all on the same port.
func New(oen, lat, clk machine.Pin, rgb [6]machine.Pin, row []machine.Pin) *Device {
	native.HUB75.SetPins(rgb, clk, row...)
	return &Device{
		cfg: Config{
			Width:      DefaultWidth,
			Height:     DefaultHeight,
			ColorDepth: DefaultColorDepth,
			// maxHeight is computed from the number of row address lines given, as we
			// cannot refer to any rows higher than we have address lines available.
			maxHeight: 1 << (len(row) + 1),
		},
		hub: native.HUB75,
		oen: oen,
		lat: lat,
		clk: clk,
		rgb: dataPins{
			up: rgbPins{r: rgb[0], g: rgb[1], b: rgb[2]},
			lo: rgbPins{r: rgb[3], g: rgb[4], b: rgb[5]},
		},
		row: row,
		buf: nil,
		fbs: bufferState{},
		pos: rowPlane{},
		val: 0,
	}
}

// Configure initializes all GPIO pins and Device settings, and allocates the
// display framebuffer.
//
// An error may be returned if invalid configuration is detected.
func (d *Device) Configure(cfg Config) error {

	// Configure total panel width (in pixels).
	if 0 != cfg.Width {
		d.cfg.Width = cfg.Width // use given width without restriction
	} else {
		d.cfg.Width = DefaultWidth // use default width when undefined
	}

	// Configure total panel height (in pixels).
	// The value `d.cfg.maxHeight` is used in several locations here. For details
	// on its purpose and validity, see the comments above its assignment inside
	// of method `(*Device).New`.
	if 0 != cfg.Height {
		if cfg.Height > d.cfg.maxHeight {
			// Bail out with error if given height exceeds maximum height. Otherwise,
			// entire rows may get dropped, or, worse, row index might wrap around and
			// overwrite correct rows
			return ErrInvalidHeight
		}
		d.cfg.Height = cfg.Height // use given height if it passes all restrictions
	} else {
		d.cfg.Height = d.cfg.maxHeight // use maximum height if undefined
	}
	// use the final height selection (H) to determine number of row pairs (H/2),
	// which is the number of iterations required to scan all matrix rows.
	d.cfg.numAddrRows = d.cfg.Height / 2

	// Configure color depth of each R,G,B component (in bits).
	if 0 != cfg.ColorDepth {
		d.cfg.ColorDepth = cfg.ColorDepth // use given depth without restriction
	} else {
		d.cfg.ColorDepth = DefaultColorDepth // use default depth when undefined
	}

	// decide if all row address lines are on the same GPIO port, which isn't a
	// requirement, but it will improve performance by efficiently setting row
	// address with a single register write.
	d.cfg.oneAddrPort, _ = d.hub.GetPinGroupAlignment(d.row...)

	// verify all RGB data pins are on the same GPIO port
	same, align := d.hub.GetPinGroupAlignment(
		d.rgb.up.r, d.rgb.up.g, d.rgb.up.b,
		d.rgb.lo.r, d.rgb.lo.g, d.rgb.lo.b)
	if !same || align == 0 {
		return ErrInvalidDataPins
	}

	// decide if CLK and RGB data lines are all on the same GPIO port, which helps
	// further increase efficiency when writing RGB data to the shift registers.
	// we verified above that all RGB data lines are on the same port, so we need
	// to compare CLK to only one of those pins (any one is fine).
	d.cfg.clkDataPort, _ = d.hub.GetPinGroupAlignment(d.rgb.up.r, d.clk)

	// configure all of our Device pins for GPIO output
	d.oen.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.lat.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.clk.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.rgb.up.r.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.rgb.up.g.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.rgb.up.b.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.rgb.lo.r.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.rgb.lo.g.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.rgb.lo.b.Configure(machine.PinConfig{Mode: machine.PinOutput})
	for i := range d.row {
		d.row[i].Configure(machine.PinConfig{Mode: machine.PinOutput})
	}

	// configure the framebuffer state
	var numBuffers int
	if d.fbs.double = cfg.DoubleBuf; d.fbs.double {
		d.fbs.fg, d.fbs.bg = 0, 1
		numBuffers = 2
	} else {
		d.fbs.fg, d.fbs.bg = 0, 0 // keep these equal so that swap() has no effect
		numBuffers = 1
	}

	// allocate the framebuffer(s)
	d.buf = make([][][]color.RGBA, numBuffers)
	for n := range d.buf {
		d.buf[n] = make([][]color.RGBA, d.cfg.Height)
		for y := range d.buf[n] {
			d.buf[n][y] = make([]color.RGBA, d.cfg.Width)
		}
	}

	return d.initialize()
}

// Size returns the current size of the display.
func (d *Device) Size() (x, y int16) {
	return int16(d.cfg.Width), int16(d.cfg.Height)
}

// SetPixel sets the color of a pixel in the framebuffer.
// If double-buffering is enabled, SetPixel writes to the back-buffer, which
// then requires a call to Display for the change to appear.
func (d *Device) SetPixel(x, y int16, c color.RGBA) {
	// if double-buffering is not enabled, then d.fbs.fg == d.fbs.bg; this means
	// the active front-buffer is modified and the change appears immediately
	// without a subsequent call to Display.
	if y >= 0 && int(y) < len(d.buf[d.fbs.bg]) {
		if x >= 0 && int(x) < len(d.buf[d.fbs.bg][y]) {
			d.buf[d.fbs.bg][y][x] = c
		}
	}
}

// GetPixel returns the color of a pixel in the framebuffer.
// If double-buffering is enabled, GetPixel returns the color of a pixel from
// the active front-buffer, which is what currently exists on the display.
func (d *Device) GetPixel(x, y int16) color.RGBA {
	// if double-buffering is not enabled, then d.fbs.fg == d.fbs.bg; this means
	// the active front-buffer is the only source of pixel colors.
	if y >= 0 && int(y) < len(d.buf[d.fbs.fg]) {
		if x >= 0 && int(x) < len(d.buf[d.fbs.fg][y]) {
			return d.buf[d.fbs.fg][y][x]
		}
	}
	return ClearColor
}

// Display draws the current framebuffer to the display.
// If double-buffering is enabled, Display swaps the front- and back-buffer so
// that the previous back-buffer is activated and drawn to the display, and the
// previous front-buffer is cleared and ready to be written into.
func (d *Device) Display() error {
	if d.fbs.double {
		d.fbs.swap()
		d.clearBuffer(d.fbs.bg)
	} else {
		// double-buffering is not enabled, so just re-enable the row-scan timer (in
		// case it isn't active) to resume screen updates.
		d.Resume()
	}
	return nil
}

// ClearDisplay clears the display.
// If double-buffering is enabled, ClearDisplay clears both the front- and back-
// buffer so that the change is immediate and all framebuffer content is erased.
func (d *Device) ClearDisplay() {
	for n := range d.buf {
		d.clearBuffer(n)
	}
}

// Resume starts or restarts updating the display.
func (d *Device) Resume() {
	d.hub.ResumeTimer(d.val, d.pos.per)
}

// Pause stops updating the display. Use Resume to restart updates.
func (d *Device) Pause() {
	d.val = d.hub.PauseTimer()
}

// initialize initializes all GPIO pin levels and Device state machines prior to
// starting the display.
func (d *Device) initialize() error {

	// initialize pin states
	d.oen.High() // set high to disable output (active low)
	d.lat.Low()  // hold all control and data lines low during init
	d.clk.Low()
	d.rgb.up.r.Low()
	d.rgb.up.g.Low()
	d.rgb.up.b.Low()
	d.rgb.lo.r.Low()
	d.rgb.lo.g.Low()
	d.rgb.lo.b.Low()
	for i := range d.row {
		d.row[i].Low()
	}

	// We can also clear the shift registers by clocking out 0-bits across the
	// entire width of the matrix chain and latching their content. Simply leave
	// all RGB data lines low (see above) during this time.
	for i := 0; i < d.cfg.Width; i++ {
		d.clk.High()
		d.clk.Low()
	}
	d.lat.High()
	d.lat.Low()

	// configure starting indices so that they rollover on first interrupt.
	d.pos = rowPlane{
		frame: 0,
		yPr:   1, // invalid row to force selectRow to set address lines
		yUp:   d.cfg.numAddrRows,
		yLo:   d.cfg.Height,
		bit:   int(d.cfg.ColorDepth),
		per:   bitPeriod,
	}
	d.ClearDisplay()
	d.hub.InitTimer(d.handleRow)

	return nil
}

// rgbBit returns the n'th bit (LSB) of each R, G, B component of the color in
// the receiver's framebuffer at column x and row y.
//
// Note that for performance efficiency, the arguments are NOT validated or
// range-checked. So be very careful you are providing valid inputs, otherwise
// this is a rather dangerous function susceptible to access violations!
func (d *Device) rgbBit(x, y, n int) (r, g, b bool) {
	cr, cg, cb, _ := d.buf[d.fbs.fg][y][x].RGBA()
	r = 0 != (cr & (1 << n))
	g = 0 != (cg & (1 << n))
	b = 0 != (cb & (1 << n))
	return
}

// handleRow is the interrupt service routine (ISR) for the main HUB75 row-scan
// timer, which handles row address selection and row data transmission.
func (d *Device) handleRow() {

	d.hub.PauseTimer()
	d.hub.ResumeTimer(0, d.pos.per)

	// disable output while we modify the LED output (column) drivers, and open
	// the LED output (column) latch with data that was transmitted to the shift
	// registers during previous interrupt. this data could be either:
	//   a) new row; i.e., illuminating the next pair of rows different from
	//      previously illuminated pair of rows (initial bitplane)
	//   b) new bitplane; i.e., illuminating the same pair of rows for twice the
	//      duration as previously illuminated (binary code modulation)
	d.oen.High()
	d.lat.High()

	// stop the row select timer, switch rows if we have incremented to a new row,
	// and then re-enable the row select timer.
	d.selectRow(d.pos.yUp)
	d.increment()

	// close the latch before clocking out the next row of data, and enable output
	d.lat.Low()
	d.oen.Low()

	// pulse color data to the next pair of rows while we wait for the timer
	for x := 0; x < d.cfg.Width; x++ {
		// for the current rows (d.pos.yUp/yLo) and current bitplane (d.pos.bit),
		// grab the corresponding bit in each R,G,B color component of the pixel in
		// column x.
		r1, g1, b1 := d.rgbBit(x, d.pos.yUp, d.pos.bit) // get upper row
		r2, g2, b2 := d.rgbBit(x, d.pos.yLo, d.pos.bit) // get lower row

		// check if we can set both RGB data and CLK at the same time.
		if d.cfg.clkDataPort {
			// set/clear all 6 data lines and CLK with a single register write.
			d.hub.ClkRgb(r1, g1, b1, r2, g2, b2)
		} else {
			// set/clear all 6 data lines with a single register write
			d.hub.SetRgb(r1, g1, b1, r2, g2, b2)

			// clock out one bit of data for the two current pixels in our active
			// bitplane (d.pos.bit): (x1,y1)=(x,d.pos.yUp), (x2,y2)=(x,d.pos.yLo)
			d.clk.High()
			d.clk.Low()

			// reset all 6 data lines (after data has been clocked out) with a single
			// register write.
			d.hub.SetRgbMask(0)
		}
	}
}

// increment updates the active row and bitplane indices by one.
func (d *Device) increment() {
	d.pos.bit++    // increment bitplane index
	d.pos.per *= 2 // double timer period
	// check for bitplane index rollover
	if d.pos.bit >= int(d.cfg.ColorDepth) {
		d.pos.bit = 0         // reset bitplane index
		d.pos.per = bitPeriod // reset timer period
		d.pos.yUp++           // update upper row index
		d.pos.yLo++           // update lower row index
		// check for upper/lower row index rollover
		if d.pos.yUp >= d.cfg.numAddrRows {
			d.pos.yUp = 0                 // reset upper row index
			d.pos.yLo = d.cfg.numAddrRows // reset lower row index
			d.pos.frame++                 // update frame index
		}
	}
}

// selectRow configures the row address control lines, which selects the active
// pair of rows to drive.
//
// Since two rows (upper- and lower-half) are always driven simultaneously,
// either one may be given. For example, with 4 address control lines (matrix
// height = 32px), providing row=3 is equivalent to row=19, as either one of
// these arguments will drive both of these rows.
func (d *Device) selectRow(row int) {
	// don't do anything if given row index exceeds total matrix height
	if row >= d.cfg.Height {
		return
	}
	// if given index refers to a row in the lower-half of the matrix, translate
	// to its corresponding row index in the upper-half.
	if row >= d.cfg.numAddrRows {
		row -= d.cfg.numAddrRows
	}
	// don't do anything if the row index is the same as previously selected
	if row == d.pos.yPr {
		return
	}
	d.pos.yPr = row // update previous row selection
	// check if all address control lines are on the same GPIO port
	if d.cfg.oneAddrPort {
		// perform the address change with a single register write
		d.hub.SetRow(row)
	} else {
		// otherwise, set/clear each row address bit individually
		for i := 0; i < len(d.row); i++ {
			// for each address line i, set high IFF the i'th bit in row is set
			d.row[i].Set(row&(1<<i) != 0)
		}
	}
}

// clearBuffer writes a clear pixel to all elements of the framebuffer, at given
// index n, of the receiver Device d.
func (d *Device) clearBuffer(n int) {
	if n >= 0 && n < len(d.buf) {
		for y := range d.buf[n] {
			for x := range d.buf[n][y] {
				d.buf[n][y][x] = ClearColor
			}
		}
	}
}
