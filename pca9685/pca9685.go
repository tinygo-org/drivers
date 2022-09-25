package pca9685

import (
	"encoding/binary"
	"time"

	"tinygo.org/x/drivers"
)

const (
	// Internal oscillator frequency.
	oscclock = 25000000
	// Max value PWM value (on or off) can take.
	maxtop       = 1<<12 - 1
	milliseconds = 1_000_000 // units in nanoseconds
)

// Dev is a handle to the PCA9685 device given an address
// (usually 0x47) and an i2c bus.
type Dev struct {
	addr uint8
	bus  drivers.I2C
	buf  [4]byte
}

type PWMConfig struct {
	Period uint64
}

// New creates a new instance of a PCA9685 device. It performs
// no IO on the i2c bus.
func New(bus drivers.I2C, addr uint8) Dev {
	return Dev{
		bus:  bus,
		addr: addr,
	}
}

// Configure enables autoincrement, sets all PWM signals to logic low (Ground)
// and finally sets the Period.
func (d Dev) Configure(cfg PWMConfig) error {
	err := d.SetAI(true)
	if err != nil {
		return err
	}
	d.SetAll(0)
	d.SetDrive(true)
	return d.SetPeriod(cfg.Period)
}

// SetPeriod updates the period of this PWM integrated circuit in nanoseconds.
// To set a particular frequency, use the following formula:
//
//	period = 1e9 / frequency
//
// In the equation above frequency is in Hertz.
//
// If you use a period of 0, a period that works well for LEDs will be picked.
//
// PCA9685 accepts frequencies inbetween [40..1000 Hz],
// or expressed as a period [1..25ms].
func (d Dev) SetPeriod(period uint64) error {
	const div = maxtop + 1
	if period == 0 {
		period = 1 * milliseconds
	}
	if period > 25*milliseconds || period < 1*milliseconds {
		return ErrBadPeriod
	}
	// Correct for overshoot in provided frequency: https://github.com/adafruit/Adafruit-PWM-Servo-Driver-Library/issues/11
	// Note: 0.96 was empirically determined to be closer. Should follow up to understand what is happening here.
	freq := 96 * 1e9 / (100 * period)
	prescale := byte(oscclock/(div*freq) - 1)
	err := d.Sleep(true) // Enable sleep to write to PRESCALE register
	if err != nil {
		return err
	}
	d.buf[0] = prescale
	err = d.writeReg(PRESCALE, d.buf[:1])
	if err != nil {
		return err
	}
	return d.Sleep(false)
}

// Top returns max value PWM can take.
func (d Dev) Top() uint32 {
	return maxtop
}

// Set sets the `on` value of a PWM channel in the range [0..15].
// Max value `on` can take is 4095.
// Example:
//
//	d.Set(1, d.Top()/4)
//
// sets the dutycycle of second (LED1) channel to 25%.
func (d Dev) Set(channel uint8, on uint32) {
	if on > maxtop {
		panic("pca9685: value must be in range 0..4095")
	}
	d.SetPhased(channel, on, 0)
}

// SetAll sets all PWM signals to a ON value. Equivalent of calling
//
//	Dev.Set(pca9685.ALLLED, value)
func (d Dev) SetAll(on uint32) {
	d.Set(ALLLED, on)
}

// IsConnected returns error if read fails or if
// driver suspects device is not connected.
func (d Dev) IsConnected() error {
	// Set data to the NOT of default MODE1 contents.
	// If read is succesful then data will be modified
	const notdefaultMODE1 = ^defaultMODE1Value
	d.buf[0] = notdefaultMODE1

	err := d.readReg(MODE1, d.buf[:1])
	if err != nil {
		return err
	} else if d.buf[0] == notdefaultMODE1 {
		return ErrInvalidMode1
	}
	return nil
}

// SetAI enables or disables autoincrement feature on device. Useful for
// writing to many consecutive registers in one shot.
func (d Dev) SetAI(ai bool) error {
	err := d.readReg(MODE1, d.buf[:1])
	if err != nil {
		return err
	}
	if ai {
		d.buf[0] |= AI
	} else {
		d.buf[0] &^= AI
	}
	err = d.writeReg(MODE1, d.buf[:1])
	return err
}

// SetDrive configures PWM output connection in MODE2 register.
//
//	false: The 16 LEDn outputs are configured with an open-drain structure.
//	true: The 16 LEDn outputs are configured with a totem pole structure.
func (d Dev) SetDrive(outdrv bool) error {
	err := d.readReg(MODE2, d.buf[:1])
	if err != nil {
		return err
	}
	if outdrv {
		d.buf[0] |= OUTDRV
	} else {
		d.buf[0] &^= OUTDRV
	}
	return d.writeReg(MODE2, d.buf[:1])
}

// Sleep sets/unsets SLEEP bit in MODE1.
//
//	if sleepEnabled
//	  Stops PWM. Allows writing to PRE_SCALE register.
//	else
//	  wakes PCA9685. Resumes PWM.
func (d Dev) Sleep(sleepEnabled bool) error {
	err := d.readReg(MODE1, d.buf[:1])
	if err != nil {
		return err
	}
	if sleepEnabled {
		d.buf[0] |= SLEEP
		return d.writeReg(MODE1, d.buf[:1])
	}
	d.buf[0] &^= SLEEP
	err = d.writeReg(MODE1, d.buf[:1])
	// It takes 500μs max for the oscillator to be up and running once SLEEP bit
	// has been set to logic 0. Timings on LEDn outputs are not guaranteed if PWM
	// control registers are accessed within the 500μs window.
	// There is no start-up delay required when using the EXTCLK pin as the PWM clock.
	time.Sleep(1000 * time.Microsecond) // Requested by datasheet.
	return err
}

// SetInverting inverts ALL PCA9685 PWMs. The channel argument merely implements PWM interface.
//
// Without inverting, a 25% duty cycle would mean the output is high for 25% of
// the time and low for the rest. Inverting flips the output as if a NOT gate
// was placed at the output, meaning that the output would be 25% low and 75%
// high with a duty cycle of 25%.
func (d Dev) SetInverting(_ uint8, inverting bool) error {
	err := d.readReg(MODE2, d.buf[:1])
	if err != nil {
		return err
	}
	if inverting {
		d.buf[0] |= INVRT
	} else {
		d.buf[0] &^= INVRT
	}
	return d.writeReg(MODE2, d.buf[:1])
}

// SetPhased sets PWM on and off mark.
// The ON time, which is programmable, will be the time the LED output
// will be asserted and the OFF time, which is also programmable, will be
// the time when the LED output will be negated.
// In this way, the phase shift becomes completely programmable.
// The resolution for the phase shift is 1⁄4096 of the target frequency.
func (d Dev) SetPhased(channel uint8, on, off uint32) {
	binary.LittleEndian.PutUint16(d.buf[:2], uint16(on)&maxtop)
	binary.LittleEndian.PutUint16(d.buf[2:4], uint16(off)&maxtop)
	onLReg, _, _, _ := LED(channel)
	d.writeReg(onLReg, d.buf[:4])
}
