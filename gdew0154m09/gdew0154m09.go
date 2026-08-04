// Package gdew0154m09 provides a driver for the Good Display GDEW0154M09
// 1.54-inch 200x200 monochrome e-paper panel used by the M5Stack CoreInk.
//
// The initialization sequence is based on the MIT-licensed M5Stack
// M5Core-Ink library:
// https://github.com/m5stack/M5Core-Ink
package gdew0154m09 // import "tinygo.org/x/drivers/gdew0154m09"

import (
	"errors"
	"image/color"
	"time"

	"tinygo.org/x/drivers"
)

const (
	// Width and Height are the physical panel dimensions in pixels.
	Width  = 200
	Height = 200

	// Baudrate and SPIMode are the recommended SPI bus configuration.
	Baudrate = 10_000_000
	SPIMode  = 3

	frameBufferSize = Width * Height / 8
)

var (
	ErrBusyTimeout        = errors.New("gdew0154m09: busy timeout")
	ErrInvalidBusyTimeout = errors.New("gdew0154m09: busy timeout must not be negative")
)

// OutputPin is the interface required from an output pin. Pins must be
// configured by the caller before they are passed to New.
type OutputPin interface {
	Set(level bool)
}

// InputPin is the interface required from an input pin. The pin must be
// configured by the caller before it is passed to New.
type InputPin interface {
	Get() bool
}

// Config contains the display configuration.
type Config struct {
	// BusyTimeout is the maximum time to wait for a panel operation. A zero
	// duration selects the default timeout.
	BusyTimeout time.Duration
}

// DefaultConfig contains the recommended display configuration.
var DefaultConfig = Config{
	BusyTimeout: 5 * time.Second,
}

// Device is a GDEW0154M09 display connected over SPI.
type Device struct {
	bus  drivers.SPI
	cs   OutputPin
	dc   OutputPin
	rst  OutputPin
	busy InputPin

	buffer      [frameBufferSize]byte
	previous    [frameBufferSize]byte
	tx          [3]byte
	busyTimeout time.Duration
}

var _ drivers.Displayer = (*Device)(nil)

// New returns a new GDEW0154M09 driver. The SPI bus and pins must already be
// configured. New performs no I/O.
func New(bus drivers.SPI, cs, dc, rst OutputPin, busy InputPin) *Device {
	d := &Device{
		bus:         bus,
		cs:          cs,
		dc:          dc,
		rst:         rst,
		busy:        busy,
		busyTimeout: DefaultConfig.BusyTimeout,
	}
	d.ClearBuffer()
	fill(d.previous[:], 0xff)
	return d
}

// Configure resets and initializes the display controller. It also clears the
// current and previous framebuffers to white without refreshing the panel.
func (d *Device) Configure(config Config) error {
	if config.BusyTimeout < 0 {
		return ErrInvalidBusyTimeout
	}
	if config.BusyTimeout == 0 {
		config.BusyTimeout = DefaultConfig.BusyTimeout
	}
	d.busyTimeout = config.BusyTimeout

	d.hardwareReset()
	if err := d.waitUntilIdle(); err != nil {
		return err
	}

	if err := d.sendCommand2(commandPanelSetting, 0xdf, 0x0e); err != nil {
		return err
	}
	if err := d.sendCommand1(commandFITIInternalCode, 0x55); err != nil {
		return err
	}
	if err := d.sendCommand1(commandUnknownAA, 0x0f); err != nil {
		return err
	}
	if err := d.sendCommand1(commandUnknownE9, 0x02); err != nil {
		return err
	}
	if err := d.sendCommand1(commandBoosterSoftStart, 0x11); err != nil {
		return err
	}
	if err := d.sendCommand1(commandPowerSequence, 0x0a); err != nil {
		return err
	}
	if err := d.sendCommand3(commandResolutionSetting, 0xc8, 0x00, 0xc8); err != nil {
		return err
	}
	if err := d.sendCommand1(commandTCONSetting, 0x00); err != nil {
		return err
	}
	if err := d.sendCommand1(commandVCOMDataInterval, defaultVCOMDataInterval); err != nil {
		return err
	}
	if err := d.sendCommand1(commandPowerSaving, 0x00); err != nil {
		return err
	}
	if err := d.sendCommand(commandPowerOn); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)
	if err := d.waitUntilIdle(); err != nil {
		return err
	}
	d.ClearBuffer()
	fill(d.previous[:], 0xff)
	return nil
}

// Size returns the dimensions of the display in pixels.
func (d *Device) Size() (width, height int16) {
	return Width, Height
}

// SetPixel updates one pixel in the framebuffer. Transparent and light colors
// are rendered white; opaque dark colors are rendered black.
func (d *Device) SetPixel(x, y int16, c color.RGBA) {
	if x < 0 || x >= Width || y < 0 || y >= Height {
		return
	}
	index := int(y)*(Width/8) + int(x)/8
	mask := byte(0x80 >> uint(x%8))
	brightness := uint16(c.R) + uint16(c.G) + uint16(c.B)
	if c.A == 0 || brightness >= 3*128 {
		d.buffer[index] |= mask
	} else {
		d.buffer[index] &^= mask
	}
}

// ClearBuffer sets every pixel in the framebuffer to white without refreshing
// the panel.
func (d *Device) ClearBuffer() {
	fill(d.buffer[:], 0xff)
}

// Display sends the previous and current framebuffers to the panel and starts
// a full refresh. It blocks until the refresh completes or the busy timeout is
// reached.
func (d *Device) Display() error {
	if err := d.sendCommand(commandDataStartOld); err != nil {
		return err
	}
	if err := d.sendData(d.previous[:]); err != nil {
		return err
	}
	time.Sleep(2 * time.Millisecond)

	if err := d.sendCommand(commandDataStartNew); err != nil {
		return err
	}
	if err := d.sendData(d.buffer[:]); err != nil {
		return err
	}
	time.Sleep(2 * time.Millisecond)

	if err := d.sendCommand(commandDisplayRefresh); err != nil {
		return err
	}
	if err := d.waitUntilIdle(); err != nil {
		return err
	}
	copy(d.previous[:], d.buffer[:])
	return nil
}

// ClearDisplay clears the framebuffer and refreshes the panel.
func (d *Device) ClearDisplay() error {
	d.ClearBuffer()
	return d.Display()
}

// DeepSleep powers off the panel and enters deep sleep. Configure must be
// called to wake and reinitialize the controller.
func (d *Device) DeepSleep() error {
	if err := d.sendCommand1(commandVCOMDataInterval, deepSleepVCOMDataInterval); err != nil {
		return err
	}
	if err := d.waitUntilIdle(); err != nil {
		return err
	}
	if err := d.sendCommand(commandPowerOff); err != nil {
		return err
	}
	return d.sendCommand1(commandDeepSleep, deepSleepCheckCode)
}

// IsBusy reports whether the active-low busy signal is asserted.
func (d *Device) IsBusy() bool {
	return !d.busy.Get()
}

func (d *Device) hardwareReset() {
	d.rst.Set(true)
	time.Sleep(10 * time.Millisecond)
	d.rst.Set(false)
	time.Sleep(100 * time.Millisecond)
	d.rst.Set(true)
	time.Sleep(100 * time.Millisecond)
}

func (d *Device) waitUntilIdle() error {
	deadline := time.Now().Add(d.busyTimeout)
	for d.IsBusy() {
		if time.Now().After(deadline) {
			return ErrBusyTimeout
		}
		time.Sleep(time.Millisecond)
	}
	return nil
}

func (d *Device) sendCommandWithData(command byte, data []byte) error {
	if err := d.sendCommand(command); err != nil {
		return err
	}
	return d.sendData(data)
}

func (d *Device) sendCommand1(command, data byte) error {
	d.tx[0] = data
	return d.sendCommandWithData(command, d.tx[:1])
}

func (d *Device) sendCommand2(command, data0, data1 byte) error {
	d.tx[0] = data0
	d.tx[1] = data1
	return d.sendCommandWithData(command, d.tx[:2])
}

func (d *Device) sendCommand3(command, data0, data1, data2 byte) error {
	d.tx[0] = data0
	d.tx[1] = data1
	d.tx[2] = data2
	return d.sendCommandWithData(command, d.tx[:3])
}

func (d *Device) sendCommand(command byte) error {
	d.dc.Set(false)
	d.cs.Set(false)
	_, err := d.bus.Transfer(command)
	d.cs.Set(true)
	return err
}

func (d *Device) sendData(data []byte) error {
	d.dc.Set(true)
	d.cs.Set(false)
	err := d.bus.Tx(data, nil)
	d.cs.Set(true)
	return err
}

func fill(buffer []byte, value byte) {
	for index := range buffer {
		buffer[index] = value
	}
}
