package hd44780

import (
	"errors"
	"time"

	"machine"
)

type WriteByteType int

const (
	FullByte   WriteByteType = 3
	HighNibble WriteByteType = 2
	LowNibble  WriteByteType = 1
)

type GPIO struct {
	dataPins []machine.Pin
	en       machine.Pin
	rw       machine.Pin
	rs       machine.Pin

	write func(data byte, wt WriteByteType)
	read  func() byte
}

func newGPIO(dataPins []machine.Pin, en, rs, rw machine.Pin, mode byte) Device {
	pins := make([]machine.Pin, len(dataPins))
	for i := 0; i < len(dataPins); i++ {
		dataPins[i].Configure(machine.PinConfig{Mode: machine.PinOutput})
		pins[i] = dataPins[i]
	}
	en.Configure(machine.PinConfig{Mode: machine.PinOutput})
	rs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	rw.Configure(machine.PinConfig{Mode: machine.PinOutput})
	rw.Low()

	gpio := GPIO{
		dataPins: pins,
		en:       en,
		rs:       rs,
		rw:       rw,
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

func (g *GPIO) WriteHighNibble(data byte) {
	g.write(data, HighNibble)
}

// SetCommandMode sets command/instruction mode
func (g *GPIO) SetCommandMode(set bool) {
	if set {
		g.rs.Low()
	} else {
		g.rs.High()
	}
}

// WriteOnly is true if you passed rw in as machine.NoPin
func (g *GPIO) WriteOnly() bool {
	return g.rw == machine.NoPin
}

// Write writes len(data) bytes from data to display driver
func (g *GPIO) Write(data []byte) (n int, err error) {
	if !g.WriteOnly() {
		g.rw.Low()
	}
	for _, d := range data {
		g.write(d, FullByte)
		n++
	}
	return n, nil
}

func (g *GPIO) write8BitMode(data byte, _ WriteByteType) {
	g.setPins(data)
	g.pulseEnable()
}

func (g *GPIO) write4BitMode(data byte, wt WriteByteType) {
	if wt&HighNibble != 0 {
		g.setPins(data >> 4)
		g.pulseEnable()
	}

	if wt&LowNibble != 0 {
		g.setPins(data)
		g.pulseEnable()
	}
}

func (g *GPIO) pulseEnable() {
	g.en.Low()
	time.Sleep(time.Microsecond)
	g.en.High()
	time.Sleep(time.Microsecond)
	g.en.Low()
	time.Sleep(100 * time.Microsecond)
}

// Read reads len(data) bytes from display RAM to data starting from RAM address counter position
// Ram address can be changed by writing address in command mode
func (g *GPIO) Read(data []byte) (n int, err error) {
	if len(data) == 0 {
		return 0, errors.New("length greater than 0 is required")
	}
	if g.WriteOnly() {
		return 0, errors.New("Read not supported if RW not wired")
	}
	g.rw.High()
	g.reconfigureGPIOMode(machine.PinInput)
	for i := 0; i < len(data); i++ {
		data[i] = g.read()
		n++
	}
	g.rw.Low()
	g.reconfigureGPIOMode(machine.PinOutput)
	return n, nil
}

func (g *GPIO) read4BitMode() byte {
	g.en.High()
	data := (g.pins() << 4 & 0xF0)
	g.en.Low()

	g.en.High()
	data |= (g.pins() & 0x0F)
	g.en.Low()
	return data
}

func (g *GPIO) read8BitMode() byte {
	g.en.High()
	data := g.pins()
	g.en.Low()
	return data
}

func (g *GPIO) reconfigureGPIOMode(mode machine.PinMode) {
	for i := 0; i < len(g.dataPins); i++ {
		g.dataPins[i].Configure(machine.PinConfig{Mode: mode})
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
