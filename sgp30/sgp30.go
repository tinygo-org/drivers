// SGP30 VOC sensor.
//
// This sensor is marked obsolete by Sensirion, but is still commonly available.
//
// Datasheet: https://sensirion.com/media/documents/984E0DD5/61644B8B/Sensirion_Gas_Sensors_Datasheet_SGP30.pdf
package sgp30

import (
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

const Address = 0x58

var (
	errInvalidCRC = errors.New("sgp30: invalid CRC")
)

type Device struct {
	bus         drivers.I2C
	commandBuf  [2]byte
	responseBuf [9]byte
	readyTime   time.Time
	co2eq       uint16
	tvoc        uint16
}

type Config struct {
	// Nothing to configure right now.
}

// New returns a new SGP30 driver instance. It does not touch the device yet,
// call Configure to configure this sensor.
func New(bus drivers.I2C) *Device {
	return &Device{
		bus: bus,
		// The sensor has a maximum powerup time of 0.6ms.
		// See table 6 in the datasheet.
		readyTime: time.Now().Add(600 * time.Microsecond),
	}
}

// Connected returns whether something (probably a SGP30) is present on the bus.
func (d *Device) Connected() bool {
	d.waitUntilReady()

	// Request serial ID.
	d.commandBuf = [2]byte{0x36, 0x82}
	err := d.bus.Tx(Address, d.commandBuf[:], nil)
	if err != nil {
		return false
	}

	// Wait 0.5ms as specified in the datasheet.
	time.Sleep(500 * time.Microsecond)

	// Read the serial ID from the sensor.
	err = d.bus.Tx(Address, nil, d.responseBuf[:9])
	if err != nil {
		return false
	}

	// Check whether the CRC matches.
	_, ok1 := readWord(d.responseBuf[:3])
	_, ok2 := readWord(d.responseBuf[3:6])
	_, ok3 := readWord(d.responseBuf[6:9])
	ok := ok1 && ok2 && ok3

	return ok
}

// Wait until a previous command has completed. This may be necessary on
// startup, for example.
func (d *Device) waitUntilReady() {
	now := time.Now()
	delay := d.readyTime.Sub(now)
	if delay > 0 {
		time.Sleep(delay)
	}
}

// Configure starts the measurement process for the SGP30 sensor.
func (d *Device) Configure(config Config) error {
	d.waitUntilReady()

	// Send the sgp30_iaq_init command.
	d.commandBuf = [2]byte{0x20, 0x03}
	err := d.bus.Tx(Address, d.commandBuf[:], nil)

	// The next command will have to wait at least 10ms.
	d.readyTime = time.Now().Add(10 * time.Millisecond)

	return err
}

// Read the current CO₂eq and TVOC values from the sensor.
// This method must be called around once per second per the datasheet as this
// is how the sensor algorithm was calibrated.
func (d *Device) Update(which drivers.Measurement) error {
	d.waitUntilReady()

	// Send sgp30_measure_iaq command.
	d.commandBuf = [2]byte{0x20, 0x08}
	err := d.bus.Tx(Address, d.commandBuf[:], nil)
	if err != nil {
		return err
	}

	// Wait until the response is ready.
	// This can take up to 12ms according to the datasheet.
	time.Sleep(12 * time.Millisecond)

	// Read the response.
	data := d.responseBuf[:6]
	err = d.bus.Tx(Address, nil, data)
	if err != nil {
		return err
	}

	// Decode the response.
	co2eq, ok1 := readWord(data[0:3])
	tvoc, ok2 := readWord(data[3:6])
	if !ok1 || !ok2 {
		return errInvalidCRC
	}
	d.co2eq = co2eq
	d.tvoc = tvoc

	return nil
}

// Returns the CO₂ equivalent value read in the previous measurement.
//
// Warning: this is _not_ an actual CO₂ value. The SGP30 can't actually read
// CO₂. Instead, it's an approximation based on various other gases in the
// environment.
func (d *Device) CO2() uint32 {
	return uint32(d.co2eq)
}

// Returns the total number of VOCs (volatile organic compounds) in parts per
// billion (ppb).
func (d *Device) TVOC() uint32 {
	return uint32(d.tvoc)
}

// Read a single 16-bit word from the sensor and check the CRC. The data
// parameter must be a slice of 3 bytes.
func readWord(data []byte) (value uint16, ok bool) {
	if len(data) != 3 {
		return 0, false
	}
	value = uint16(data[0])<<8 | uint16(data[1])
	crc := uint8(0xff)
	for i := 0; i < 2; i++ {
		crc ^= data[i]
		for b := 0; b < 8; b++ {
			if crc&0x80 != 0 {
				crc = (crc << 1) ^ 0x31
			} else {
				crc <<= 1
			}
		}
	}
	ok = crc == data[2]
	return
}
