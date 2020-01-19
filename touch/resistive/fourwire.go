package resistive

import (
	"machine"

	"tinygo.org/x/drivers/touch"
)

type FourWireConfig struct {

	// Y+ pin, must be capable of analog reads
	YP machine.Pin

	// Y- pin, must be capable of analog reads
	YM machine.Pin

	// X+ pin, must be capable of analog reads
	XP machine.Pin

	// X- pin, must be capable of analog reads
	XM machine.Pin

	// AnalogResolution is the resolution in bits of the ADC used for reading
	AnalogResolution int
}

type FourWireTouchscreen struct {
	yp machine.ADC
	ym machine.ADC
	xp machine.ADC
	xm machine.ADC

	samples   []uint16
	scaleBits int
}

func (res *FourWireTouchscreen) Configure(config *FourWireConfig) error {

	res.yp = machine.ADC{config.YP}
	res.ym = machine.ADC{config.YM}
	res.xp = machine.ADC{config.XP}
	res.xm = machine.ADC{config.XM}

	res.samples = make([]uint16, 2)

	return nil
}

func (res *FourWireTouchscreen) ReadTouchPoint() (p touch.Point) {
	p.X = int(res.ReadX())
	p.Y = int(res.ReadY())
	p.Z = int(res.ReadZ())
	return
}

func (res *FourWireTouchscreen) ReadX() uint16 {
	res.ym.Pin.Configure(machine.PinConfig{machine.PinInput})
	res.ym.Pin.Low()

	res.xp.Pin.Configure(machine.PinConfig{machine.PinOutput})
	res.xp.Pin.High()

	res.xm.Pin.Configure(machine.PinConfig{machine.PinOutput})
	res.xm.Pin.Low()

	res.yp.Configure()

	res.samples[0] = res.yp.Get() >> 2
	res.samples[1] = res.yp.Get() >> 2
	return 1023 - (((res.samples[0] + res.samples[1]) / 2) >> 4)
}

func (res *FourWireTouchscreen) ReadY() uint16 {
	res.xm.Pin.Configure(machine.PinConfig{machine.PinOutput})
	res.xm.Pin.Low()

	res.yp.Pin.Configure(machine.PinConfig{machine.PinOutput})
	res.yp.Pin.High()

	res.ym.Pin.Configure(machine.PinConfig{machine.PinOutput})
	res.ym.Pin.Low()

	res.xp.Configure()

	res.samples[0] = res.xp.Get() >> 2
	res.samples[1] = res.xp.Get() >> 2
	return 1023 - (((res.samples[0] + res.samples[1]) / 2) >> 4)
}

func (res *FourWireTouchscreen) ReadZ() uint16 {
	res.xp.Pin.Configure(machine.PinConfig{machine.PinOutput})
	res.xp.Pin.Low()

	res.ym.Pin.Configure(machine.PinConfig{machine.PinOutput})
	res.ym.Pin.High()

	res.xm.Configure()
	res.yp.Configure()

	z1 := res.xm.Get()
	z2 := res.yp.Get()

	return (1023 - (z2>>6 - z1>>6))
}
