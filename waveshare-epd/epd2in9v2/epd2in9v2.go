// Package epd2in9v2 implements a driver for the Waveshare 2.9in V2 black and white e-paper display.
//
// This is for the V2 device using the SSD1680 chipset. For the V1 device (using IL3820),
// use the epd2in9 package instead.
//
// Datasheet:
// https://files.waveshare.com/upload/7/79/2.9inch-e-paper-v2-specification.pdf
// https://cdn-learn.adafruit.com/assets/assets/000/097/631/original/SSD1680_Datasheet.pdf?1607625960
//
// Reference: https://github.com/waveshareteam/e-Paper/tree/master/RaspberryPi_JetsonNano/c/lib/e-Paper
package epd2in9v2 // import "tinygo.org/x/drivers/waveshare-epd/epd2in9v2"

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
)

type Config struct {
	Width    int16
	Height   int16
	Rotation Rotation
	Speed    Speed
	Blocking bool
}

type Device struct {
	bus          drivers.SPI
	cs           machine.Pin
	dc           machine.Pin
	rst          machine.Pin
	busy         machine.Pin
	width        int16
	height       int16
	buffer       []uint8
	bufferLength uint32
	rotation     Rotation
	speed        Speed
	blocking     bool
}

type Rotation uint8
type Speed uint8

// LUT for normal full refresh (~2s)
var lutDefault = [159]uint8{
	0x80, 0x66, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x0, 0x0, 0x0,
	0x10, 0x66, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20, 0x0, 0x0, 0x0,
	0x80, 0x66, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x0, 0x0, 0x0,
	0x10, 0x66, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x14, 0x8, 0x0, 0x0, 0x0, 0x0, 0x1,
	0xA, 0xA, 0x0, 0xA, 0xA, 0x0, 0x1,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x14, 0x8, 0x0, 0x1, 0x0, 0x0, 0x1,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x0, 0x0, 0x0,
	0x22, 0x17, 0x41, 0x0, 0x32, 0x36,
}

// LUT for fast full refresh (~1s)
var lutFast = [159]uint8{
	0x90, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x60, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x90, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x60, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x19, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x24, 0x42, 0x22, 0x22, 0x23, 0x32, 0x00, 0x00, 0x00,
	0x22, 0x17, 0x41, 0xAE, 0x32, 0x38,
}

// LUT for partial refresh
var lutPartial = [159]uint8{
	0x0, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x80, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x40, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0A, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2,
	0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x0, 0x0, 0x0,
	0x22, 0x17, 0x41, 0xB0, 0x32, 0x36,
}

// New returns a new epd2in9v2 driver. Pass in a fully configured SPI bus.
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
	d.speed = cfg.Speed
	d.blocking = cfg.Blocking
	d.bufferLength = (uint32(d.width) * uint32(d.height)) / 8
	d.buffer = make([]uint8, d.bufferLength)
	for i := uint32(0); i < d.bufferLength; i++ {
		d.buffer[i] = 0xFF
	}

	d.Reset()
	time.Sleep(100 * time.Millisecond)

	d.WaitUntilIdle()
	d.SendCommand(SW_RESET)
	d.WaitUntilIdle()

	d.SendCommand(DRIVER_OUTPUT_CONTROL)
	d.SendData(uint8((d.height - 1) & 0xFF))
	d.SendData(uint8((d.height - 1) >> 8))
	d.SendData(0x00)

	d.SendCommand(DATA_ENTRY_MODE)
	d.SendData(0x03)

	d.setWindow(0, 0, d.width-1, d.height-1)

	if cfg.Speed == SPEED_FAST {
		d.SendCommand(BORDER_WAVEFORM_CONTROL)
		d.SendData(0x05)
	}

	d.SendCommand(DISPLAY_UPDATE_CONTROL_1)
	d.SendData(0x00)
	d.SendData(0x80)

	d.setCursor(0, 0)
	d.WaitUntilIdle()

	switch cfg.Speed {
	case SPEED_FAST:
		d.setLUTByHost(&lutFast)
	default:
		d.setLUTByHost(&lutDefault)
	}
}

// HardwareReset resets the device via the RST pin.
func (d *Device) Reset() {
	d.rst.High()
	time.Sleep(10 * time.Millisecond)
	d.rst.Low()
	time.Sleep(2 * time.Millisecond)
	d.rst.High()
	time.Sleep(10 * time.Millisecond)
}

// SendCommand sends a command byte to the display.
func (d *Device) SendCommand(command uint8) {
	d.dc.Low()
	d.cs.Low()
	d.bus.Transfer(command)
	d.cs.High()
}

// SendData sends a data byte to the display.
func (d *Device) SendData(data uint8) {
	d.dc.High()
	d.cs.Low()
	d.bus.Transfer(data)
	d.cs.High()
}

// WaitUntilIdle waits until the display is ready.
// On SSD1680, BUSY pin is HIGH when busy, LOW when idle.
func (d *Device) WaitUntilIdle() {
	for d.busy.Get() {
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(50 * time.Millisecond)
}

// IsBusy returns the busy status of the display.
func (d *Device) IsBusy() bool {
	return d.busy.Get()
}

// SetPixel modifies the internal buffer in a single pixel.
// Uses color.RGBA where black (0,0,0) = black pixel, anything else = white pixel.
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	x, y = d.xy(x, y)

	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return
	}
	byteIndex := (y * (d.width / 8)) + (x / 8)
	if c.R == 0 && c.G == 0 && c.B == 0 {
		d.buffer[byteIndex] &^= 0x80 >> uint8(x%8)
	} else {
		d.buffer[byteIndex] |= 0x80 >> uint8(x%8)
	}
}

// Display sends the buffer to the screen.
func (d *Device) Display() error {
	if d.blocking {
		d.WaitUntilIdle()
	}

	d.setCursor(0, 0)
	d.SendCommand(WRITE_RAM_BW)
	for i := uint32(0); i < d.bufferLength; i++ {
		d.SendData(d.buffer[i])
	}

	d.turnOnDisplay()

	if d.blocking {
		d.WaitUntilIdle()
	}
	return nil
}

// DisplayWithBase writes the buffer to both BW and RED RAM then refreshes.
// This is useful before partial updates to set the base image.
func (d *Device) DisplayWithBase() error {
	if d.blocking {
		d.WaitUntilIdle()
	}

	d.setCursor(0, 0)
	d.SendCommand(WRITE_RAM_BW)
	for i := uint32(0); i < d.bufferLength; i++ {
		d.SendData(d.buffer[i])
	}

	d.setCursor(0, 0)
	d.SendCommand(WRITE_RAM_RED)
	for i := uint32(0); i < d.bufferLength; i++ {
		d.SendData(d.buffer[i])
	}

	d.turnOnDisplay()

	if d.blocking {
		d.WaitUntilIdle()
	}
	return nil
}

// DisplayPartial performs a partial refresh of the display.
// Call DisplayWithBase first to set the base image before using partial updates.
func (d *Device) DisplayPartial() error {
	d.rst.Low()
	time.Sleep(1 * time.Millisecond)
	d.rst.High()
	time.Sleep(2 * time.Millisecond)

	d.setLUT(&lutPartial)

	d.SendCommand(OTP_SELECTION_CONTROL)
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x40)
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x00)

	d.SendCommand(BORDER_WAVEFORM_CONTROL)
	d.SendData(0x80)

	d.SendCommand(DISPLAY_UPDATE_CONTROL_2)
	d.SendData(0xC0)
	d.SendCommand(MASTER_ACTIVATION)
	d.WaitUntilIdle()

	d.setWindow(0, 0, d.width-1, d.height-1)
	d.setCursor(0, 0)

	d.SendCommand(WRITE_RAM_BW)
	for i := uint32(0); i < d.bufferLength; i++ {
		d.SendData(d.buffer[i])
	}

	d.turnOnDisplayPartial()
	d.WaitUntilIdle()
	return nil
}

// ClearDisplay erases the display.
func (d *Device) ClearDisplay() {
	d.ClearBuffer()
	d.Display()
}

// ClearBuffer sets the buffer to 0xFF (white).
func (d *Device) ClearBuffer() {
	for i := uint32(0); i < d.bufferLength; i++ {
		d.buffer[i] = 0xFF
	}
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	if d.rotation == ROTATION_90 || d.rotation == ROTATION_270 {
		return d.height, d.width
	}
	return d.width, d.height
}

// SetRotation changes the rotation (clock-wise) of the device.
func (d *Device) SetRotation(rotation Rotation) {
	d.rotation = rotation
}

// SetBlocking changes the blocking flag of the device.
func (d *Device) SetBlocking(blocking bool) {
	d.blocking = blocking
}

// SetSpeed changes the refresh speed and reconfigures the device.
func (d *Device) SetSpeed(speed Speed) {
	d.Configure(Config{
		Width:    d.width,
		Height:   d.height,
		Rotation: d.rotation,
		Speed:    speed,
		Blocking: d.blocking,
	})
}

// Sleep puts the display into deep sleep mode. A hardware reset is needed to wake it.
func (d *Device) Sleep() {
	d.SendCommand(DEEP_SLEEP_MODE)
	d.SendData(0x01)
	time.Sleep(100 * time.Millisecond)
}

// PowerOff disables the display analog/clock. Lighter than Sleep.
func (d *Device) PowerOff() {
	d.SendCommand(DISPLAY_UPDATE_CONTROL_2)
	d.SendData(0x03)
	d.SendCommand(MASTER_ACTIVATION)
	d.WaitUntilIdle()
}

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

func (d *Device) setWindow(xStart, yStart, xEnd, yEnd int16) {
	d.SendCommand(SET_RAM_X_ADDRESS)
	d.SendData(uint8((xStart >> 3) & 0xFF))
	d.SendData(uint8((xEnd >> 3) & 0xFF))

	d.SendCommand(SET_RAM_Y_ADDRESS)
	d.SendData(uint8(yStart & 0xFF))
	d.SendData(uint8((yStart >> 8) & 0xFF))
	d.SendData(uint8(yEnd & 0xFF))
	d.SendData(uint8((yEnd >> 8) & 0xFF))
}

func (d *Device) setCursor(x, y int16) {
	d.SendCommand(SET_RAM_X_COUNTER)
	d.SendData(uint8(x & 0xFF))

	d.SendCommand(SET_RAM_Y_COUNTER)
	d.SendData(uint8(y & 0xFF))
	d.SendData(uint8((y >> 8) & 0xFF))
}

func (d *Device) turnOnDisplay() {
	d.SendCommand(DISPLAY_UPDATE_CONTROL_2)
	d.SendData(0xC7)
	d.SendCommand(MASTER_ACTIVATION)
	d.WaitUntilIdle()
}

func (d *Device) turnOnDisplayPartial() {
	d.SendCommand(DISPLAY_UPDATE_CONTROL_2)
	d.SendData(0x0F)
	d.SendCommand(MASTER_ACTIVATION)
	d.WaitUntilIdle()
}

func (d *Device) setLUT(lut *[159]uint8) {
	d.SendCommand(WRITE_LUT_REGISTER)
	for i := 0; i < 153; i++ {
		d.SendData(lut[i])
	}
	d.WaitUntilIdle()
}

func (d *Device) setLUTByHost(lut *[159]uint8) {
	d.setLUT(lut)

	d.SendCommand(END_OPTION)
	d.SendData(lut[153])

	d.SendCommand(GATE_DRIVING_VOLTAGE)
	d.SendData(lut[154])

	d.SendCommand(SOURCE_DRIVING_VOLTAGE)
	d.SendData(lut[155])
	d.SendData(lut[156])
	d.SendData(lut[157])

	d.SendCommand(WRITE_VCOM_REGISTER)
	d.SendData(lut[158])
}
