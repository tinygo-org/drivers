// Package epd2in7 implements a driver for the Waveshare 2.7in V1 black and
// white e-paper panel. This panel uses an IL91874-style controller, not the
// IL3820 (epd2in13) or SSD1680 (epd2in9v2) controllers used by other
// packages in this repository, so it needs its own init sequence and LUT.
//
// This controller only updates its BUSY pin after a GET_STATUS command, so
// WaitUntilIdle sends one before every read. Reading BUSY on its own leaves
// it stuck at the post-reset level and never returns.
//
// Datasheet: https://www.waveshare.com/wiki/2.7inch_e-Paper_HAT
package epd2in7 // import "tinygo.org/x/drivers/waveshare-epd/epd2in7"

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
)

type Config struct {
	Width        int16 // Width is the display resolution
	Height       int16
	LogicalWidth int16 // LogicalWidth must be a multiple of 8 and same size or bigger than Width
	Rotation     drivers.Rotation
}

type Device struct {
	bus          drivers.SPI
	cs           machine.Pin
	dc           machine.Pin
	rst          machine.Pin
	busy         machine.Pin
	logicalWidth int16
	width        int16
	height       int16
	buffer       []uint8
	bufferLength uint32
	rotation     drivers.Rotation
}

// New returns a new epd2in7 driver. Pass in a fully configured SPI bus.
func New(bus drivers.SPI, csPin, dcPin, rstPin, busyPin machine.Pin) Device {
	csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	dcPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	rstPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	busyPin.Configure(machine.PinConfig{Mode: machine.PinInput})
	return Device{
		bus:  bus,
		cs:   csPin,
		dc:   dcPin,
		rst:  rstPin,
		busy: busyPin,
	}
}

// Configure sets up the device.
func (d *Device) Configure(cfg Config) {
	if cfg.LogicalWidth != 0 {
		d.logicalWidth = cfg.LogicalWidth
	} else {
		d.logicalWidth = EPD_WIDTH
	}
	if cfg.Width != 0 {
		d.width = cfg.Width
	} else {
		d.width = EPD_WIDTH
	}
	if cfg.Height != 0 {
		d.height = cfg.Height
	} else {
		d.height = EPD_HEIGHT
	}
	d.rotation = cfg.Rotation
	d.bufferLength = (uint32(d.logicalWidth) * uint32(d.height)) / 8
	d.buffer = make([]uint8, d.bufferLength)
	for i := uint32(0); i < d.bufferLength; i++ {
		d.buffer[i] = 0xFF
	}

	d.cs.Low()
	d.dc.Low()
	d.rst.Low()

	d.Reset()

	d.SendCommand(POWER_SETTING)
	d.SendData(0x03) // VDS_EN, VDG_EN
	d.SendData(0x00) // VCOM_HV, VGHL_LV[1], VGHL_LV[0]
	d.SendData(0x2b) // VDH
	d.SendData(0x2b) // VDL
	d.SendData(0x09) // VDHR

	d.SendCommand(BOOSTER_SOFT_START)
	d.SendData(0x07)
	d.SendData(0x07)
	d.SendData(0x17)

	// Power optimization. Values come from the vendor driver and are not
	// documented in the datasheet register map.
	d.SendCommand(0xF8)
	d.SendData(0x60)
	d.SendData(0xA5)
	d.SendCommand(0xF8)
	d.SendData(0x89)
	d.SendData(0xA5)
	d.SendCommand(0xF8)
	d.SendData(0x90)
	d.SendData(0x00)
	d.SendCommand(0xF8)
	d.SendData(0x93)
	d.SendData(0x2A)
	d.SendCommand(0xF8)
	d.SendData(0xA0)
	d.SendData(0xA5)
	d.SendCommand(0xF8)
	d.SendData(0xA1)
	d.SendData(0x00)
	d.SendCommand(0xF8)
	d.SendData(0x73)
	d.SendData(0x41)

	d.SendCommand(PARTIAL_DISPLAY_REFRESH)
	d.SendData(0x00)

	d.SendCommand(POWER_ON)
	d.WaitUntilIdle()

	d.SendCommand(PANEL_SETTING)
	d.SendData(0xAF) // KW-BF   KWR-AF    BWROTP 0f

	d.SendCommand(PLL_CONTROL)
	d.SendData(0x3A) // 3A 100Hz   29 150Hz   39 200Hz   31 171Hz

	d.SendCommand(VCOM_AND_DATA_INTERVAL_SETTING)
	d.SendData(0x57)

	d.SendCommand(VCM_DC_SETTING)
	d.SendData(0x12)

	d.SetLUT()
}

// Reset resets the device.
func (d *Device) Reset() {
	d.rst.Low()
	time.Sleep(200 * time.Millisecond)
	d.rst.High()
	time.Sleep(200 * time.Millisecond)
}

// DeepSleep puts the display into deep sleep. The panel keeps the last
// image with no power, but a hardware reset is needed to wake it again.
func (d *Device) DeepSleep() {
	d.SendCommand(VCOM_AND_DATA_INTERVAL_SETTING)
	d.SendData(0x17) // border floating
	d.SendCommand(VCM_DC_SETTING)
	d.SendCommand(PANEL_SETTING)
	time.Sleep(100 * time.Millisecond)

	d.SendCommand(POWER_SETTING) // VG&VS to 0V fast
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x00)
	time.Sleep(100 * time.Millisecond)

	d.SendCommand(POWER_OFF)
	d.WaitUntilIdle()
	d.SendCommand(DEEP_SLEEP)
	d.SendData(0xA5)
}

// SendCommand sends a command to the display.
func (d *Device) SendCommand(command uint8) {
	d.sendDataCommand(true, command)
}

// SendData sends a data byte to the display.
func (d *Device) SendData(data uint8) {
	d.sendDataCommand(false, data)
}

// sendDataCommand sends image data or a command to the screen.
func (d *Device) sendDataCommand(isCommand bool, data uint8) {
	if isCommand {
		d.dc.Low()
	} else {
		d.dc.High()
	}
	d.cs.Low()
	d.bus.Transfer(data)
	d.cs.High()
}

// SetLUT sets the look up tables for a full update. Values come from the
// vendor EPD_2in7.c driver, section EPD_2in7_lut_vcom_dc and following.
func (d *Device) SetLUT() {
	lutVcom := []uint8{
		0x00, 0x00,
		0x00, 0x08, 0x00, 0x00, 0x00, 0x02,
		0x60, 0x28, 0x28, 0x00, 0x00, 0x01,
		0x00, 0x14, 0x00, 0x00, 0x00, 0x01,
		0x00, 0x12, 0x12, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	lutWW := []uint8{
		0x40, 0x08, 0x00, 0x00, 0x00, 0x02,
		0x90, 0x28, 0x28, 0x00, 0x00, 0x01,
		0x40, 0x14, 0x00, 0x00, 0x00, 0x01,
		0xA0, 0x12, 0x12, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	lutBW := []uint8{
		0x40, 0x08, 0x00, 0x00, 0x00, 0x02,
		0x90, 0x28, 0x28, 0x00, 0x00, 0x01,
		0x40, 0x14, 0x00, 0x00, 0x00, 0x01,
		0xA0, 0x12, 0x12, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	lutBB := []uint8{
		0x80, 0x08, 0x00, 0x00, 0x00, 0x02,
		0x90, 0x28, 0x28, 0x00, 0x00, 0x01,
		0x80, 0x14, 0x00, 0x00, 0x00, 0x01,
		0x50, 0x12, 0x12, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	lutWB := []uint8{
		0x80, 0x08, 0x00, 0x00, 0x00, 0x02,
		0x90, 0x28, 0x28, 0x00, 0x00, 0x01,
		0x80, 0x14, 0x00, 0x00, 0x00, 0x01,
		0x50, 0x12, 0x12, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	d.SendCommand(LUT_FOR_VCOM)
	for _, v := range lutVcom {
		d.SendData(v)
	}
	d.SendCommand(LUT_WHITE_TO_WHITE)
	for _, v := range lutWW {
		d.SendData(v)
	}
	d.SendCommand(LUT_BLACK_TO_WHITE)
	for _, v := range lutBW {
		d.SendData(v)
	}
	d.SendCommand(LUT_WHITE_TO_BLACK)
	for _, v := range lutBB {
		d.SendData(v)
	}
	d.SendCommand(LUT_BLACK_TO_BLACK)
	for _, v := range lutWB {
		d.SendData(v)
	}
}

// SetPixel modifies the internal buffer at a single pixel.
// The display has 2 colors: black and white. We use RGBA(0,0,0,255) as
// black, anything else as white.
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	x, y = d.xy(x, y)
	if x < 0 || x >= d.logicalWidth || y < 0 || y >= d.height {
		return
	}
	byteIndex := (uint32(x) + uint32(y)*uint32(d.logicalWidth)) / 8
	if c.R == 0 && c.G == 0 && c.B == 0 { // black
		d.buffer[byteIndex] &^= 0x80 >> uint8(x%8)
	} else { // white
		d.buffer[byteIndex] |= 0x80 >> uint8(x%8)
	}
}

// Display sends the buffer to the screen and refreshes it.
func (d *Device) Display() error {
	d.SendCommand(DATA_START_TRANSMISSION_2)
	for i := uint32(0); i < d.bufferLength; i++ {
		d.SendData(d.buffer[i])
	}

	d.SendCommand(DISPLAY_REFRESH)
	time.Sleep(100 * time.Millisecond)
	d.WaitUntilIdle()
	return nil
}

// ClearDisplay erases the device SRAM and refreshes the panel to white.
func (d *Device) ClearDisplay() {
	d.SendCommand(DATA_START_TRANSMISSION_1)
	for i := uint32(0); i < d.bufferLength; i++ {
		d.SendData(0xFF)
	}
	d.SendCommand(DATA_START_TRANSMISSION_2)
	for i := uint32(0); i < d.bufferLength; i++ {
		d.SendData(0xFF)
	}

	d.SendCommand(DISPLAY_REFRESH)
	time.Sleep(100 * time.Millisecond)
	d.WaitUntilIdle()
}

// WaitUntilIdle waits until the display is ready. This controller only
// updates the BUSY pin after a GET_STATUS command, unlike other panels in
// this repository, so send it before every read.
// Source: EPD_2in7_ReadBusy in the vendor EPD_2in7.c driver.
func (d *Device) WaitUntilIdle() {
	for {
		d.SendCommand(GET_STATUS)
		if d.busy.Get() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// IsBusy returns the busy status of the display.
func (d *Device) IsBusy() bool {
	return d.busy.Get()
}

// ClearBuffer sets the buffer to 0xFF (white).
func (d *Device) ClearBuffer() {
	for i := uint32(0); i < d.bufferLength; i++ {
		d.buffer[i] = 0xFF
	}
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	if d.rotation == drivers.Rotation90 || d.rotation == drivers.Rotation270 {
		return d.height, d.logicalWidth
	}
	return d.logicalWidth, d.height
}

// Rotation returns the current rotation of the device.
func (d *Device) Rotation() drivers.Rotation {
	return d.rotation
}

// SetRotation changes the rotation of the device.
func (d *Device) SetRotation(rotation drivers.Rotation) error {
	d.rotation = rotation
	return nil
}

// xy changes the coordinates according to the rotation.
func (d *Device) xy(x, y int16) (int16, int16) {
	switch d.rotation {
	case drivers.Rotation0:
		return x, y
	case drivers.Rotation90:
		return d.width - y - 1, x
	case drivers.Rotation180:
		return d.width - x - 1, d.height - y - 1
	case drivers.Rotation270:
		return y, d.height - x - 1
	}
	return x, y
}
