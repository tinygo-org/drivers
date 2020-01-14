package resistive

import (
	"machine"

	"tinygo.org/x/drivers/touch"
)

type FourWireTouchscreen struct {
	YP machine.Pin
	YM machine.Pin
	XP machine.Pin
	XM machine.Pin

	yp machine.ADC
	xm machine.ADC
	xp machine.ADC

	samples []uint16
}

func (res *FourWireTouchscreen) Configure() {
	res.yp = machine.ADC{res.YP}
	res.xm = machine.ADC{res.XM}
	res.xp = machine.ADC{res.XP}
	res.samples = make([]uint16, 2)
}

func (res *FourWireTouchscreen) GetTouchPoint() (p touch.Point) {
	p.X = int(res.ReadX())
	p.Y = int(res.ReadY())
	p.Z = int(res.ReadZ())
	return
}

func (res *FourWireTouchscreen) ReadX() uint16 {
	res.YM.Configure(machine.PinConfig{machine.PinInput})
	res.YM.Low()

	res.XP.Configure(machine.PinConfig{machine.PinOutput})
	res.XM.Configure(machine.PinConfig{machine.PinOutput})
	res.XP.High()
	res.XM.Low()

	res.yp.Configure()

	res.samples[0] = res.yp.Get() >> 2
	res.samples[1] = res.yp.Get() >> 2
	return 1023 - (((res.samples[0] + res.samples[1]) / 2) >> 4)
}

func (res *FourWireTouchscreen) ReadY() uint16 {
	res.XM.Configure(machine.PinConfig{machine.PinInput})
	res.XM.Low()

	res.YP.Configure(machine.PinConfig{machine.PinOutput})
	res.YM.Configure(machine.PinConfig{machine.PinOutput})
	res.YP.High()
	res.YM.Low()

	res.xp.Configure()

	res.samples[0] = res.xp.Get() >> 2
	res.samples[1] = res.xp.Get() >> 2
	return 1023 - (((res.samples[0] + res.samples[1]) / 2) >> 4)
}

func (res *FourWireTouchscreen) ReadZ() uint16 {
	res.XP.Configure(machine.PinConfig{machine.PinOutput})
	res.XP.Low()

	res.YM.Configure(machine.PinConfig{machine.PinOutput})
	res.YM.High()

	res.xm.Configure()
	res.yp.Configure()

	z1 := res.xm.Get()
	z2 := res.yp.Get()

	return (1023 - (z2>>6 - z1>>6))
}
