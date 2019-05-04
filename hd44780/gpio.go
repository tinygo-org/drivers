package hd44780

import (
	"errors"

	"machine"
)

type GPIO struct {
	dataPins []machine.GPIO
	e        machine.GPIO
	rw       machine.GPIO
	rs       machine.GPIO

	write func(data byte)
	read  func() byte
}

func newGPIO(dataPins []uint8, e, rs, rw uint8, mode byte) Device {

	pins := make([]machine.GPIO, len(dataPins))
	for i := 0; i < len(dataPins); i++ {
		m := machine.GPIO{Pin: dataPins[i]}
		m.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
		pins[i] = m
	}
	enable := machine.GPIO{e}
	enable.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	registerSelect := machine.GPIO{rs}
	registerSelect.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	readWrite := machine.GPIO{rw}
	readWrite.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	readWrite.Low()

	gpio := GPIO{
		dataPins: pins,
		e:        enable,
		rs:       registerSelect,
		rw:       readWrite,
	}

	if mode == DATA_LENGTH_4BIT {
		gpio.write = gpio.write4BitMode
		gpio.read = gpio.read4BitMode
	} else {
		gpio.write = gpio.write8BitMode
		gpio.read = gpio.read8BitMode
	}

	return Device{
		bus:        &gpio,
		datalength: mode,
	}
}

// SetCommandMode sets command/instruction mode
func (g *GPIO) SetCommandMode(set bool) {
	if set {
		g.rs.Low()
	} else {
		g.rs.High()
	}
}

// Write writes len(data) bytes from data to display driver
func (g *GPIO) Write(data []byte) (n int, err error) {
	g.rw.Low()
	for _, d := range data {
		g.write(d)
		n++
	}
	return n, nil
}

func (g *GPIO) write8BitMode(data byte) {
	g.e.High()
	g.setPins(data)
	g.e.Low()
}

func (g *GPIO) write4BitMode(data byte) {
	g.e.High()
	g.setPins(data >> 4)
	g.e.Low()

	g.e.High()
	g.setPins(data)
	g.e.Low()
}

// Read reads len(data) bytes from display RAM to data starting from RAM address counter position
// Ram address can be changed by writing address in command mode
func (g *GPIO) Read(data []byte) (n int, err error) {
	if len(data) == 0 {
		return 0, errors.New("Length greater than 0 is required")
	}
	g.rw.High()
	g.reconfigureGPIOMode(machine.GPIO_INPUT)
	for i := 0; i < len(data); i++ {
		data[i] = g.read()
		n++
	}
	g.reconfigureGPIOMode(machine.GPIO_OUTPUT)
	return n, nil
}

func (g *GPIO) read4BitMode() byte {
	g.e.High()
	data := (g.pins() << 4 & 0xF0)
	g.e.Low()
	g.e.High()
	data |= (g.pins() & 0x0F)
	g.e.Low()
	return data
}

func (g *GPIO) read8BitMode() byte {
	g.e.High()
	data := g.pins()
	g.e.Low()
	return data
}

func (g *GPIO) reconfigureGPIOMode(mode machine.GPIOMode) {
	for i := 0; i < len(g.dataPins); i++ {
		g.dataPins[i].Configure(machine.GPIOConfig{Mode: mode})
	}
}

// setPins sets high or low state on all data pins depending on data
func (g *GPIO) setPins(data byte) {
	mask := byte(1)
	for i := 0; i < len(g.dataPins); i++ {
		if (data & mask) != 0 {
			g.dataPins[i].High()
		} else {
			g.dataPins[i].Low()
		}
		mask = mask << 1
	}
}

// pins returns current state of data pins. MSB is D7
func (g *GPIO) pins() byte {
	bits := byte(0)
	for i := uint8(0); i < uint8(len(g.dataPins)); i++ {
		if g.dataPins[i].Get() {
			bits |= (1 << i)
		}
	}
	return bits
}
