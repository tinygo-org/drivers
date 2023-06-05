// Package st7789 implements a driver for the ST7789 TFT displays, it comes in various screen sizes.
//
// Datasheets: https://cdn-shop.adafruit.com/product-files/3787/3787_tft_QT154H2201__________20190228182902.pdf
//
//	http://www.newhavendisplay.com/appnotes/datasheets/LCDs/ST7789V.pdf
package st7789 // import "tinygo.org/x/drivers/st7789"

import (
	"image/color"
	"machine"
	"math"
	"time"

	"errors"

	"tinygo.org/x/drivers"
)

// Rotation controls the rotation used by the display.
//
// Deprecated: use drivers.Rotation instead.
type Rotation = drivers.Rotation

// The color format used on the display, like RGB565, RGB666, and RGB444.
type ColorFormat uint8

// FrameRate controls the frame rate used by the display.
type FrameRate uint8

var (
	errOutOfBounds = errors.New("rectangle coordinates outside display area")
)

// Device wraps an SPI connection.
type Device struct {
	bus             drivers.SPI
	dcPin           machine.Pin
	resetPin        machine.Pin
	csPin           machine.Pin
	blPin           machine.Pin
	width           int16
	height          int16
	columnOffsetCfg int16
	rowOffsetCfg    int16
	columnOffset    int16
	rowOffset       int16
	rotation        drivers.Rotation
	frameRate       FrameRate
	batchLength     int32
	isBGR           bool
	vSyncLines      int16
	cmdBuf          [1]byte
	buf             [6]byte
}

// Config is the configuration for the display
type Config struct {
	Width        int16
	Height       int16
	Rotation     drivers.Rotation
	RowOffset    int16
	ColumnOffset int16
	FrameRate    FrameRate
	VSyncLines   int16

	// Gamma control. Look in the LCD panel datasheet or provided example code
	// to find these values. If not set, the defaults will be used.
	PVGAMCTRL []uint8 // Positive voltage gamma control (14 bytes)
	NVGAMCTRL []uint8 // Negative voltage gamma control (14 bytes)
}

// New creates a new ST7789 connection. The SPI wire must already be configured.
func New(bus drivers.SPI, resetPin, dcPin, csPin, blPin machine.Pin) Device {
	dcPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	resetPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	blPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return Device{
		bus:      bus,
		dcPin:    dcPin,
		resetPin: resetPin,
		csPin:    csPin,
		blPin:    blPin,
	}
}

// Configure initializes the display with default configuration
func (d *Device) Configure(cfg Config) {
	if cfg.Width != 0 {
		d.width = cfg.Width
	} else {
		d.width = 240
	}
	if cfg.Height != 0 {
		d.height = cfg.Height
	} else {
		d.height = 240
	}

	d.rotation = cfg.Rotation
	d.rowOffsetCfg = cfg.RowOffset
	d.columnOffsetCfg = cfg.ColumnOffset

	if cfg.FrameRate != 0 {
		d.frameRate = cfg.FrameRate
	} else {
		d.frameRate = FRAMERATE_60
	}

	if cfg.VSyncLines >= 2 && cfg.VSyncLines <= 254 {
		d.vSyncLines = cfg.VSyncLines
	} else {
		d.vSyncLines = 16
	}

	d.batchLength = int32(d.width)
	if d.height > d.width {
		d.batchLength = int32(d.height)
	}
	d.batchLength += d.batchLength & 1

	// Reset the device
	d.resetPin.High()
	time.Sleep(50 * time.Millisecond)
	d.resetPin.Low()
	time.Sleep(50 * time.Millisecond)
	d.resetPin.High()
	time.Sleep(50 * time.Millisecond)

	// Common initialization
	d.startWrite()
	d.sendCommand(SWRESET, nil) // Soft reset
	d.endWrite()
	time.Sleep(150 * time.Millisecond) //
	d.startWrite()

	d.sendCommand(SLPOUT, nil) // Exit sleep mode

	// Memory initialization
	d.setColorFormat(ColorRGB565) // Set color mode to 16-bit color
	time.Sleep(10 * time.Millisecond)

	d.setRotation(d.rotation) // Memory orientation

	d.setWindow(0, 0, d.width, d.height)   // Full draw window
	d.fillScreen(color.RGBA{0, 0, 0, 255}) // Clear screen

	// Framerate
	d.sendCommand(FRCTRL2, []byte{byte(d.frameRate)}) // Frame rate for normal mode (default 60Hz)

	// Frame vertical sync and "porch"
	//
	// Front and back porch controls vertical scanline sync time before and after
	// a frame, where memory can be safely written without tearing.
	//
	fp := uint8(d.vSyncLines / 2)         // Split the desired pause half and half
	bp := uint8(d.vSyncLines - int16(fp)) // between front and back porch.

	d.sendCommand(PORCTRL, []byte{
		bp,   // Back porch 5bit     (0x7F max 0x08 default)
		fp,   // Front porch 5bit    (0x7F max 0x08 default)
		0x00, // Seprarate porch     (TODO: what is this?)
		0x22, // Idle mode porch     (4bit-back 4bit-front 0x22 default)
		0x22, // Partial mode porch  (4bit-back 4bit-front 0x22 default)
	})

	// Ready to display
	d.sendCommand(INVON, nil)         // Inversion ON
	time.Sleep(10 * time.Millisecond) //

	// Set gamma tables, if configured.
	if len(cfg.PVGAMCTRL) == 14 {
		d.sendCommand(GMCTRP1, cfg.PVGAMCTRL) // PVGAMCTRL: Positive Voltage Gamma Control
	}
	if len(cfg.NVGAMCTRL) == 14 {
		d.sendCommand(GMCTRN1, cfg.NVGAMCTRL) // NVGAMCTRL: Negative Voltage Gamma Control
	}

	d.sendCommand(NORON, nil)         // Normal mode ON
	time.Sleep(10 * time.Millisecond) //

	d.sendCommand(DISPON, nil)        // Screen ON
	time.Sleep(10 * time.Millisecond) //

	d.endWrite()
	d.blPin.High() // Backlight ON
}

// Send a command with data to the display. It does not change the chip select
// pin (it must be low when calling). The DC pin is left high after return,
// meaning that data can be sent right away.
func (d *Device) sendCommand(command uint8, data []byte) error {
	d.cmdBuf[0] = command
	d.dcPin.Low()
	err := d.bus.Tx(d.cmdBuf[:1], nil)
	d.dcPin.High()
	if len(data) != 0 {
		err = d.bus.Tx(data, nil)
	}
	return err
}

// startWrite must be called at the beginning of all exported methods to set the
// chip select pin low.
func (d *Device) startWrite() {
	if d.csPin != machine.NoPin {
		d.csPin.Low()
	}
}

// endWrite must be called at the end of all exported methods to set the chip
// select pin high.
func (d *Device) endWrite() {
	if d.csPin != machine.NoPin {
		d.csPin.High()
	}
}

// Sync waits for the display to hit the next VSYNC pause
func (d *Device) Sync() {
	d.SyncToScanLine(0)
}

// SyncToScanLine waits for the display to hit a specific scanline
//
// A scanline value of 0 will forward to the beginning of the next VSYNC,
// even if the display is currently in a VSYNC pause.
//
// Syncline values appear to increment once for every two vertical
// lines on the display.
//
// NOTE: Use GetHighestScanLine and GetLowestScanLine to obtain the highest
// and lowest useful values. Values are affected by front and back porch
// vsync settings (derived from VSyncLines configuration option).
func (d *Device) SyncToScanLine(scanline uint16) {
	scan := d.GetScanLine()

	// Sometimes GetScanLine returns erroneous 0 on first call after draw, so double check
	if scan == 0 {
		scan = d.GetScanLine()
	}

	if scanline == 0 {
		// we dont know where we are in an ongoing vsync so go around
		for scan < 1 {
			time.Sleep(1 * time.Millisecond)
			scan = d.GetScanLine()
		}
		for scan > 0 {
			scan = d.GetScanLine()
		}
	} else {
		// go around unless we're very close to the target
		for scan > scanline+4 {
			time.Sleep(1 * time.Millisecond)
			scan = d.GetScanLine()
		}
		for scan < scanline {
			scan = d.GetScanLine()
		}
	}
}

// GetScanLine reads the current scanline value from the display
func (d *Device) GetScanLine() uint16 {
	d.startWrite()
	data := []uint8{0x00, 0x00}
	d.dcPin.Low()
	d.bus.Transfer(GSCAN)
	d.dcPin.High()
	for i := range data {
		data[i], _ = d.bus.Transfer(0xFF)
	}
	scanline := uint16(data[0])<<8 + uint16(data[1])
	d.endWrite()
	return scanline
}

// GetHighestScanLine calculates the last scanline id in the frame before VSYNC pause
func (d *Device) GetHighestScanLine() uint16 {
	// Last scanline id appears to be backporch/2 + 320/2
	return uint16(math.Ceil(float64(d.vSyncLines)/2)/2) + 160
}

// GetLowestScanLine calculate the first scanline id to appear after VSYNC pause
func (d *Device) GetLowestScanLine() uint16 {
	// First scanline id appears to be backporch/2 + 1
	return uint16(math.Ceil(float64(d.vSyncLines)/2)/2) + 1
}

// Display does nothing, there's no buffer as it might be too big for some boards
func (d *Device) Display() error {
	return nil
}

// SetPixel sets a pixel in the screen
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	if x < 0 || y < 0 ||
		(((d.rotation == drivers.Rotation0 || d.rotation == drivers.Rotation180) && (x >= d.width || y >= d.height)) ||
			((d.rotation == drivers.Rotation90 || d.rotation == drivers.Rotation270) && (x >= d.height || y >= d.width))) {
		return
	}
	d.FillRectangle(x, y, 1, 1, c)
}

// setWindow prepares the screen to be modified at a given rectangle
func (d *Device) setWindow(x, y, w, h int16) {
	x += d.columnOffset
	y += d.rowOffset
	copy(d.buf[:4], []uint8{uint8(x >> 8), uint8(x), uint8((x + w - 1) >> 8), uint8(x + w - 1)})
	d.sendCommand(CASET, d.buf[:4])
	copy(d.buf[:4], []uint8{uint8(y >> 8), uint8(y), uint8((y + h - 1) >> 8), uint8(y + h - 1)})
	d.sendCommand(RASET, d.buf[:4])
	d.sendCommand(RAMWR, nil)
}

// FillRectangle fills a rectangle at a given coordinates with a color
func (d *Device) FillRectangle(x, y, width, height int16, c color.RGBA) error {
	d.startWrite()
	err := d.fillRectangle(x, y, width, height, c)
	d.endWrite()
	return err
}

func (d *Device) fillRectangle(x, y, width, height int16, c color.RGBA) error {
	k, i := d.Size()
	if x < 0 || y < 0 || width <= 0 || height <= 0 ||
		x >= k || (x+width) > k || y >= i || (y+height) > i {
		return errors.New("rectangle coordinates outside display area")
	}
	d.setWindow(x, y, width, height)
	c565 := RGBATo565(c)
	c1 := uint8(c565 >> 8)
	c2 := uint8(c565)

	data := make([]uint8, d.batchLength*2)
	for i := int32(0); i < d.batchLength; i++ {
		data[i*2] = c1
		data[i*2+1] = c2
	}
	j := int32(width) * int32(height)
	for j > 0 {
		// The DC pin is already set to data in the setWindow call, so we can
		// just write bytes on the SPI bus.
		if j >= d.batchLength {
			d.bus.Tx(data, nil)
		} else {
			d.bus.Tx(data[:j*2], nil)
		}
		j -= d.batchLength
	}
	return nil
}

// DrawRGBBitmap8 copies an RGB bitmap to the internal buffer at given coordinates
func (d *Device) DrawRGBBitmap8(x, y int16, data []uint8, w, h int16) error {
	k, i := d.Size()
	if x < 0 || y < 0 || w <= 0 || h <= 0 ||
		x >= k || (x+w) > k || y >= i || (y+h) > i {
		return errOutOfBounds
	}
	d.startWrite()
	d.setWindow(x, y, w, h)
	d.bus.Tx(data, nil)
	d.endWrite()
	return nil
}

// FillRectangleWithBuffer fills buffer with a rectangle at a given coordinates.
func (d *Device) FillRectangleWithBuffer(x, y, width, height int16, buffer []color.RGBA) error {
	i, j := d.Size()
	if x < 0 || y < 0 || width <= 0 || height <= 0 ||
		x >= i || (x+width) > i || y >= j || (y+height) > j {
		return errors.New("rectangle coordinates outside display area")
	}
	if int32(width)*int32(height) != int32(len(buffer)) {
		return errors.New("buffer length does not match with rectangle size")
	}
	d.startWrite()
	d.setWindow(x, y, width, height)

	k := int32(width) * int32(height)
	data := make([]uint8, d.batchLength*2)
	offset := int32(0)
	for k > 0 {
		for i := int32(0); i < d.batchLength; i++ {
			if offset+i < int32(len(buffer)) {
				c565 := RGBATo565(buffer[offset+i])
				c1 := uint8(c565 >> 8)
				c2 := uint8(c565)
				data[i*2] = c1
				data[i*2+1] = c2
			}
		}
		// The DC pin is already set to data in the setWindow call, so we don't
		// have to set it here.
		if k >= d.batchLength {
			d.bus.Tx(data, nil)
		} else {
			d.bus.Tx(data[:k*2], nil)
		}
		k -= d.batchLength
		offset += d.batchLength
	}
	d.endWrite()
	return nil
}

// DrawFastVLine draws a vertical line faster than using SetPixel
func (d *Device) DrawFastVLine(x, y0, y1 int16, c color.RGBA) {
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	d.FillRectangle(x, y0, 1, y1-y0+1, c)
}

// DrawFastHLine draws a horizontal line faster than using SetPixel
func (d *Device) DrawFastHLine(x0, x1, y int16, c color.RGBA) {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	d.FillRectangle(x0, y, x1-x0+1, 1, c)
}

// FillScreen fills the screen with a given color
func (d *Device) FillScreen(c color.RGBA) {
	d.startWrite()
	d.fillScreen(c)
	d.endWrite()
}

func (d *Device) fillScreen(c color.RGBA) {
	if d.rotation == NO_ROTATION || d.rotation == ROTATION_180 {
		d.fillRectangle(0, 0, d.width, d.height, c)
	} else {
		d.fillRectangle(0, 0, d.height, d.width, c)
	}
}

// Control the color format that is used when writing to the screen.
// The default is RGB565, setting it to any other value will break functions
// like SetPixel, FillRectangle, etc. Instead, you can write color data in the
// specified color format using DrawRGBBitmap8.
func (d *Device) SetColorFormat(format ColorFormat) {
	d.startWrite()
	d.setColorFormat(format)
	d.endWrite()
}

func (d *Device) setColorFormat(format ColorFormat) {
	// Lower 4 bits set the color format used in SPI.
	// Upper 4 bits set the color format used in the direct RGB interface.
	// The RGB interface is not currently supported, so it is left at a
	// reasonable default. Also, the RGB interface doesn't support RGB444.
	colmod := byte(format) | 0x50
	d.sendCommand(COLMOD, []byte{colmod})
}

// Rotation returns the current rotation of the device.
func (d *Device) Rotation() drivers.Rotation {
	return d.rotation
}

// SetRotation changes the rotation of the device (clock-wise)
func (d *Device) SetRotation(rotation Rotation) error {
	d.rotation = rotation
	d.startWrite()
	err := d.setRotation(rotation)
	d.endWrite()
	return err
}

func (d *Device) setRotation(rotation Rotation) error {
	madctl := uint8(0)
	switch rotation % 4 {
	case drivers.Rotation0:
		madctl = MADCTL_MX | MADCTL_MY
		d.rowOffset = d.rowOffsetCfg
		d.columnOffset = d.columnOffsetCfg
	case drivers.Rotation90:
		madctl = MADCTL_MY | MADCTL_MV
		d.rowOffset = d.columnOffsetCfg
		d.columnOffset = d.rowOffsetCfg
	case drivers.Rotation180:
		d.rowOffset = 0
		d.columnOffset = 0
	case drivers.Rotation270:
		madctl = MADCTL_MX | MADCTL_MV
		d.rowOffset = 0
		d.columnOffset = 0
	}
	if d.isBGR {
		madctl |= MADCTL_BGR
	}
	return d.sendCommand(MADCTL, []byte{madctl})
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	if d.rotation == drivers.Rotation0 || d.rotation == drivers.Rotation180 {
		return d.width, d.height
	}
	return d.height, d.width
}

// EnableBacklight enables or disables the backlight
func (d *Device) EnableBacklight(enable bool) {
	if enable {
		d.blPin.High()
	} else {
		d.blPin.Low()
	}
}

// Set the sleep mode for this LCD panel. When sleeping, the panel uses a lot
// less power. The LCD won't display an image anymore, but the memory contents
// will be kept.
func (d *Device) Sleep(sleepEnabled bool) error {
	if sleepEnabled {
		d.startWrite()
		d.sendCommand(SLPIN, nil)
		d.endWrite()
		time.Sleep(5 * time.Millisecond) // 5ms required by the datasheet
	} else {
		// Turn the LCD panel back on.
		d.startWrite()
		d.sendCommand(SLPOUT, nil)
		d.endWrite()
		// Note: the st7789 documentation says that it is needed to wait at
		// least 120ms before going to sleep again. Sleeping here would not be
		// practical (delays turning on the screen too much), so just hope the
		// screen won't need to sleep again for at least 120ms.
		// In practice, it's unlikely the user will set the display to sleep
		// again within 120ms.
	}
	return nil
}

// InvertColors inverts the colors of the screen
func (d *Device) InvertColors(invert bool) {
	d.startWrite()
	if invert {
		d.sendCommand(INVON, nil)
	} else {
		d.sendCommand(INVOFF, nil)
	}
	d.endWrite()
}

// IsBGR changes the color mode (RGB/BGR)
func (d *Device) IsBGR(bgr bool) {
	d.isBGR = bgr
}

// SetScrollArea sets an area to scroll with fixed top and bottom parts of the display.
func (d *Device) SetScrollArea(topFixedArea, bottomFixedArea int16) {
	copy(d.buf[:6], []uint8{
		uint8(topFixedArea >> 8), uint8(topFixedArea),
		uint8(d.height - topFixedArea - bottomFixedArea>>8), uint8(d.height - topFixedArea - bottomFixedArea),
		uint8(bottomFixedArea >> 8), uint8(bottomFixedArea)})
	d.startWrite()
	d.sendCommand(VSCRDEF, d.buf[:6])
	d.endWrite()
}

// SetScroll sets the vertical scroll address of the display.
func (d *Device) SetScroll(line int16) {
	d.buf[0] = uint8(line >> 8)
	d.buf[1] = uint8(line)
	d.startWrite()
	d.sendCommand(VSCRSADD, d.buf[:2])
	d.endWrite()
}

// StopScroll returns the display to its normal state.
func (d *Device) StopScroll() {
	d.startWrite()
	d.sendCommand(NORON, nil)
	d.endWrite()
}

// RGBATo565 converts a color.RGBA to uint16 used in the display
func RGBATo565(c color.RGBA) uint16 {
	r, g, b, _ := c.RGBA()
	return uint16((r & 0xF800) +
		((g & 0xFC00) >> 5) +
		((b & 0xF800) >> 11))
}
