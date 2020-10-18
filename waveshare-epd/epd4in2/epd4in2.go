// Package epd4in2 implements a driver for Waveshare 4.2in black and white e-paper device.
//
// Derived from:
//   https://github.com/tinygo-org/drivers/tree/master/waveshare-epd
//   https://github.com/waveshare/e-Paper/blob/master/Arduino/epd4in2/epd4in2.cpp
//
// Datasheet: https://www.waveshare.com/wiki/4.2inch_e-Paper_Module
//
package epd4in2

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
)

type Config struct {
	Width        int16 // Width is the display resolution
	Height       int16
	LogicalWidth int16    // LogicalWidth must be a multiple of 8 and same size or bigger than Width
	Rotation     Rotation // Rotation is clock-wise
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
	rotation     Rotation
}

type Rotation uint8

// New returns a new epd4in2 driver. Pass in a fully configured SPI bus.
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
	d.SendData(0xff) // VDHR
	d.SendCommand(BOOSTER_SOFT_START)
	d.SendData(0x17)
	d.SendData(0x17)
	d.SendData(0x17) //07 0f 17 1f 27 2F 37 2f
	d.SendCommand(POWER_ON)
	d.WaitUntilIdle()
	d.SendCommand(PANEL_SETTING)
	d.SendData(0xbf) // KW-BF   KWR-AF  BWROTP 0f
	d.SendData(0x0b)
	d.SendCommand(PLL_CONTROL)
	d.SendData(0x3c) // 3A 100HZ   29 150Hz 39 200HZ  31 171HZ
}

// Reset resets the device
func (d *Device) Reset() {
	d.rst.Low()
	time.Sleep(200 * time.Millisecond)
	d.rst.High()
	time.Sleep(200 * time.Millisecond)
}

// DeepSleep puts the display into deepsleep
func (d *Device) DeepSleep() {
	d.SendCommand(VCOM_AND_DATA_INTERVAL_SETTING)
	d.SendData(0x17)              //border floating
	d.SendCommand(VCM_DC_SETTING) //VCOM to 0V
	d.SendCommand(PANEL_SETTING)
	time.Sleep(100 * time.Millisecond)

	d.SendCommand(POWER_SETTING) //VG&VS to 0V fast
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x00)
	time.Sleep(100 * time.Millisecond)

	d.SendCommand(POWER_OFF) //power off
	d.WaitUntilIdle()
	d.SendCommand(DEEP_SLEEP) //deep sleep
	d.SendData(0xA5)
}

// SendCommand sends a command to the display
func (d *Device) SendCommand(command uint8) {
	d.sendDataCommand(true, command)
}

// SendData sends a data byte to the display
func (d *Device) SendData(data uint8) {
	d.sendDataCommand(false, data)
}

// sendDataCommand sends image data or a command to the screen
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

// SetLUT sets the look up tables for full or partial updates
func (d *Device) SetLUT() {
	lut_vcom0 := []uint8{
		0x00, 0x17, 0x00, 0x00, 0x00, 0x02,
		0x00, 0x17, 0x17, 0x00, 0x00, 0x02,
		0x00, 0x0A, 0x01, 0x00, 0x00, 0x01,
		0x00, 0x0E, 0x0E, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, // 44 bytes, unlike the others
	}
	lut_ww := []uint8{
		0x40, 0x17, 0x00, 0x00, 0x00, 0x02,
		0x90, 0x17, 0x17, 0x00, 0x00, 0x02,
		0x40, 0x0A, 0x01, 0x00, 0x00, 0x01,
		0xA0, 0x0E, 0x0E, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	lut_bw := []uint8{
		0x40, 0x17, 0x00, 0x00, 0x00, 0x02,
		0x90, 0x17, 0x17, 0x00, 0x00, 0x02,
		0x40, 0x0A, 0x01, 0x00, 0x00, 0x01,
		0xA0, 0x0E, 0x0E, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	lut_bb := []uint8{
		0x80, 0x17, 0x00, 0x00, 0x00, 0x02,
		0x90, 0x17, 0x17, 0x00, 0x00, 0x02,
		0x80, 0x0A, 0x01, 0x00, 0x00, 0x01,
		0x50, 0x0E, 0x0E, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	lut_wb := []uint8{
		0x80, 0x17, 0x00, 0x00, 0x00, 0x02,
		0x90, 0x17, 0x17, 0x00, 0x00, 0x02,
		0x80, 0x0A, 0x01, 0x00, 0x00, 0x01,
		0x50, 0x0E, 0x0E, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	d.SendCommand(LUT_FOR_VCOM) //vcom
	for count := 0; count < 44; count++ {
		d.SendData(lut_vcom0[count])
	}

	d.SendCommand(LUT_WHITE_TO_WHITE) //ww --
	for count := 0; count < 42; count++ {
		d.SendData(lut_ww[count])
	}

	d.SendCommand(LUT_BLACK_TO_WHITE) //bw r
	for count := 0; count < 42; count++ {
		d.SendData(lut_bw[count])
	}

	d.SendCommand(LUT_WHITE_TO_BLACK) //wb w
	for count := 0; count < 42; count++ {
		d.SendData(lut_bb[count])
	}

	d.SendCommand(LUT_BLACK_TO_BLACK) //bb b
	for count := 0; count < 42; count++ {
		d.SendData(lut_wb[count])
	}
}

// SetPixel modifies the internal buffer in a single pixel.
// The display have 2 colors: black and white
// We use RGBA(0,0,0, 255) as white (transparent)
// Anything else as black
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	x, y = d.xy(x, y)
	if x < 0 || x >= d.logicalWidth || y < 0 || y >= d.height {
		return
	}
	byteIndex := (uint32(x) + uint32(y)*uint32(d.logicalWidth)) / 8
	if c.R == 0 && c.G == 0 && c.B == 0 { // TRANSPARENT / WHITE
		d.buffer[byteIndex] |= 0x80 >> uint8(x%8)
	} else { // WHITE / EMPTY
		d.buffer[byteIndex] &^= 0x80 >> uint8(x%8)
	}
}

// Display sends the buffer to the screen.
func (d *Device) Display() error {
	d.SendCommand(RESOLUTION_SETTING)
	d.SendData(uint8(d.height >> 8))
	d.SendData(uint8(d.logicalWidth & 0xff))
	d.SendData(uint8(d.height >> 8))
	d.SendData(uint8(d.height & 0xff))

	d.SendCommand(VCM_DC_SETTING)
	d.SendData(0x12)

	d.SendCommand(VCOM_AND_DATA_INTERVAL_SETTING)
	d.SendCommand(0x97) //VBDF 17|D7 VBDW 97  VBDB 57  VBDF F7  VBDW 77  VBDB 37  VBDR B7

	d.SendCommand(DATA_START_TRANSMISSION_1)
	var i int16
	for i = 0; i < d.logicalWidth/8*d.height; i++ {
		d.SendData(0xFF) // bit set: white, bit reset: black
	}
	time.Sleep(2 * time.Millisecond)
	d.SendCommand(DATA_START_TRANSMISSION_2)
	for i = 0; i < d.logicalWidth/8*d.height; i++ {
		d.SendData(d.buffer[i])
	}
	time.Sleep(2 * time.Millisecond)

	d.SetLUT()

	d.SendCommand(DISPLAY_REFRESH)
	time.Sleep(100 * time.Millisecond)
	d.WaitUntilIdle()

	return nil
}

// ClearDisplay erases the device SRAM
func (d *Device) ClearDisplay() {
	d.SendCommand(RESOLUTION_SETTING)
	d.SendData(uint8(d.height >> 8))
	d.SendData(uint8(d.logicalWidth & 0xff))
	d.SendData(uint8(d.height >> 8))
	d.SendData(uint8(d.height & 0xff))

	d.SendCommand(DATA_START_TRANSMISSION_1)
	time.Sleep(2 * time.Millisecond)
	var i int16
	for i = 0; i < d.logicalWidth/8*d.height; i++ {
		d.SendData(0xFF)
	}
	time.Sleep(2 * time.Millisecond)
	d.SendCommand(DATA_START_TRANSMISSION_2)
	time.Sleep(2 * time.Millisecond)
	for i = 0; i < d.logicalWidth/8*d.height; i++ {
		d.SendData(0xFF)
	}
	time.Sleep(2 * time.Millisecond)

	d.SetLUT()
	d.SendCommand(DISPLAY_REFRESH)
	time.Sleep(100 * time.Millisecond)
	d.WaitUntilIdle()
}

// WaitUntilIdle waits until the display is ready
func (d *Device) WaitUntilIdle() {
	for d.busy.Get() {
		time.Sleep(100 * time.Millisecond)
	}
}

// IsBusy returns the busy status of the display
func (d *Device) IsBusy() bool {
	return d.busy.Get()
}

// ClearBuffer sets the buffer to 0xFF (white)
func (d *Device) ClearBuffer() {
	for i := uint32(0); i < d.bufferLength; i++ {
		d.buffer[i] = 0xFF
	}
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	if d.rotation == ROTATION_90 || d.rotation == ROTATION_270 {
		return d.height, d.logicalWidth
	}
	return d.logicalWidth, d.height
}

// SetRotation changes the rotation (clock-wise) of the device
func (d *Device) SetRotation(rotation Rotation) {
	d.rotation = rotation
}

// xy chages the coordinates according to the rotation
func (d *Device) xy(x, y int16) (int16, int16) {
	switch d.rotation {
	case NO_ROTATION:
		return x, y
	case ROTATION_90:
		return d.width - y - 1, x
	case ROTATION_180:
		return d.width - x - 1, d.height - y - 1
	case ROTATION_270:
		return y, d.height - x - 1
	}
	return x, y
}
