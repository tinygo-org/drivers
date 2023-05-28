// Package amg88xx provides a driver for the AMG88XX Thermal Camera
//
// Datasheet:
// https://cdn-learn.adafruit.com/assets/assets/000/043/261/original/Grid-EYE_SPECIFICATIONS%28Reference%29.pdf
package amg88xx // import "tinygo.org/x/drivers/amg88xx"

import (
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// Device wraps an I2C connection to a AMG88xx device.
type Device struct {
	bus             drivers.I2C
	Address         uint16
	data            []uint8
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
func (d *Device) Configure(cfg Config) {
	d.data = make([]uint8, 128)

	d.SetPCTL(NORMAL_MODE)
	d.SetReset(INITIAL_RESET)
	d.SetFrameRate(FPS_10)

	time.Sleep(100 * time.Millisecond)
}

// ReadPixels returns the 64 values (8x8 grid) of the sensor converted to  millicelsius
func (d *Device) ReadPixels(buffer *[64]int16) {
	legacy.ReadRegister(d.bus, uint8(d.Address), PIXEL_OFFSET, d.data)
	for i := 0; i < 64; i++ {
		buffer[i] = int16((uint16(d.data[2*i+1]) << 8) | uint16(d.data[2*i]))
		if (buffer[i] & (1 << 11)) > 0 { // temperature negative
			buffer[i] &= ^(1 << 11)
			buffer[i] = -buffer[i]
		}
		buffer[i] *= PIXEL_TEMP_CONVERSION
	}
}

// SetPCTL sets the PCTL
func (d *Device) SetPCTL(pctl uint8) {
	legacy.WriteRegister(d.bus, uint8(d.Address), PCTL, []byte{pctl})
}

// SetReset sets the reset value
func (d *Device) SetReset(rst uint8) {
	legacy.WriteRegister(d.bus, uint8(d.Address), RST, []byte{rst})
}

// SetFrameRate configures the frame rate
func (d *Device) SetFrameRate(framerate uint8) {
	legacy.WriteRegister(d.bus, uint8(d.Address), FPSC, []byte{framerate & 0x01})
}

// SetMovingAverageMode sets the moving average mode
func (d *Device) SetMovingAverageMode(mode bool) {
	var value uint8
	if mode {
		value = 1
	}
	legacy.WriteRegister(d.bus, uint8(d.Address), AVE, []byte{value << 5})
}

// SetInterruptLevels sets the interrupt levels
func (d *Device) SetInterruptLevels(high int16, low int16) {
	d.SetInterruptLevelsHysteresis(high, low, (high*95)/100)
}

// SetInterruptLevelsHysteresis sets the interrupt levels with hysteresis
func (d *Device) SetInterruptLevelsHysteresis(high int16, low int16, hysteresis int16) {
	high = high / PIXEL_TEMP_CONVERSION
	if high < -4095 {
		high = -4095
	}
	if high > 4095 {
		high = 4095
	}
	legacy.WriteRegister(d.bus, uint8(d.Address), INTHL, []byte{uint8(high & 0xFF)})
	legacy.WriteRegister(d.bus, uint8(d.Address), INTHL, []byte{uint8((high & 0xFF) >> 4)})

	low = low / PIXEL_TEMP_CONVERSION
	if low < -4095 {
		low = -4095
	}
	if low > 4095 {
		low = 4095
	}
	legacy.WriteRegister(d.bus, uint8(d.Address), INTHL, []byte{uint8(low & 0xFF)})
	legacy.WriteRegister(d.bus, uint8(d.Address), INTHL, []byte{uint8((low & 0xFF) >> 4)})

	hysteresis = hysteresis / PIXEL_TEMP_CONVERSION
	if hysteresis < -4095 {
		hysteresis = -4095
	}
	if hysteresis > 4095 {
		hysteresis = 4095
	}
	legacy.WriteRegister(d.bus, uint8(d.Address), INTHL, []byte{uint8(hysteresis & 0xFF)})
	legacy.WriteRegister(d.bus, uint8(d.Address), INTHL, []byte{uint8((hysteresis & 0xFF) >> 4)})
}

// EnableInterrupt enables the interrupt pin on the device
func (d *Device) EnableInterrupt() {
	d.interruptEnable = 1
	legacy.WriteRegister(d.bus, uint8(d.Address), INTC, []byte{((uint8(d.interruptMode) << 1) | d.interruptEnable) & 0x03})
}

// DisableInterrupt disables the interrupt pin on the device
func (d *Device) DisableInterrupt() {
	d.interruptEnable = 0
	legacy.WriteRegister(d.bus, uint8(d.Address), INTC, []byte{((uint8(d.interruptMode) << 1) | d.interruptEnable) & 0x03})
}

// SetInterruptMode sets the interrupt mode
func (d *Device) SetInterruptMode(mode InterruptMode) {
	d.interruptMode = mode
	legacy.WriteRegister(d.bus, uint8(d.Address), INTC, []byte{((uint8(d.interruptMode) << 1) | d.interruptEnable) & 0x03})
}

// GetInterrupt reads the state of the triggered interrupts
func (d *Device) GetInterrupt() []uint8 {
	data := make([]uint8, 8)
	legacy.ReadRegister(d.bus, uint8(d.Address), INT_OFFSET, data)
	return data
}

// ClearInterrupt clears any triggered interrupts
func (d *Device) ClearInterrupt() {
	d.SetReset(FLAG_RESET)
}

// ReadThermistor reads the onboard thermistor
func (d *Device) ReadThermistor() int16 {
	data := make([]uint8, 2)
	legacy.ReadRegister(d.bus, uint8(d.Address), TTHL, data)
	return (int16((uint16(data[1])<<8)|uint16(data[0])) * THERMISTOR_CONVERSION) / 10
}
