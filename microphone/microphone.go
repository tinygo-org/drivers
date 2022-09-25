// Package microphone implements a driver for a PDM microphone.
// For example, the Adafruit PDM MEMS breakout board (https://www.adafruit.com/product/3492)
//
// Datasheet: https://cdn-learn.adafruit.com/assets/assets/000/049/977/original/MP34DT01-M.pdf
package microphone // import "tinygo.org/x/drivers/microphone"

import (
	"machine"
	"math"
)

const (
	defaultSampleRate        = 22000
	quantizeSteps            = 64
	msForSPLSample           = 50
	defaultSampleCountForSPL = (defaultSampleRate / 1000) * msForSPLSample
	defaultGain              = 9.0
	defaultRefLevel          = 0.00002
)

// Device wraps an I2S connection to a PDM microphone device.
type Device struct {
	bus machine.I2S

	// data buffer used for SPL sound pressure level samples
	data []int32

	// buf buffer used for sinc filter
	buf []uint32

	// SampleCountForSPL is number of samples aka size of data buffer to be used
	// for sound pressure level measurement.
	// Once Configure() is called, changing this value has no effect.
	SampleCountForSPL int

	// Gain setting used to calculate sound pressure level
	Gain float64

	// ReferenceLevel setting used to calculate sound pressure level.
	ReferenceLevel float64
}

// New creates a new microphone connection. The I2S bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus machine.I2S) Device {
	return Device{
		bus:               bus,
		SampleCountForSPL: defaultSampleCountForSPL,
		Gain:              defaultGain,
		ReferenceLevel:    defaultRefLevel,
	}
}

// Configure the microphone.
func (d *Device) Configure() {
	d.data = make([]int32, d.SampleCountForSPL)
	d.buf = make([]uint32, (quantizeSteps / 16))
}

// Read the raw microphone data.
func (d *Device) Read(r []int32) (int, error) {
	count := len(r)

	// get the next group of samples
	machine.I2S0.Read(d.buf)

	if len(r) > len(d.buf) {
		count = len(d.buf)
	}
	for i := 0; i < count; i++ {
		r[i] = int32(d.buf[i])
	}

	return count, nil
}

// ReadWithFilter reads the microphone and filters the buffer using the sinc filter.
func (d *Device) ReadWithFilter(r []int32) (int, error) {
	// read/filter the samples
	var sum uint16
	for i := 0; i < len(r); i++ {

		// get the next group of samples
		machine.I2S0.Read(d.buf)

		// filter
		sum = applySincFilter(d.buf)

		// adjust to 10 bit value
		s := int32(sum >> 6)

		// make it close to 0-offset signed
		s -= 512

		r[i] = s
	}

	return len(r), nil
}

// GetSoundPressure returns the sound pressure in milli-decibels.
func (d *Device) GetSoundPressure() (int32, int32) {
	// read/filter the samples
	d.ReadWithFilter(d.data)

	// remove offset
	var avg int32
	for i := 0; i < len(d.data); i++ {
		avg += d.data[i]
	}
	avg /= int32(len(d.data))

	for i := 0; i < len(d.data); i++ {
		d.data[i] -= avg
	}

	// get max value
	var maxval int32
	for i := 0; i < len(d.data); i++ {
		v := d.data[i]
		if v < 0 {
			v = -v
		}
		if maxval < v {
			maxval = v
		}
	}

	// calculate SPL
	spl := float64(maxval) / 1023.0 * d.Gain
	spl = 20 * math.Log10(spl/d.ReferenceLevel)

	return int32(spl * 1000), maxval
}

// sinc filter for 44 khz with 64 samples
// each value matches the corresponding bit in the 8-bit value
// for that sample.
//
// For more information: https://en.wikipedia.org/wiki/Sinc_filter
var sincfilter = [quantizeSteps]uint16{
	0, 2, 9, 21, 39, 63, 94, 132,
	179, 236, 302, 379, 467, 565, 674, 792,
	920, 1055, 1196, 1341, 1487, 1633, 1776, 1913,
	2042, 2159, 2263, 2352, 2422, 2474, 2506, 2516,
	2506, 2474, 2422, 2352, 2263, 2159, 2042, 1913,
	1776, 1633, 1487, 1341, 1196, 1055, 920, 792,
	674, 565, 467, 379, 302, 236, 179, 132,
	94, 63, 39, 21, 9, 2, 0, 0,
}

// applySincFilter uses the sinc filter to process a single set of sample values.
func applySincFilter(samples []uint32) (result uint16) {
	var sample uint16
	pos := 0
	for j := 0; j < len(samples); j++ {
		// takes only the low order 16-bits
		sample = uint16(samples[j] & 0xffff)
		for i := 0; i < 16; i++ {
			if (sample & 0x1) > 0 {
				result += sincfilter[pos]
				pos++
			}
			sample >>= 1
		}
	}

	return
}
