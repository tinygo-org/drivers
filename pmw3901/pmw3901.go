// Package pmw3901 implements a driver for the PixArt PMW3901 optical flow sensor.
//
// Datasheet: https://www.codico.com/de/mpattachment/file/download/id/952/
package pmw3901 // import "tinygo.org/x/drivers/pmw3901"

import (
	"errors"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/pin"
)

var (
	ErrDeviceNotConnected = errors.New("pmw3901: device not connected")
	ErrTimeout            = errors.New("pmw3901: timeout")
)

// Frame dimensions in pixels.
const (
	FrameWidth  = 35
	FrameHeight = 35
	FrameSize   = FrameWidth * FrameHeight
)

// frameTimeout limits the wait for the frame and for each pixel.
const frameTimeout = time.Second

// Motion holds the result of a motion burst read.
type Motion struct {
	Motion      uint8
	Observation uint8
	DeltaX      int16
	DeltaY      int16
	Squal       uint8
	RawDataSum  uint8
	MaxRawData  uint8
	MinRawData  uint8
	Shutter     uint16
}

// Valid reports whether motion was detected with usable surface quality.
func (m Motion) Valid() bool {
	// Same check as get_motion in https://github.com/pimoroni/pmw3901-python
	return m.Motion&motionMOT != 0 && !(m.Squal < 0x19 && m.Shutter>>8 == 0x1F)
}

// Device wraps an SPI connection to a PMW3901 device.
type Device struct {
	bus drivers.SPI
	cs  pin.Output
	buf [12]byte
}

// New creates a new PMW3901 connection. The bus must already be configured
// in mode 3 at up to 2 MHz, and cs as an output. It does not touch the device.
func New(bus drivers.SPI, cs pin.Output) Device {
	return Device{
		bus: bus,
		cs:  cs,
	}
}

// Configure resets the device, checks the product ID, and writes the
// performance optimization registers.
func (d *Device) Configure() error {
	// Reset the SPI port.
	d.cs.Set(true)
	time.Sleep(time.Millisecond)
	d.cs.Set(false)
	time.Sleep(time.Millisecond)
	d.cs.Set(true)
	time.Sleep(time.Millisecond)

	if err := d.WriteRegister(RegPowerUpReset, powerUpResetValue); err != nil {
		return err
	}
	time.Sleep(5 * time.Millisecond)

	if !d.Connected() {
		return ErrDeviceNotConnected
	}

	for reg := uint8(RegMotion); reg <= RegDeltaYH; reg++ {
		if _, err := d.ReadRegister(reg); err != nil {
			return err
		}
	}
	time.Sleep(time.Millisecond)

	if err := d.writeRegisters(initRegisters1); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return d.writeRegisters(initRegisters2)
}

// Connected reports whether a PMW3901 answers.
func (d *Device) Connected() bool {
	id, err := d.ReadRegister(RegProductID)
	if err != nil || id != ProductID {
		return false
	}
	inv, err := d.ReadRegister(RegInverseProductID)
	return err == nil && inv == InverseProductID
}

// ReadMotion returns the motion counts since the last read.
func (d *Device) ReadMotion() (dx, dy int16, err error) {
	// Reading Motion latches the delta registers.
	if _, err = d.ReadRegister(RegMotion); err != nil {
		return
	}
	var b [4]uint8
	for i := range b {
		if b[i], err = d.ReadRegister(RegDeltaXL + uint8(i)); err != nil {
			return
		}
	}
	return int16(uint16(b[0]) | uint16(b[1])<<8), int16(uint16(b[2]) | uint16(b[3])<<8), nil
}

// ReadMotionBurst reads all motion registers in a single transfer.
func (d *Device) ReadMotionBurst() (Motion, error) {
	b := d.buf[:]
	for i := range b {
		b[i] = 0
	}
	d.buf[0] = RegMotionBurst
	d.cs.Set(false)
	err := d.bus.Tx(d.buf[:1], nil)
	if err == nil {
		d.buf[0] = 0
		wait(tSRAD)
		err = d.bus.Tx(b, b)
	}
	d.cs.Set(true)
	wait(tSRR)
	if err != nil {
		return Motion{}, err
	}
	return Motion{
		Motion:      b[0],
		Observation: b[1],
		DeltaX:      int16(uint16(b[2]) | uint16(b[3])<<8),
		DeltaY:      int16(uint16(b[4]) | uint16(b[5])<<8),
		Squal:       b[6],
		RawDataSum:  b[7],
		MaxRawData:  b[8],
		MinRawData:  b[9],
		Shutter:     uint16(b[10])<<8 | uint16(b[11]),
	}, nil
}

// SetLED turns the LED_N output on or off.
func (d *Device) SetLED(on bool) error {
	val := uint8(0x00)
	if on {
		val = 0x1C
	}
	// Register bank and value from the Bitcraze PMW3901 driver, see LICENSE-bitcraze.
	if err := d.WriteRegister(regBank, 0x14); err != nil {
		return err
	}
	if err := d.WriteRegister(0x6F, val); err != nil {
		return err
	}
	return d.WriteRegister(regBank, 0x00)
}

// ReadFrame captures a raw 35x35 frame with 8-bit pixels, row by row.
// It stops motion tracking, call Configure to track motion again.
func (d *Device) ReadFrame(frame *[FrameSize]byte) error {
	// Sequence and pixel format from the Pimoroni driver, see LICENSE-pimoroni.
	if err := d.writeRegisters(frameRegisters1); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)
	if err := d.writeRegisters(frameRegisters2); err != nil {
		return err
	}

	start := time.Now()
	for {
		status, err := d.ReadRegister(RegRawDataGrabStat)
		if err != nil {
			return err
		}
		if status&rawDataMask != 0 {
			break
		}
		if time.Since(start) > frameTimeout {
			return ErrTimeout
		}
	}

	if err := d.WriteRegister(RegRawDataGrab, 0x00); err != nil {
		return err
	}

	start = time.Now()
	for i := 0; i < FrameSize; {
		data, err := d.ReadRegister(RegRawDataGrab)
		if err != nil {
			return err
		}
		switch data & rawDataMask {
		case rawDataUpper:
			frame[i] = frame[i]&0x03 | (data&0x3F)<<2
		case rawDataLower:
			frame[i] = frame[i]&0xFC | (data&0x0C)>>2
			i++
			start = time.Now()
		}
		if time.Since(start) > frameTimeout {
			return ErrTimeout
		}
	}
	return nil
}

// SPI timings, see datasheet section 2.4 table 5.
const (
	tSRAD         = 35 * time.Microsecond
	tSRR          = 20 * time.Microsecond
	tSWW          = 45 * time.Microsecond
	tSCLKNCSWrite = 35 * time.Microsecond
)

// ReadRegister reads an 8-bit register.
func (d *Device) ReadRegister(reg uint8) (uint8, error) {
	d.buf[0] = reg &^ spiWrite
	d.cs.Set(false)
	err := d.bus.Tx(d.buf[:1], nil)
	if err == nil {
		wait(tSRAD)
		d.buf[0] = 0
		err = d.bus.Tx(d.buf[:1], d.buf[:1])
	}
	d.cs.Set(true)
	wait(tSRR)
	return d.buf[0], err
}

// WriteRegister writes an 8-bit register.
func (d *Device) WriteRegister(reg, val uint8) error {
	d.buf[0] = reg | spiWrite
	d.buf[1] = val
	d.cs.Set(false)
	err := d.bus.Tx(d.buf[:2], nil)
	wait(tSCLKNCSWrite)
	d.cs.Set(true)
	wait(tSWW)
	return err
}

// wait busy-waits, since time.Sleep can be much too coarse for these short delays.
func wait(d time.Duration) {
	for start := time.Now(); time.Since(start) < d; {
	}
}

// writeRegisters writes a string of register and value pairs.
func (d *Device) writeRegisters(regs string) error {
	for i := 0; i+1 < len(regs); i += 2 {
		if err := d.WriteRegister(regs[i], regs[i+1]); err != nil {
			return err
		}
	}
	return nil
}
