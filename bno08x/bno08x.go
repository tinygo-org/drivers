// Package bno08x provides a TinyGo driver for the Adafruit BNO08x 9-DOF IMU sensors.
//
// This driver implements the CEVA SH-2 protocol over the SHTP transport layer,
// providing access to orientation, motion, and environmental sensors.
//
// Datasheet: https://www.ceva-ip.com/wp-content/uploads/BNO080_085-Datasheet.pdf
package bno08x

import (
	"time"

	"machine"
)

// Device represents a BNO08x sensor device.
type Device struct {
	bus       machine.I2C
	address   uint16
	resetPin  Pin
	readChunk int

	hal  *halI2C
	shtp *shtp
	sh2  *sh2Protocol

	queue      [8]SensorValue
	queueHead  int
	queueTail  int
	queueCount int

	productIDs ProductIDs
	lastReset  bool
}

// Pin represents a GPIO pin (or NoPin if not used).
type Pin interface {
	Configure(config PinConfig)
	High()
	Low()
}

// PinConfig holds pin configuration.
type PinConfig struct {
	Mode PinMode
}

// PinMode represents the pin mode.
type PinMode uint8

const (
	// PinOutput sets the pin as an output.
	PinOutput PinMode = iota
)

// NoPin is a placeholder for when no pin is used.
var NoPin noPin

type noPin struct{}

func (noPin) Configure(PinConfig) {}
func (noPin) High()               {}
func (noPin) Low()                {}

// Config holds configuration options for the device.
type Config struct {
	// Address is the I2C address (default: 0x4A).
	Address uint16

	// ResetPin is the optional hardware reset pin.
	ResetPin Pin

	// ReadChunk is the I2C read chunk size (default: 32 bytes).
	ReadChunk int

	// StartupDelay is the delay after reset (default: 10ms).
	StartupDelay time.Duration
}

const (
	// DefaultAddress is the default I2C address.
	DefaultAddress = 0x4A
)

// New creates a new BNO08x device.
func New(bus *machine.I2C) *Device {
	return &Device{
		bus:       *bus,
		address:   DefaultAddress,
		resetPin:  NoPin,
		readChunk: i2cDefaultChunk,
	}
}

// Configure initializes the sensor and prepares it for use.
func (d *Device) Configure(cfg Config) error {
	if cfg.Address != 0 {
		d.address = cfg.Address
	}
	if cfg.ReadChunk > 0 {
		d.readChunk = cfg.ReadChunk
	}
	if cfg.ResetPin != nil && cfg.ResetPin != NoPin {
		d.resetPin = cfg.ResetPin
		d.resetPin.Configure(PinConfig{Mode: PinOutput})
	}
	if cfg.StartupDelay <= 0 {
		cfg.StartupDelay = 100 * time.Millisecond
	}

	d.hal = newHAL(d)
	d.shtp = newSHTP(d.hal)
	d.sh2 = newSH2Protocol(d)

	d.queueHead = 0
	d.queueTail = 0
	d.queueCount = 0
	d.productIDs = ProductIDs{}
	d.lastReset = false

	if err := d.hal.open(); err != nil {
		return err
	}

	// Now that handlers are registered, perform reset
	// Try hardware reset first if available
	if d.resetPin != nil && d.resetPin != NoPin {
		d.hardwareReset()
		time.Sleep(cfg.StartupDelay)
	} else {
		// No hardware reset pin - try soft reset via I2C raw packet first
		// This is what Adafruit does in hal_open
		if err := d.softResetI2C(); err != nil {
			// If that fails, try soft reset via SHTP protocol
			_ = d.sh2.softReset()
			time.Sleep(50 * time.Millisecond)
		}
	}

	// Wait for reset notification by actively polling
	// The sensor should send reset complete message shortly after reset
	deadline := time.Now().Add(1000 * time.Millisecond)
	pollCount := 0
	for time.Now().Before(deadline) {
		pollCount++
		if err := d.service(); err != nil {
			// Ignore errors during initial polling - sensor might not be ready
			time.Sleep(1 * time.Millisecond)
			continue
		}
		if d.lastReset {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	if !d.lastReset {
		return errTimeout
	}

	// NOTE: We intentionally skip the Initialize command (sh2_initialize)
	// Testing revealed that sending the Initialize command (0xF2 0x00 0x04 0x01...)
	// prevents the BNO08x from sending sensor reports on channel 3.
	// The sensor works correctly without this command after a soft reset.
	// The Arduino library likely works because it does a hardware reset which
	// may put the sensor in a different state, or their initialization sequence
	// differs in a way that doesn't trigger this issue.

	// Request product IDs
	if err := d.sh2.requestProductIDs(); err != nil {
		return err
	}

	// Wait for product IDs with polling delay
	deadline = time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if err := d.service(); err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		if d.productIDs.NumEntries > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if d.productIDs.NumEntries == 0 {
		return errTimeout
	}

	return nil
}

// EnableReport enables a specific sensor report at the given interval.
func (d *Device) EnableReport(id SensorID, intervalUs uint32) error {
	err := d.sh2.enableReport(id, intervalUs)
	if err != nil {
		return err
	}

	// Poll a few times to let the sensor process the command
	// and potentially send acknowledgment
	for i := 0; i < 10; i++ {
		_ = d.service()
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

// GetSensorConfig retrieves the current configuration for a sensor.
func (d *Device) GetSensorConfig(id SensorID) (SensorConfig, error) {
	return d.sh2.getSensorConfig(id)
}

// SetSensorConfig sets the configuration for a sensor.
func (d *Device) SetSensorConfig(id SensorID, config SensorConfig) error {
	return d.sh2.setSensorConfig(id, config)
}

// WasReset returns true if the sensor signaled a reset since the last call.
func (d *Device) WasReset() bool {
	if d.lastReset {
		d.lastReset = false
		return true
	}
	return false
}

// GetSensorEvent retrieves the next available sensor event if present.
func (d *Device) GetSensorEvent() (SensorValue, bool) {
	if d.queueCount == 0 {
		if err := d.service(); err != nil {
			return SensorValue{}, false
		}
		if d.queueCount == 0 {
			return SensorValue{}, false
		}
	}

	value := d.queue[d.queueHead]
	d.queueHead = (d.queueHead + 1) % len(d.queue)
	d.queueCount--

	return value, true
}

// ProductIDs returns the cached product identification information.
func (d *Device) ProductIDs() ProductIDs {
	return d.productIDs
}

// Service processes pending sensor data.
// This is called automatically by GetSensorEvent but can be called manually
// for more control over timing.
func (d *Device) Service() error {
	return d.service()
}

func (d *Device) enqueue(value SensorValue) {
	next := (d.queueTail + 1) % len(d.queue)
	if d.queueCount == len(d.queue) {
		// Queue full, drop oldest
		d.queueHead = (d.queueHead + 1) % len(d.queue)
		d.queueCount--
	}
	d.queue[d.queueTail] = value
	d.queueTail = next
	d.queueCount++
}

func (d *Device) service() error {
	if d.shtp == nil {
		return nil
	}
	for {
		processed, err := d.shtp.poll()
		if err != nil {
			return err
		}
		if !processed {
			break
		}
	}
	return nil
}

func (d *Device) hardwareReset() {
	if d.resetPin == nil || d.resetPin == NoPin {
		return
	}
	d.resetPin.High()
	time.Sleep(10 * time.Millisecond)
	d.resetPin.Low()
	time.Sleep(10 * time.Millisecond)
	d.resetPin.High()
	time.Sleep(10 * time.Millisecond)
}

func (d *Device) softResetI2C() error {
	// Send soft reset packet via I2C as per Adafruit implementation
	// Format: [length_low, length_high, channel, sequence, command]
	// This is: 5 bytes total, channel 1 (executable), command 1 (reset)
	softResetPacket := []byte{5, 0, 1, 0, 1}

	// Try up to 5 times
	var err error
	for attempts := 0; attempts < 5; attempts++ {
		err = d.bus.Tx(d.address, softResetPacket, nil)
		if err == nil {
			// Success - wait for sensor to process reset
			time.Sleep(300 * time.Millisecond)
			return nil
		}
		time.Sleep(30 * time.Millisecond)
	}

	return err
}
