package ili9341

import (
	"errors"
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
)

type Config struct {
	Width            int16
	Height           int16
	Rotation         drivers.Rotation
	DisplayInversion bool
}

type Device struct {
	width    int16
	height   int16
	rotation drivers.Rotation
	driver   driver

	x0, x1 int16 // cached address window; prevents useless/expensive
	y0, y1 int16 // syscalls to PASET and CASET

	dc  machine.Pin
	cs  machine.Pin
	rst machine.Pin
	rd  machine.Pin
}

var cmdBuf [6]byte

var initCmd = []byte{
	0xEF, 3, 0x03, 0x80, 0x02,
	0xCF, 3, 0x00, 0xC1, 0x30,
	0xED, 4, 0x64, 0x03, 0x12, 0x81,
	0xE8, 3, 0x85, 0x00, 0x78,
	0xCB, 5, 0x39, 0x2C, 0x00, 0x34, 0x02,
	0xF7, 1, 0x20,
	0xEA, 2, 0x00, 0x00,
	PWCTR1, 1, 0x23, // Power control VRH[5:0]
	PWCTR2, 1, 0x10, // Power control SAP[2:0];BT[3:0]
	VMCTR1, 2, 0x3e, 0x28, // VCM control
	VMCTR2, 1, 0x86, // VCM control2
	MADCTL, 1, 0x48, // Memory Access Control
	VSCRSADD, 1, 0x00, // Vertical scroll zero
	PIXFMT, 1, 0x55,
	FRMCTR1, 2, 0x00, 0x18,
	DFUNCTR, 3, 0x08, 0x82, 0x27, // Display Function Control
	0xF2, 1, 0x00, // 3Gamma Function Disable
	GAMMASET, 1, 0x01, // Gamma curve selected
	GMCTRP1, 15, 0x0F, 0x31, 0x2B, 0x0C, 0x0E, 0x08, // Set Gamma
	0x4E, 0xF1, 0x37, 0x07, 0x10, 0x03, 0x0E, 0x09, 0x00,
	GMCTRN1, 15, 0x00, 0x0E, 0x14, 0x03, 0x11, 0x07, // Set Gamma
	0x31, 0xC1, 0x48, 0x08, 0x0F, 0x0C, 0x31, 0x36, 0x0F,
}

// Configure prepares display for use
func (d *Device) Configure(config Config) {

	if config.Width == 0 {
		config.Width = TFTWIDTH
	}
	if config.Height == 0 {
		config.Height = TFTHEIGHT
	}
	d.width = config.Width
	d.height = config.Height
	d.rotation = config.Rotation

	// try to pick an initial cache miss for one of the points
	d.x0, d.x1 = -(d.width + 1), d.x0
	d.y0, d.y1 = -(d.height + 1), d.y0

	output := machine.PinConfig{machine.PinOutput}

	// configure chip select if there is one
	if d.cs != machine.NoPin {
		d.cs.Configure(output)
		d.cs.High() // deselect
	}

	d.dc.Configure(output)
	d.dc.High() // data mode

	// driver-specific configuration
	d.driver.configure(&config)

	if d.rd != machine.NoPin {
		d.rd.Configure(output)
		d.rd.High()
	}

	// reset the display
	if d.rst != machine.NoPin {
		// configure hardware reset if there is one
		d.rst.Configure(output)
		d.rst.High()
		delay(100)
		d.rst.Low()
		delay(100)
		d.rst.High()
		delay(200)
	} else {
		// if no hardware reset, send software reset
		d.sendCommand(SWRESET, nil)
		delay(150)
	}

	if config.DisplayInversion {
		initCmd = append(initCmd, INVON, 0x80)
	}

	initCmd = append(initCmd,
		SLPOUT, 0x80, // Exit Sleep
		DISPON, 0x80, // Display on
		0x00, // End of list
	)
	for i, c := 0, len(initCmd); i < c; {
		cmd := initCmd[i]
		if cmd == 0x00 {
			break
		}
		x := initCmd[i+1]
		numArgs := int(x & 0x7F)
		d.sendCommand(cmd, initCmd[i+2:i+2+numArgs])
		if x&0x80 > 0 {
			delay(150)
		}
		i += numArgs + 2
	}

	d.SetRotation(d.rotation)
}

// Size returns the current size of the display.
func (d *Device) Size() (x, y int16) {
	switch d.rotation {
	case Rotation90, Rotation270, Rotation90Mirror, Rotation270Mirror:
		return d.height, d.width
	default: // Rotation0, Rotation180, etc
		return d.width, d.height
	}
}

// SetPixel modifies the internal buffer.
func (d *Device) SetPixel(x, y int16, c color.RGBA) {
	d.setWindow(x, y, 1, 1)
	c565 := RGBATo565(c)
	d.startWrite()
	d.driver.write16(c565)
	d.endWrite()
}

// Display sends the buffer (if any) to the screen.
func (d *Device) Display() error {
	return nil
}

// EnableTEOutput enables the TE ("tearing effect") line.
// The TE line goes high when the screen is not currently being updated and can
// be used to start drawing. When used correctly, it can avoid tearing entirely.
func (d *Device) EnableTEOutput(on bool) {
	if on {
		cmdBuf[0] = 0
		d.sendCommand(TEON, cmdBuf[:1]) // M=0 (V-blanking only, no H-blanking)
	} else {
		d.sendCommand(TEOFF, nil) // TEOFF
	}
}

// DrawRGBBitmap copies an RGB bitmap to the internal buffer at given coordinates
func (d *Device) DrawRGBBitmap(x, y int16, data []uint16, w, h int16) error {
	k, i := d.Size()
	if x < 0 || y < 0 || w <= 0 || h <= 0 ||
		x >= k || (x+w) > k || y >= i || (y+h) > i {
		return errors.New("rectangle coordinates outside display area")
	}
	d.setWindow(x, y, w, h)
	d.startWrite()
	d.driver.write16sl(data)
	d.endWrite()
	return nil
}

// DrawRGBBitmap8 copies an RGB bitmap to the internal buffer at given coordinates
func (d *Device) DrawRGBBitmap8(x, y int16, data []uint8, w, h int16) error {
	k, i := d.Size()
	if x < 0 || y < 0 || w <= 0 || h <= 0 ||
		x >= k || (x+w) > k || y >= i || (y+h) > i {
		return errors.New("rectangle coordinates outside display area")
	}
	d.setWindow(x, y, w, h)
	d.startWrite()
	d.driver.write8sl(data)
	d.endWrite()
	return nil
}

// FillRectangle fills a rectangle at given coordinates with a color
func (d *Device) FillRectangle(x, y, width, height int16, c color.RGBA) error {
	k, i := d.Size()
	if x < 0 || y < 0 || width <= 0 || height <= 0 ||
		x >= k || (x+width) > k || y >= i || (y+height) > i {
		return errors.New("rectangle coordinates outside display area")
	}
	d.setWindow(x, y, width, height)
	c565 := RGBATo565(c)
	d.startWrite()
	d.driver.write16n(c565, int(width)*int(height))
	d.endWrite()
	return nil
}

// DrawRectangle draws a rectangle at given coordinates with a color
func (d *Device) DrawRectangle(x, y, w, h int16, c color.RGBA) error {
	if err := d.DrawFastHLine(x, x+w-1, y, c); err != nil {
		return err
	}
	if err := d.DrawFastHLine(x, x+w-1, y+h-1, c); err != nil {
		return err
	}
	if err := d.DrawFastVLine(x, y, y+h-1, c); err != nil {
		return err
	}
	if err := d.DrawFastVLine(x+w-1, y, y+h-1, c); err != nil {
		return err
	}
	return nil
}

// DrawFastVLine draws a vertical line faster than using SetPixel
func (d *Device) DrawFastVLine(x, y0, y1 int16, c color.RGBA) error {
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	return d.FillRectangle(x, y0, 1, y1-y0+1, c)
}

// DrawFastHLine draws a horizontal line faster than using SetPixel
func (d *Device) DrawFastHLine(x0, x1, y int16, c color.RGBA) error {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	return d.FillRectangle(x0, y, x1-x0+1, 1, c)
}

// FillScreen fills the screen with a given color
func (d *Device) FillScreen(c color.RGBA) {
	if d.rotation == Rotation0 || d.rotation == Rotation180 {
		d.FillRectangle(0, 0, d.width, d.height, c)
	} else {
		d.FillRectangle(0, 0, d.height, d.width, c)
	}
}

// Set the sleep mode for this LCD panel. When sleeping, the panel uses a lot
// less power. The LCD won't display an image anymore, but the memory contents
// will be kept.
func (d *Device) Sleep(sleepEnabled bool) error {
	if sleepEnabled {
		// Shut down LCD panel.
		d.sendCommand(SLPIN, nil)
		time.Sleep(5 * time.Millisecond) // 5ms required by the datasheet
	} else {
		// Turn the LCD panel back on.
		d.sendCommand(SLPOUT, nil)
		// Note: the ili9341 documentation says that it is needed to wait at
		// least 120ms before going to sleep again. Sleeping here would not be
		// practical (delays turning on the screen too much), so just hope the
		// screen won't need to sleep again for at least 120ms.
		// In practice, it's unlikely the user will set the display to sleep
		// again within 120ms.
	}
	return nil
}

// Rotation returns the current rotation of the device.
func (d *Device) Rotation() drivers.Rotation {
	return d.rotation
}

// GetRotation returns the current rotation of the device.
//
// Deprecated: use Rotation instead.
func (d *Device) GetRotation() drivers.Rotation {
	return d.rotation
}

// SetRotation changes the rotation of the device (clock-wise).
func (d *Device) SetRotation(rotation drivers.Rotation) error {
	madctl := uint8(0)
	switch rotation % 8 {
	case Rotation0:
		madctl = MADCTL_MX | MADCTL_BGR
	case Rotation90:
		madctl = MADCTL_MV | MADCTL_BGR
	case Rotation180:
		madctl = MADCTL_MY | MADCTL_BGR | MADCTL_ML
	case Rotation270:
		madctl = MADCTL_MX | MADCTL_MY | MADCTL_MV | MADCTL_BGR | MADCTL_ML
	case Rotation0Mirror:
		madctl = MADCTL_BGR
	case Rotation90Mirror:
		madctl = MADCTL_MY | MADCTL_MV | MADCTL_BGR | MADCTL_ML
	case Rotation180Mirror:
		madctl = MADCTL_MX | MADCTL_MY | MADCTL_BGR | MADCTL_ML
	case Rotation270Mirror:
		madctl = MADCTL_MX | MADCTL_MY | MADCTL_MV | MADCTL_BGR | MADCTL_ML
	}
	cmdBuf[0] = madctl
	d.sendCommand(MADCTL, cmdBuf[:1])
	d.rotation = rotation
	return nil
}

// SetScrollArea sets an area to scroll with fixed top/bottom or left/right parts of the display
// Rotation affects scroll direction
func (d *Device) SetScrollArea(topFixedArea, bottomFixedArea int16) {
	cmdBuf[0] = uint8(topFixedArea >> 8)
	cmdBuf[1] = uint8(topFixedArea)
	cmdBuf[2] = uint8(d.height - topFixedArea - bottomFixedArea>>8)
	cmdBuf[3] = uint8(d.height - topFixedArea - bottomFixedArea)
	cmdBuf[4] = uint8(bottomFixedArea >> 8)
	cmdBuf[5] = uint8(bottomFixedArea)
	d.sendCommand(VSCRDEF, cmdBuf[:6])
}

// SetScroll sets the vertical scroll address of the display.
func (d *Device) SetScroll(line int16) {
	cmdBuf[0] = uint8(line >> 8)
	cmdBuf[1] = uint8(line)
	d.sendCommand(VSCRSADD, cmdBuf[:2])
}

// StopScroll returns the display to its normal state
func (d *Device) StopScroll() {
	d.sendCommand(NORON, nil)
}

// setWindow prepares the screen to be modified at a given rectangle
func (d *Device) setWindow(x, y, w, h int16) {
	//x += d.columnOffset
	//y += d.rowOffset
	x1 := x + w - 1
	if x != d.x0 || x1 != d.x1 {
		cmdBuf[0] = uint8(x >> 8)
		cmdBuf[1] = uint8(x)
		cmdBuf[2] = uint8(x1 >> 8)
		cmdBuf[3] = uint8(x1)
		d.sendCommand(CASET, cmdBuf[:4])
		d.x0, d.x1 = x, x1
	}
	y1 := y + h - 1
	if y != d.y0 || y1 != d.y1 {
		cmdBuf[0] = uint8(y >> 8)
		cmdBuf[1] = uint8(y)
		cmdBuf[2] = uint8(y1 >> 8)
		cmdBuf[3] = uint8(y1)
		d.sendCommand(PASET, cmdBuf[:4])
		d.y0, d.y1 = y, y1
	}
	d.sendCommand(RAMWR, nil)
}

//go:inline
func (d *Device) startWrite() {
	if d.cs != machine.NoPin {
		d.cs.Low()
	}
}

//go:inline
func (d *Device) endWrite() {
	if d.cs != machine.NoPin {
		d.cs.High()
	}
}

func (d *Device) sendCommand(cmd byte, data []byte) {
	d.startWrite()
	d.dc.Low()
	d.driver.write8(cmd)
	d.dc.High()
	if data != nil {
		d.driver.write8sl(data)
	}
	d.endWrite()
}

type driver interface {
	configure(config *Config)
	write8(b byte)
	write8n(b byte, n int)
	write8sl(b []byte)
	write16(data uint16)
	write16n(data uint16, n int)
	write16sl(data []uint16)
}

func delay(m int) {
	t := time.Now().UnixNano() + int64(time.Duration(m*1000)*time.Microsecond)
	for time.Now().UnixNano() < t {
	}
}

// RGBATo565 converts a color.RGBA to uint16 used in the display
func RGBATo565(c color.RGBA) uint16 {
	r, g, b, _ := c.RGBA()
	return uint16((r & 0xF800) +
		((g & 0xFC00) >> 5) +
		((b & 0xF800) >> 11))
}
