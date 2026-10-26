// Package epd2in7gray drives the Waveshare 2.7in V1 e-paper panel in four
// shades of gray. It uses 2 bits per pixel, not 16-level grayscale.
package epd2in7gray // import "tinygo.org/x/drivers/waveshare-epd/epd2in7gray"

import (
	"errors"
	"image/color"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/pin"
	"tinygo.org/x/drivers/pixel"
)

const (
	// Width is the native panel width in pixels.
	Width = 176
	// Height is the native panel height in pixels.
	Height = 264
	// DefaultBusyTimeout is the BUSY wait limit when Config.BusyTimeout is zero.
	DefaultBusyTimeout = 30 * time.Second
)

var (
	ErrInvalidDimensions  = errors.New("epd2in7gray: invalid dimensions")
	ErrInvalidRotation    = errors.New("epd2in7gray: unsupported rotation")
	ErrInvalidBusyTimeout = errors.New("epd2in7gray: busy timeout must not be negative")
	ErrBusyTimeout        = errors.New("epd2in7gray: busy timeout")
)

// Config is the configuration for the display.
type Config struct {
	// Width and Height default to 176x264. Width must be a multiple of 8.
	Width  int16
	Height int16
	// LogicalWidth is the framebuffer stride, not the visible width.
	// It defaults to Width and must be a multiple of 8 and at least Width.
	LogicalWidth int16
	Rotation     drivers.Rotation
	// BusyTimeout defaults to DefaultBusyTimeout when zero.
	BusyTimeout time.Duration
	// OnInitStage runs after reset and each initialization command completes.
	// A nil callback disables tracing. The callback may log or pause setup.
	OnInitStage func(stage string)
}

// Device is a Waveshare 2.7in V1 e-paper panel in four-shade mode.
type Device struct {
	bus  drivers.SPI
	cs   pin.OutputFunc
	dc   pin.OutputFunc
	rst  pin.OutputFunc
	busy pin.InputFunc

	width        int16
	height       int16
	logicalWidth int16
	rotation     drivers.Rotation
	busyTimeout  time.Duration
	buffer       []byte
	tx           [64]byte
}

var _ drivers.Displayer = (*Device)(nil)

// New creates a driver with a white framebuffer and performs no I/O.
// The SPI bus and all pins must already be configured.
func New(bus drivers.SPI, cs, dc, rst pin.Output, busy pin.Input) *Device {
	d := &Device{
		bus:          bus,
		cs:           cs.Set,
		dc:           dc.Set,
		rst:          rst.Set,
		busy:         busy.Get,
		width:        Width,
		height:       Height,
		logicalWidth: Width,
		busyTimeout:  DefaultBusyTimeout,
		buffer:       make([]byte, Width*Height/4),
	}
	d.ClearBuffer()
	return d
}

// Configure resets the controller and selects four-shade full refreshes.
// It clears the framebuffer to white but does not refresh the panel.
func (d *Device) Configure(cfg Config) error {
	if cfg.Width == 0 {
		cfg.Width = Width
	}
	if cfg.Height == 0 {
		cfg.Height = Height
	}
	if cfg.LogicalWidth == 0 {
		cfg.LogicalWidth = cfg.Width
	}
	if cfg.Width < 0 || cfg.Width > Width || cfg.Width%8 != 0 ||
		cfg.Height < 0 || cfg.Height > Height ||
		cfg.LogicalWidth < cfg.Width || cfg.LogicalWidth%8 != 0 {
		return ErrInvalidDimensions
	}
	if cfg.Rotation > drivers.Rotation270 {
		return ErrInvalidRotation
	}
	if cfg.BusyTimeout < 0 {
		return ErrInvalidBusyTimeout
	}
	if cfg.BusyTimeout == 0 {
		cfg.BusyTimeout = DefaultBusyTimeout
	}

	d.width = cfg.Width
	d.height = cfg.Height
	d.logicalWidth = cfg.LogicalWidth
	d.rotation = cfg.Rotation
	d.busyTimeout = cfg.BusyTimeout
	size := int(d.logicalWidth) * int(d.height) / 4
	if len(d.buffer) != size {
		d.buffer = make([]byte, size)
	}
	d.ClearBuffer()

	d.cs(true)
	d.dc(false)
	d.rst(true)
	time.Sleep(200 * time.Millisecond)
	d.rst(false)
	time.Sleep(2 * time.Millisecond)
	d.rst(true)
	time.Sleep(200 * time.Millisecond)
	if cfg.OnInitStage != nil {
		cfg.OnInitStage("reset")
	}

	for _, setting := range initialization {
		data := setting.data
		if setting.command == commandResolution {
			d.tx[0] = byte(d.width >> 8)
			d.tx[1] = byte(d.width)
			d.tx[2] = byte(d.height >> 8)
			d.tx[3] = byte(d.height)
			data = d.tx[:4]
		}
		if err := d.writeRegister(setting.command, data); err != nil {
			return err
		}
		if setting.command == commandPowerOn {
			if err := d.waitUntilIdle(); err != nil {
				return err
			}
		}
		if cfg.OnInitStage != nil {
			cfg.OnInitStage(setting.stage)
		}
	}
	return nil
}

// SetPixel sets a buffered pixel using RGB luminance. Alpha is ignored.
// The four shades are black, dark gray, light gray, and white.
func (d *Device) SetPixel(x, y int16, c color.RGBA) {
	w, h := d.Size()
	if x < 0 || y < 0 || x >= w || y >= h {
		return
	}
	switch d.rotation {
	case drivers.Rotation90:
		x, y = d.width-y-1, x
	case drivers.Rotation180:
		x, y = d.width-x-1, d.height-y-1
	case drivers.Rotation270:
		x, y = y, d.height-x-1
	}
	index := (int(y)*int(d.logicalWidth) + int(x)) / 4
	shift := uint(6 - (x%4)*2)
	shade := byte(pixel.NewGrayscale2bit(c.R, c.G, c.B))
	d.buffer[index] = d.buffer[index]&^(3<<shift) | shade<<shift
}

// ClearBuffer fills the framebuffer with white without refreshing the panel.
func (d *Device) ClearBuffer() {
	for i := range d.buffer {
		d.buffer[i] = 0xff
	}
}

// Display sends both grayscale bitplanes and waits for a full refresh.
func (d *Device) Display() error {
	return d.display(false)
}

// ClearDisplay refreshes the panel to white without changing the framebuffer.
func (d *Device) ClearDisplay() error {
	return d.display(true)
}

func (d *Device) display(white bool) error {
	if err := d.sendPlane(commandDataStart1, 1, white); err != nil {
		return err
	}
	if err := d.sendPlane(commandDataStart2, 0, white); err != nil {
		return err
	}
	for _, setting := range grayscaleLUT {
		if err := d.writeRegister(setting.command, setting.data); err != nil {
			return err
		}
	}
	if err := d.sendCommand(commandDisplayRefresh); err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)
	return d.waitUntilIdle()
}

func (d *Device) sendPlane(command byte, bit uint, white bool) error {
	if err := d.sendCommand(command); err != nil {
		return err
	}
	n := 0
	for y := 0; y < int(d.height); y++ {
		row := y * int(d.logicalWidth) / 4
		for x := 0; x < int(d.width)/8; x++ {
			d.tx[n] = 0xff
			if !white {
				index := row + x*2
				d.tx[n] = packPlane(d.buffer[index], d.buffer[index+1], bit)
			}
			n++
			if n == len(d.tx) {
				if err := d.sendData(d.tx[:n]); err != nil {
					return err
				}
				n = 0
			}
		}
	}
	if n != 0 {
		return d.sendData(d.tx[:n])
	}
	return nil
}

func packPlane(first, second byte, bit uint) byte {
	pixels := uint16(first)<<8 | uint16(second)
	var plane byte
	for i := 0; i < 8; i++ {
		plane = plane<<1 | byte(pixels>>(14+bit)&1)
		pixels <<= 2
	}
	return plane
}

// Size returns the visible dimensions after rotation, excluding row padding.
func (d *Device) Size() (w, h int16) {
	if d.rotation == drivers.Rotation90 || d.rotation == drivers.Rotation270 {
		return d.height, d.width
	}
	return d.width, d.height
}

// Rotation returns the clockwise display rotation.
func (d *Device) Rotation() drivers.Rotation {
	return d.rotation
}

// SetRotation changes the drawing rotation without modifying buffered pixels.
func (d *Device) SetRotation(rotation drivers.Rotation) error {
	if rotation > drivers.Rotation270 {
		return ErrInvalidRotation
	}
	d.rotation = rotation
	return nil
}

// DeepSleep powers off the panel while retaining its image.
// Call Configure to wake it and restore the grayscale controller settings.
func (d *Device) DeepSleep() error {
	if err := d.writeRegister(commandVCOMDataInterval, []byte{0xf7}); err != nil {
		return err
	}
	if err := d.sendCommand(commandPowerOff); err != nil {
		return err
	}
	if err := d.waitUntilIdle(); err != nil {
		return err
	}
	return d.writeRegister(commandDeepSleep, []byte{0xa5})
}

func (d *Device) waitUntilIdle() error {
	deadline := time.Now().Add(d.busyTimeout)
	for {
		if err := d.sendCommand(commandGetStatus); err != nil {
			return err
		}
		if d.busy() {
			time.Sleep(200 * time.Millisecond)
			return nil
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return ErrBusyTimeout
		}
		time.Sleep(min(remaining, 10*time.Millisecond))
	}
}

func (d *Device) writeRegister(command byte, data []byte) error {
	if err := d.sendCommand(command); err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	return d.sendData(data)
}

func (d *Device) sendCommand(command byte) error {
	return d.sendByte(false, command)
}

func (d *Device) sendData(data []byte) error {
	for _, value := range data {
		if err := d.sendByte(true, value); err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) sendByte(isData bool, value byte) error {
	d.dc(isData)
	d.cs(false)
	_, err := d.bus.Transfer(value)
	d.cs(true)
	return err
}
