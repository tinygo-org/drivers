package ili9341

import (
	"errors"
	"fmt"
	"image/color"
	"time"

	"github.com/tinygo-org/tinygo/src/machine"
)

const _debug = false

type Config struct {
	Width    int16
	Height   int16
	Rotation Rotation
}

type Device struct {
	width    int16
	height   int16
	rotation Rotation
	driver   driver

	dc  machine.Pin
	cs  machine.Pin
	rst machine.Pin
	rd  machine.Pin
}

func (d *Device) Configure(config Config) {

	if config.Width == 0 {
		config.Width = TFTWIDTH
	}
	if config.Height == 0 {
		config.Height = TFTHEIGHT
	}
	d.width = config.Width
	d.height = config.Height

	output := machine.PinConfig{machine.PinOutput}
	if _debug {
		println("height, width == ", d.width, d.height)
	}

	// configure chip select if there is one
	if _debug {
		println("configuring cs")
	}
	if d.cs != machine.NoPin {
		d.cs.Configure(output)
		d.cs.High() // deselect
	}

	if _debug {
		println("configuring dc")
	}
	d.dc.Configure(output)
	d.dcHigh() // data mode

	// driver-specific configuration
	if _debug {
		println("configuring driver")
	}
	d.driver.configure(&config)

	if _debug {
		println("configuring rd")
	}
	if d.rd != machine.NoPin {
		d.rd.Configure(output)
		d.rd.High()
	}

	// reset the display
	if _debug {
		println("configuring rst")
	}
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

	if _debug {
		println("configuring initCmd")
	}
	initCmd := []byte{
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
		SLPOUT, 0x80, // Exit Sleep
		DISPON, 0x80, // Display on
		0x00, // End of list
	}
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

}

// Size returns the current size of the display.
func (d *Device) Size() (x, y int16) {
	return d.width, d.height
}

// SetPixel modifies the internal buffer.
func (d *Device) SetPixel(x, y int16, c color.RGBA) {
	d.setWindow(x, y, 1, 1)
	c565 := RGBATo565(c)
	d.startWrite()
	d.driver.writeByte(byte(c565 >> 8))
	d.driver.writeByte(byte(c565))
	d.endWrite()
}

// Display sends the buffer (if any) to the screen.
func (d *Device) Display() error {
	return nil
}

// FillRectangle fills a rectangle at a given coordinates with a color
func (d *Device) FillRectangle(x, y, width, height int16, c color.RGBA) error {
	k, i := d.Size()
	if x < 0 || y < 0 || width <= 0 || height <= 0 ||
		x >= k || (x+width) > k || y >= i || (y+height) > i {
		return errors.New("rectangle coordinates outside display area")
	}
	d.setWindow(x, y, width, height)
	c565 := RGBATo565(c)
	d.startWrite()
	d.writeColor(c565, int(width)*int(height))
	d.endWrite()
	return nil
}

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

// setWindow prepares the screen to be modified at a given rectangle
func (d *Device) setWindow(x, y, w, h int16) {
	//x += d.columnOffset
	//y += d.rowOffset
	d.sendCommand(CASET, []uint8{
		uint8(x << 8), uint8(x), uint8((x + w - 1) >> 8), uint8(x + w - 1),
	})
	d.sendCommand(PASET, []uint8{
		uint8(y >> 8), uint8(y), uint8((y + h - 1) >> 8), uint8(y + h - 1),
	})
	d.sendCommand(RAMWR, nil)
}

func (d *Device) writeColor(c565 uint16, l int) {
	hi := uint8(c565 >> 8)
	lo := uint8(c565)
	for i := 0; i < l; i++ {
		d.driver.writeByte(hi)
		d.driver.writeByte(lo)
	}
}

func (d *Device) startWrite() {
	d.csLow()
}

func (d *Device) endWrite() {
	d.csHigh()
}

func (d *Device) sendCommand(cmd byte, data []byte) {
	if _debug {
		fmt.Printf("sending command: %02X", cmd)
		for _, b := range data {
			fmt.Printf(" %02X", b)
		}
		println()
	}
	d.csLow()
	d.dcLow()
	d.driver.writeByte(cmd)
	d.dcHigh()
	for _, b := range data {
		d.driver.writeByte(b)
	}
	d.csHigh()
}

//go:inline
func (d *Device) csHigh() {
	if d.cs != machine.NoPin {
		d.cs.High()
	}
}

//go:inline
func (d *Device) csLow() {
	if d.cs != machine.NoPin {
		d.cs.Low()
	}
}

//go:inline
func (d *Device) dcHigh() {
	d.dc.High()
}

//go:inline
func (d *Device) dcLow() {
	d.dc.Low()
}

type driver interface {
	configure(config *Config)
	writeByte(b byte)
}

/*
type spiDriver struct {
	spi *machine.SPI
}

func NewSPI(spi *machine.SPI, dc, cs, rst, rd machine.Pin) *Device {
	return &Device{
		dc:  dc,
		cs:  cs,
		rd:  rd,
		rst: rst,
		driver: &spiDriver{
			spi: spi,
		},
	}
}

func (sd *spiDriver) configure(config *Config) {
}

func (sd *spiDriver) writeByte(b byte) {
	sd.spi.Transfer(b)
}
*/

func delay(micros int) {
	/*
		time.Sleep(time.Duration(micros) * time.Millisecond)
	*/
	t := time.Now().UnixNano() + int64(time.Duration(micros*1000)*time.Microsecond)
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
