package resistive

import (
	"machine"

	"tinygo.org/x/drivers/touch"
)

// FourWire represents a resistive touchscreen with a four-wire interface as
// described in http://ww1.microchip.com/downloads/en/Appnotes/doc8091.pdf
type FourWire struct {
	yp machine.ADC
	ym machine.ADC
	xp machine.ADC
	xm machine.ADC

	readSamples int
}

// FourWireConfig is passed to the Configure method. All of the pins must be
// specified for this to be a valid configuration. ReadSamples is optional, and
// if not set with default to 2.
type FourWireConfig struct {

	// Y+ pin, must be capable of analog reads
	YP machine.Pin

	// Y- pin, must be capable of analog reads
	YM machine.Pin

	// X+ pin, must be capable of analog reads
	XP machine.Pin

	// X- pin, must be capable of analog reads
	XM machine.Pin

	// If set, each call to ReadTouchPoint() will sample the X, Y, and Z values
	// and average them.  This can help smooth out spurious readings, for example
	// ones that result from the capacitance of a TFT under the touchscreen
	ReadSamples int
}

// Configure should be called once before starting to read the device
func (res *FourWire) Configure(config *FourWireConfig) error {

	res.yp = machine.ADC{Pin: config.YP}
	res.ym = machine.ADC{Pin: config.YM}
	res.xp = machine.ADC{Pin: config.XP}
	res.xm = machine.ADC{Pin: config.XM}

	if config.ReadSamples < 1 {
		res.readSamples = 2
	} else {
		res.readSamples = config.ReadSamples
	}

	return nil
}

// ReadTouchPoint reads a single touch.Point from the device.  If the device
// was configured with ReadSamples > 1, each value will be sampled that many
// times and averaged to smooth over spurious results of the analog reads.
func (res *FourWire) ReadTouchPoint() (p touch.Point) {
	p.X = int(sample(res.ReadX, res.readSamples))
	p.Y = int(sample(res.ReadY, res.readSamples))
	p.Z = int(sample(res.ReadZ, res.readSamples))
	return
}

// sample the results of the provided function and average the results
func sample(fn func() uint16, numSamples int) (v uint16) {
	sum := 0
	for n := 0; n < numSamples; n++ {
		sum += int(fn())
	}
	return uint16(sum / numSamples)
}

// ReadX reads the "raw" X-value on a 16-bit scale without multiple sampling
func (res *FourWire) ReadX() uint16 {
	res.ym.Pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})

	res.xp.Pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	res.xp.Pin.High()

	res.xm.Pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	res.xm.Pin.Low()

	res.yp.Configure(machine.ADCConfig{})

	return 0xFFFF - res.yp.Get()
}

// ReadY reads the "raw" Y-value on a 16-bit scale without multiple sampling
func (res *FourWire) ReadY() uint16 {
	res.xm.Pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})

	res.yp.Pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	res.yp.Pin.High()

	res.ym.Pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	res.ym.Pin.Low()

	res.xp.Configure(machine.ADCConfig{})

	return 0xFFFF - res.xp.Get()
}

// ReadZ reads the "raw" Z-value on a 16-bit scale without multiple sampling
func (res *FourWire) ReadZ() uint16 {
	res.xp.Pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	res.xp.Pin.Low()

	res.ym.Pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	res.ym.Pin.High()

	res.xm.Configure(machine.ADCConfig{})
	res.yp.Configure(machine.ADCConfig{})

	z1 := res.xm.Get()
	z2 := res.yp.Get()

	return 0xFFFF - (z2 - z1)
}
