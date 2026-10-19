// Package amg88xx provides a driver for the AMG88XX Thermal Camera
//
// Datasheet:
// https://cdn-learn.adafruit.com/assets/assets/000/043/261/original/Grid-EYE_SPECIFICATIONS%28Reference%29.pdf
package amg88xx // import "tinygo.org/x/drivers/amg88xx"

import (
	"time"

	"tinygo.org/x/drivers"
)

// Device wraps an I2C connection to a AMG88xx device.
type Device struct {
	bus             drivers.I2C
	Address         uint16
	data            [128]uint8
	buf             [7]uint8
	interruptMode   InterruptMode
	interruptEnable uint8
}

type InterruptMode uint8

type Config struct {
}

// New creates a new AMG88xx connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: AddressHigh,
	}
}

// Configure sets up the device for communication
func (d *Device) Configure(cfg Config) error {
	if err := d.SetPCTL(NORMAL_MODE); err != nil {
		return err
	}
	if err := d.SetReset(INITIAL_RESET); err != nil {
		return err
	}
	if err := d.SetFrameRate(FPS_10); err != nil {
		return err
	}

	return d.waitFrame()
}

// ReadPixels reads the 64 values (8x8 grid) of the sensor converted to millicelsius
func (d *Device) ReadPixels(buffer *[64]int32) error {
	if err := d.readRegisters(PIXEL_OFFSET, d.data[:]); err != nil {
		return err
	}
	for i := 0; i < 64; i++ {
		raw := uint16(d.data[2*i+1])<<8 | uint16(d.data[2*i])
		// 12 bit two's complement, see datasheet temperature register.
		buffer[i] = int32(int16(raw<<4)>>4) * PIXEL_TEMP_CONVERSION
	}
	return nil
}

// SetPCTL sets the power control mode
func (d *Device) SetPCTL(pctl uint8) error {
	return d.writeRegister(PCTL, pctl)
}

// SetReset sets the reset value
func (d *Device) SetReset(rst uint8) error {
	return d.writeRegister(RST, rst)
}

// SetFrameRate configures the frame rate
func (d *Device) SetFrameRate(framerate uint8) error {
	return d.writeRegister(FPSC, framerate&0x01)
}

// SetMovingAverageMode sets the moving average mode
func (d *Device) SetMovingAverageMode(mode bool) error {
	var value uint8
	if mode {
		value = 1 << 5
	}
	return d.writeRegister(AVE, value)
}

// SetInterruptLevels sets the interrupt levels in millicelsius, with the
// hysteresis set to 95% of the high level.
func (d *Device) SetInterruptLevels(high int32, low int32) error {
	return d.SetInterruptLevelsHysteresis(high, low, (high*95)/100)
}

// SetInterruptLevelsHysteresis sets the interrupt levels and hysteresis in millicelsius
func (d *Device) SetInterruptLevelsHysteresis(high int32, low int32, hysteresis int32) error {
	d.buf[0] = INTHL
	putLevel(d.buf[1:3], high)
	putLevel(d.buf[3:5], low)
	putLevel(d.buf[5:7], hysteresis)
	return d.bus.Tx(d.Address, d.buf[:7], nil)
}

// EnableInterrupt enables the interrupt pin on the device
func (d *Device) EnableInterrupt() error {
	d.interruptEnable = 1
	return d.writeInterruptControl()
}

// DisableInterrupt disables the interrupt pin on the device
func (d *Device) DisableInterrupt() error {
	d.interruptEnable = 0
	return d.writeInterruptControl()
}

// SetInterruptMode sets the interrupt mode
func (d *Device) SetInterruptMode(mode InterruptMode) error {
	d.interruptMode = mode
	return d.writeInterruptControl()
}

// ReadInterruptTable reads which pixels triggered an interrupt.
// Bit n of table[row] is set for the pixel at row*8+n.
func (d *Device) ReadInterruptTable(table *[8]uint8) error {
	return d.readRegisters(INT_OFFSET, table[:])
}

// PixelInterrupt reports whether a pixel is flagged in an interrupt table.
func PixelInterrupt(table *[8]uint8, pixel int) bool {
	return table[pixel/8]&(1<<(pixel%8)) != 0
}

// ClearInterrupt clears any triggered interrupts
func (d *Device) ClearInterrupt() error {
	return d.SetReset(FLAG_RESET)
}

// ReadStatus reads the status register. Check the result with the
// STATUS_INTF, STATUS_OVF_IRS and STATUS_OVF_THS flags.
func (d *Device) ReadStatus() (uint8, error) {
	if err := d.readRegisters(STAT, d.buf[1:2]); err != nil {
		return 0, err
	}
	return d.buf[1], nil
}

// ClearStatus clears the given status flags.
func (d *Device) ClearStatus(flags uint8) error {
	return d.writeRegister(SCLR, flags&(STATUS_INTF|STATUS_OVF_IRS|STATUS_OVF_THS))
}

// ReadThermistor reads the onboard thermistor in millicelsius
func (d *Device) ReadThermistor() (int32, error) {
	if err := d.readRegisters(TTHL, d.buf[1:3]); err != nil {
		return 0, err
	}
	// Sign and absolute value format, see datasheet thermistor register.
	raw := int32(uint16(d.buf[2])<<8|uint16(d.buf[1])) & 0x07FF
	if d.buf[2]&0x08 != 0 {
		raw = -raw
	}
	return raw * THERMISTOR_CONVERSION / 10, nil
}

// waitFrame waits for the first frame after a reset, since the pixel
// registers read as zero until then. See datasheet section (2) Reset Register.
func (d *Device) waitFrame() error {
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		if err := d.readRegisters(PIXEL_OFFSET, d.data[:]); err != nil {
			return err
		}
		for _, b := range d.data {
			if b != 0 {
				return nil
			}
		}
	}
	return nil
}

func (d *Device) writeInterruptControl() error {
	return d.writeRegister(INTC, ((uint8(d.interruptMode)<<1)|d.interruptEnable)&0x03)
}

func (d *Device) writeRegister(reg, value uint8) error {
	d.buf[0] = reg
	d.buf[1] = value
	return d.bus.Tx(d.Address, d.buf[:2], nil)
}

func (d *Device) readRegisters(reg uint8, data []uint8) error {
	d.buf[0] = reg
	return d.bus.Tx(d.Address, d.buf[:1], data)
}

// putLevel stores a millicelsius level as 12 bit two's complement.
func putLevel(b []uint8, level int32) {
	level /= PIXEL_TEMP_CONVERSION
	if level < -2048 {
		level = -2048
	}
	if level > 2047 {
		level = 2047
	}
	b[0] = uint8(level)
	b[1] = uint8(level>>8) & 0x0F
}
