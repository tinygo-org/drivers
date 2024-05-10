// Package uc8151 implements a driver for e-ink displays controlled by UC8151
//
// Inspired by https://github.com/pimoroni/pimoroni-pico/blob/main/drivers/uc8151/uc8151.cpp
// Additional inspiration from https://github.com/antirez/uc8151_micropython
// Datasheet: https://www.buydisplay.com/download/ic/UC8151C.pdf
package uc8151 // import "tinygo.org/x/drivers/uc8151"

import (
	"errors"
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/pixel"
)

var (
	errOutOfRange = errors.New("out of screen range")
)

type Config struct {
	Width       int16
	Height      int16
	Rotation    drivers.Rotation // Rotation is clock-wise
	Speed       Speed            // Value from DEFAULT, SLOW, MEDIUM, FAST, FASTER, TURBO
	Blocking    bool             // block on calls to display or return immediately
	FlickerFree bool             // if we should avoid flickering
	UpdateAfter int              // if we are using flicker-free mode, how often we should update the screen
}

type Device struct {
	bus                      drivers.SPI
	cs                       machine.Pin
	dc                       machine.Pin
	rst                      machine.Pin
	busy                     machine.Pin
	width                    int16
	height                   int16
	buffer                   []uint8
	bufferLength             uint32
	rotation                 drivers.Rotation
	speed                    Speed
	blocking                 bool
	flickerFree              bool
	updateCount, updateAfter int
}

type Speed uint8

// New returns a new uc8151 driver. Pass in a fully configured SPI bus.
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
	d.flickerFree = cfg.FlickerFree
	d.updateAfter = cfg.UpdateAfter
	d.bufferLength = (uint32(d.width) * uint32(d.height)) / 8
	d.buffer = make([]uint8, d.bufferLength)
	for i := uint32(0); i < d.bufferLength; i++ {
		d.buffer[i] = 0xFF
	}

	d.Reset()

	d.SendCommand(PSR)
	if d.speed == 0 {
		d.SendData(RES_128x296 | LUT_OTP | FORMAT_BW | SHIFT_RIGHT | BOOSTER_ON | RESET_NONE | SCAN_UP)
	} else {
		d.SendData(RES_128x296 | LUT_REG | FORMAT_BW | SHIFT_RIGHT | BOOSTER_ON | RESET_NONE | SCAN_UP)
	}

	d.SetLUT(d.speed, d.flickerFree)

	d.SendCommand(PWR)
	d.SendData(VDS_INTERNAL | VDG_INTERNAL)
	d.SendData(VCOM_VD | VGHL_16V)
	d.SendData(0b100110) // +10v VDH
	d.SendData(0b100110) // -10v VDL
	d.SendData(0b000011) // VDHR default (For red pixels, not used here)

	d.SendCommand(PON)
	d.WaitUntilIdle()

	d.SendCommand(BTST)
	d.SendData(START_10MS | STRENGTH_3 | OFF_6_58US)
	d.SendData(START_10MS | STRENGTH_3 | OFF_6_58US)
	d.SendData(START_10MS | STRENGTH_3 | OFF_6_58US)

	d.SendCommand(PFS)
	d.SendData(FRAMES_4)

	d.SendCommand(TSE)
	d.SendData(TEMP_INTERNAL | OFFSET_0)

	d.SendCommand(TCON)
	d.SendData(0x22)

	d.SendCommand(CDI)
	d.SendData(0b11_00_1100)

	d.SendCommand(PLL)
	d.SendData(HZ_100)

	d.SendCommand(POF)
	d.WaitUntilIdle()
}

// Reset resets the device
func (d *Device) Reset() {
	d.rst.Low()
	time.Sleep(10 * time.Millisecond)
	d.rst.High()
	time.Sleep(10 * time.Millisecond)
	d.WaitUntilIdle()
}

// PowerOff power off the device
func (d *Device) PowerOff() {
	d.SendCommand(POF)
}

// PowerOn power on the device
func (d *Device) PowerOn() {
	d.SendCommand(PON)
}

// SendCommand sends a command to the display
func (d *Device) SendCommand(command uint8) {
	d.dc.Low()
	d.cs.Low()
	d.bus.Transfer(command)
	d.cs.High()
}

// SendData sends a data byte to the display
func (d *Device) SendData(data ...uint8) {
	d.dc.High()
	d.cs.Low()
	d.bus.Tx(data, nil)
	d.cs.High()
}

// SetPixel modifies the internal buffer in a single pixel.
// The display have 2 colors: black and white
// We use RGBA(0, 0, 0) as white (transparent)
// Anything else as black
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	x, y = d.xy(x, y)

	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return
	}
	byteIndex := x/8 + y*(d.width/8)
	if c.R != 0 || c.G != 0 || c.B != 0 {
		d.buffer[byteIndex] |= 0x80 >> uint8(x%8)
	} else {
		d.buffer[byteIndex] &^= 0x80 >> uint8(x%8)
	}
}

// DrawBitmap copies the bitmap to the screen at the given coordinates.
func (d *Device) DrawBitmap(x, y int16, bitmap pixel.Image[pixel.Monochrome]) error {
	dw, dh := d.Size()
	bw, bh := bitmap.Size()
	if x < 0 || x+int16(bw) > dw || y < 0 || y+int16(bh) > dh {
		return errOutOfRange
	}

	for i := 0; i < bw; i++ {
		for j := 0; j < bh; j++ {
			d.SetPixel(x+int16(i), y+int16(j), bitmap.Get(i, j).RGBA())
		}
	}

	return nil
}

// Display sends the buffer to the screen.
func (d *Device) Display() error {
	if d.blocking {
		d.WaitUntilIdle()
	}

	if d.flickerFree && d.updateAfter != 0 && d.updateCount%d.updateAfter == 0 {
		// we need full refresh here
		d.SetLUT(MEDIUM, false)
	} else {
		d.SetLUT(d.speed, d.flickerFree)
	}
	d.updateCount++

	d.PowerOn()

	d.SendCommand(PTOU)
	d.SendCommand(DTM2)
	d.SendData(d.buffer...)

	d.SendCommand(DSP)
	d.SendCommand(DRF)

	d.SetLUT(d.speed, d.flickerFree)

	if d.blocking {
		d.WaitUntilIdle()
		d.PowerOff()
	}

	return nil
}

// DisplayRect sends only an area of the buffer to the screen.
// The rectangle points need to be a multiple of 8 in the screen.
// They might not work as expected if the screen is rotated.
func (d *Device) DisplayRect(x int16, y int16, width int16, height int16) error {
	if d.blocking {
		d.WaitUntilIdle()
	}

	x, y = d.xy(x, y)
	if x < 0 || y < 0 || x >= d.width || y >= d.height || width < 0 || height < 0 {
		return errors.New("wrong rectangle")
	}
	switch d.rotation {
	case drivers.Rotation0:
		width, height = height, width
		x -= width
	case drivers.Rotation90:
		x -= width - 1
		y -= height - 1
	case drivers.Rotation180:
		width, height = height, width
		y -= height
	}
	x &= 0xF8
	width &= 0xF8
	width = x + width // reuse variables
	if width >= d.width {
		width = d.width
	}
	height = y + height
	if height > d.height {
		height = d.height
	}

	d.SendCommand(PON)
	d.SendCommand(PTIN)
	d.SendCommand(PTL)

	d.SendData(uint8(x))
	d.SendData(uint8(x+width-1) | 0x07)
	d.SendData(uint8(y >> 8))
	d.SendData(uint8(y))
	d.SendData(uint8((y + height - 1) >> 8))
	d.SendData(uint8(y + height - 1))
	d.SendData(0x01)

	d.SendCommand(DTM2)
	x = x / 8
	width = width / 8
	for ; y < height; y++ {
		for i := x; i < width; i++ {
			d.SendData(d.buffer[i+y*(d.width/8)])
		}
	}

	d.SendCommand(DSP)
	d.SendCommand(DRF)

	if d.blocking {
		d.WaitUntilIdle()
		d.PowerOff()
	}
	return nil
}

// ClearDisplay erases the device SRAM
func (d *Device) ClearDisplay() {
	ff := d.flickerFree
	d.flickerFree = false
	defer func() {
		d.flickerFree = ff
	}()

	d.ClearBuffer()
	d.Display()
}

// WaitUntilIdle waits until the display is ready
func (d *Device) WaitUntilIdle() {
	for !d.busy.Get() {
		time.Sleep(10 * time.Millisecond)
	}
}

// IsBusy returns the busy status of the display
func (d *Device) IsBusy() bool {
	return d.busy.Get()
}

// ClearBuffer sets the buffer to 0xFF (white)
func (d *Device) ClearBuffer() {
	for i := uint32(0); i < d.bufferLength; i++ {
		d.buffer[i] = 0x00
	}
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	if d.rotation == drivers.Rotation90 || d.rotation == drivers.Rotation270 {
		return d.height, d.width
	}
	return d.width, d.height
}

// Rotation returns the currently configured rotation.
func (d *Device) Rotation() drivers.Rotation {
	return d.rotation
}

// SetRotation changes the rotation (clock-wise) of the device
func (d *Device) SetRotation(rotation drivers.Rotation) error {
	d.rotation = rotation
	return nil
}

// Set the sleep mode for this display.
func (d *Device) Sleep(sleepEnabled bool) error {
	if sleepEnabled {
		d.PowerOff()
		return nil
	}

	d.PowerOn()
	return nil
}

// SetBlocking changes the blocking flag of the device
func (d *Device) SetBlocking(blocking bool) {
	d.blocking = blocking
}

// xy changes the coordinates according to the rotation
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

// SetSpeed changes the refresh speed of the device (the display needs to re-configure)
func (d *Device) SetSpeed(speed Speed) {
	d.Configure(Config{
		Width:    d.width,
		Height:   d.height,
		Rotation: d.rotation,
		Speed:    speed,
		Blocking: d.blocking,
	})
}

// Invert sets the display' invert mode
func (d *Device) Invert(invert bool) {
	if invert {
		d.SendData(0x5C)
	} else {
		d.SendData(0x4C)
	}
}

// SetLUT sets the look up tables for full or partial updates based on
// the speed and flicker-free mode.
// Based on code from https://github.com/antirez/uc8151_micropython
func (d *Device) SetLUT(speed Speed, flickerFree bool) error {
	var lut LUTSet

	// Num. of frames for single direction change.
	period := 64
	p := uint8(period / (2 ^ (int(speed) - 1)))
	if p < 1 {
		p = 1
	}

	// Num. of frames for back-and-forth change.
	hperiod := period % 2
	hp := uint8(hperiod / (2 ^ (int(speed) - 1)))
	if hp < 1 {
		hp = 1
	}

	if speed < FAST && !flickerFree {
		// For low speed everything is charge-neutral, even WB/BW.

		// Phase 1: long go-inverted-color.
		lut.VCOM.SetRow(0, 0x00, [4]uint8{p, 0x00, 0x00, 0x00}, 0x02)
		lut.BW.SetRow(0, 0b01_000000, [4]uint8{p, 0x00, 0x00, 0x00}, 0x02)
		lut.WB.SetRow(0, 0b10_000000, [4]uint8{p, 0x00, 0x00, 0x00}, 0x02)

		// Phase 2: short ping/pong.
		lut.VCOM.SetRow(1, 0x00, [4]uint8{hp, hp, 0x00, 0x00}, 0x02)
		lut.BW.SetRow(1, 0b10_01_0000, [4]uint8{hp, hp, 0x00, 0x00}, 0x01)
		lut.WB.SetRow(1, 0b01_10_0000, [4]uint8{hp, hp, 0x00, 0x00}, 0x01)

		// Phase 3: long go-target-color.
		lut.VCOM.SetRow(2, 0x00, [4]uint8{p, 0x00, 0x00, 0x00}, 0x02)
		lut.BW.SetRow(2, 0b10_000000, [4]uint8{p, 0x00, 0x00, 0x00}, 0x02)
		lut.WB.SetRow(2, 0b01_000000, [4]uint8{p, 0x00, 0x00, 0x00}, 0x02)

		// For this speed, we use the same LUTs for WW/BB as well.
		copy(lut.WW[:], lut.BW[:])
		copy(lut.BB[:], lut.WB[:])
	} else {
		// Speed >= FAST
		// For greater than 3 we use non charge-neutral LUTs for WB/BW
		// since the inpulse is short and it gets reversed when the
		// pixel changes color, so that's not a problem for the display,
		// however we still need to use charge-neutral LUTs for WW/BB.
		lut.VCOM.SetRow(0, 0x00, [4]uint8{p, p, p, p}, 0x01)
		lut.BW.SetRow(0, 0b10_00_00_00, [4]uint8{p * 4, 0x00, 0x00, 0x00}, 0x01)
		lut.WB.SetRow(0, 0b01_00_00_00, [4]uint8{p * 4, 0x00, 0x00, 0x00}, 0x01)
		lut.WW.SetRow(0, 0b01_10_00_00, [4]uint8{p * 2, p * 2, 0x00, 0x00}, 0x01)
		lut.BB.SetRow(0, 0b10_01_00_00, [4]uint8{p * 2, p * 2, 0x00, 0x00}, 0x01)
	}

	if flickerFree {
		// If no flickering mode is enabled, we use an empty
		// waveform BB and WW. The screen will need to be periodically fully refreshed.
		lut.WW.Clear()
		lut.BB.Clear()
	}

	d.SendCommand(LUT_VCOM)
	d.SendData(append(lut.VCOM[:], []uint8{0, 0}...)...)

	d.SendCommand(LUT_BW)
	d.SendData(lut.BW[:]...)

	d.SendCommand(LUT_WB)
	d.SendData(lut.WB[:]...)

	d.SendCommand(LUT_WW)
	d.SendData(lut.WW[:]...)

	d.SendCommand(LUT_BB)
	d.SendData(lut.BB[:]...)

	return nil
}
