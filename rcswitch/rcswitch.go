// Package rcswitch provides a library to control 433/315MHz devices like power outlet sockets.
package rcswitch // import "tinygo.org/x/drivers/rcswitch"

import (
	"errors"
	"fmt"
	"machine"
	"strconv"
	"strings"
	"time"
)

type waveform struct {
	high, low int // Number of high pulses, followed by number of low pulses.
}

type protocol struct {
	pulseLen                 time.Duration
	syncBit, zeroBit, oneBit waveform
	inverted                 bool
}

// protocols as specified in from https://github.com/sui77/rc-switch
var protocols = []protocol{
	// protocol 1
	{pulseLen: 350, syncBit: waveform{1, 31}, zeroBit: waveform{1, 3}, oneBit: waveform{3, 1}},
	// protocol 2
	{pulseLen: 650, syncBit: waveform{1, 10}, zeroBit: waveform{1, 2}, oneBit: waveform{2, 1}},
	// protocol 3
	{pulseLen: 100, syncBit: waveform{30, 71}, zeroBit: waveform{4, 11}, oneBit: waveform{9, 6}},
	// protocol 4
	{pulseLen: 380, syncBit: waveform{1, 6}, zeroBit: waveform{1, 3}, oneBit: waveform{3, 1}},
	// protocol 5
	{pulseLen: 500, syncBit: waveform{6, 14}, zeroBit: waveform{1, 2}, oneBit: waveform{2, 1}},
	// protocol 6 (HT6P20B)
	{pulseLen: 450, syncBit: waveform{23, 1}, zeroBit: waveform{1, 2}, oneBit: waveform{2, 1}, inverted: true},
	// protocol 7 (HS2303-PT, i. e. used in AUKEY Remote)
	{pulseLen: 150, syncBit: waveform{2, 62}, zeroBit: waveform{1, 6}, oneBit: waveform{6, 1}},
	// protocol 8 Conrad RS-200 RX
	{pulseLen: 200, syncBit: waveform{3, 130}, zeroBit: waveform{7, 16}, oneBit: waveform{3, 16}},
	// protocol 9 Conrad RS-200 TX
	{pulseLen: 200, syncBit: waveform{130, 7}, zeroBit: waveform{16, 7}, oneBit: waveform{16, 3}},
	// protocol 10 (1ByOne Doorbell)
	{pulseLen: 365, syncBit: waveform{18, 1}, zeroBit: waveform{3, 1}, oneBit: waveform{1, 3}, inverted: true},
	// protocol 11 (HT12E)
	{pulseLen: 270, syncBit: waveform{36, 1}, zeroBit: waveform{1, 2}, oneBit: waveform{2, 1}, inverted: true},
	// protocol 12 (SM5212)
	{pulseLen: 320, syncBit: waveform{36, 1}, zeroBit: waveform{1, 2}, oneBit: waveform{2, 1}, inverted: true},
}

// Config is configuration for a switch
// Family is only used for Type C. In the most common case family is unused.
// Type A (most common): family: "", group: binary string (e.g. "11011"), device: binary string (e.g, "10000").
// Type B: family: "", group: string 1-4 (e.g. "1"), device: string 1-4 (e.g, "2").
// Type C: family: string a-f (e.g. "b"), group: string 1-4 (e.g. "1"), device: string 1-4 (e.g, "2").
// Type D: family: "", group: string a-d (e.g. "a"), device: string 1-3 (e.g, "2").
type Config struct {
	Family, Group, Device string
	Repeat, Protocol      int
}

// Device represents a RC switch
type Device struct {
	scratch [12]byte
	pin     machine.Pin
	c       Config
}

// New returns a Device
func New(pin machine.Pin) Device {
	pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d := Device{
		pin: pin,
	}

	return d
}

// Configure sets a Device configuration
func (d *Device) Configure(c Config) error {
	if c.Repeat <= 0 {
		c.Repeat = 10
	}

	// "protocol notation", protocol 1 is protocols[0] and so on
	if c.Protocol <= 0 {
		c.Protocol = 1
	}
	if c.Protocol > len(protocols) {
		return fmt.Errorf("Protocol %d not supported valid: 1-%d", c.Protocol, len(protocols))
	}

	d.c = c
	return nil
}

// On turns a switch on
func (d *Device) On() error {
	if err := d.getCodeWord(true); err != nil {
		return err
	}
	d.sendTriState()
	return nil
}

// Off turns a switch off
func (d *Device) Off() error {
	if err := d.getCodeWord(false); err != nil {
		return err
	}
	d.sendTriState()
	return nil
}

func (d *Device) sendTriState() {
	d.triStateToBinary()
	d.send()
}

func (d *Device) send() {
	protocol := protocols[d.c.Protocol-1]
	dt := protocol.pulseLen * time.Microsecond
	zero := protocol.zeroBit
	one := protocol.oneBit
	sync := protocol.syncBit

	high, low := d.pin.High, d.pin.Low
	if protocol.inverted {
		high, low = low, high
	}

	for i := 0; i < d.c.Repeat; i++ {
		for _, w := range d.scratch {
			for b := 1; b >= 0; b-- {
				var h, l int
				switch (w >> b) & 0x1 {
				case 0:
					h, l = zero.high, zero.low
				case 1:
					h, l = one.high, one.low
				}
				high()
				time.Sleep(time.Duration(h) * dt)
				low()
				time.Sleep(time.Duration(l) * dt)
			}
		}

		// sync
		high()
		time.Sleep(time.Duration(sync.high) * dt)
		low()
		time.Sleep(time.Duration(sync.low) * dt)
	}

	// disable transmit
	d.pin.Low()
}

func (d *Device) getCodeWord(status bool) error {
	if d.c.Family != "" { // Type C
		return d.getCodeWordC(status)
	}

	if len(d.c.Group) > 1 && len(d.c.Device) > 1 { // Type A
		return d.getCodeWordA(status)
	}

	if len(d.c.Group) == 1 && len(d.c.Device) == 1 { // Type B or D
		// both have an integer device
		dev, err := strconv.Atoi(d.c.Device)
		if err != nil {
			return errors.New("Protocols B/D must have device string convertible to int")
		}
		g, err := strconv.Atoi(d.c.Group)
		if err != nil { // Type B
			return d.getCodeWordB(g, dev, status)
		} else { // Type D
			return d.getCodeWordD(dev, status)
		}
	}

	return errors.New("family, group, device combination not supported")
}

func (d *Device) getCodeWordA(status bool) error {
	if len(d.c.Group) != 5 {
		return errors.New("Group len != 5 encoded as binary (e.g., 11011)")
	}
	if len(d.c.Device) != 5 {
		return errors.New("Device len != 5 encoded as binary (e.g., 10000)")
	}

	for i, b := range d.c.Group + d.c.Device {
		if b == '0' {
			d.scratch[i] = 'F'
		} else {
			d.scratch[i] = '0'
		}
	}

	if status {
		d.scratch[10], d.scratch[11] = '0', 'F'
	} else {
		d.scratch[10], d.scratch[11] = 'F', '0'
	}

	return nil
}

// TODO: This is untested, if you can test it, please send a pull request removing this comment
func (d *Device) getCodeWordB(group, device int, status bool) error {
	if group < 1 || group > 4 || device < 1 || device > 4 {
		return errors.New("Group and device must be between 1 to 4")
	}

	var pos int
	for i := 1; i <= 4; i++ {
		if group == i {
			d.scratch[pos] = '0'
		} else {
			d.scratch[pos] = 'F'
		}
		pos++
	}

	for i := 1; i <= 4; i++ {
		if device == i {
			d.scratch[pos] = '0'
		} else {
			d.scratch[pos] = 'F'
		}
		pos++
	}

	for i := 0; i < 3; i++ {
		d.scratch[pos] = 'F'
		pos++
	}

	if status {
		d.scratch[pos] = 'F'
	} else {
		d.scratch[pos] = '0'
	}

	return nil
}

// TODO: This is untested, if you can test it, please send a pull request removing this comment
func (d *Device) getCodeWordC(status bool) error {
	if len(d.c.Family) != 1 {
		return errors.New("Family has to be a single character")
	}

	f, err := strconv.ParseUint(d.c.Family, 16, 8) // implicetly contains a..f check
	if err != nil {
		return err
	}

	g, err := strconv.Atoi(d.c.Group)
	if err != nil {
		return err
	}
	if g < 1 || g > 4 {
		return errors.New("Group not between 1 and 4")
	}

	dev, err := strconv.Atoi(d.c.Device)
	if err != nil {
		return err
	}
	if dev < 1 || dev > 4 {
		return errors.New("Device not between 1 and 4")
	}

	var pos int

	for i := uint(0); i < 4; i++ {
		if (f & 0x1) == 0x1 {
			d.scratch[pos] = 'F'
		} else {
			d.scratch[pos] = '0'
		}
		pos++
		f >>= 1
	}

	conf := func(i int) {
		iu := uint(i) - 1
		if iu&0x1 == 1 {
			d.scratch[pos] = 'F'
		} else {
			d.scratch[pos] = '0'
		}
		pos++
		if iu&0x2 == 1 {
			d.scratch[pos] = 'F'
		} else {
			d.scratch[pos] = '0'
		}
		pos++
	}

	conf(dev)
	conf(g)

	// status
	d.scratch[pos] = '0'
	pos++
	d.scratch[pos] = 'F'
	pos++
	d.scratch[pos] = 'F'
	pos++

	if status {
		d.scratch[pos] = 'F'
	} else {
		d.scratch[pos] = '0'
	}

	return nil
}

// TODO: This is untested, if you can test it, please send a pull request removing this comment
func (d *Device) getCodeWordD(device int, status bool) error {
	if len(d.c.Group) != 1 {
		return errors.New("Group must be a single string char")
	}

	switch strings.ToLower(d.c.Group) {
	case "a":
		d.scratch[0], d.scratch[1], d.scratch[2], d.scratch[3] = '1', 'F', 'F', 'F'
	case "b":
		d.scratch[0], d.scratch[1], d.scratch[2], d.scratch[3] = 'F', '1', 'F', 'F'
	case "c":
		d.scratch[0], d.scratch[1], d.scratch[2], d.scratch[3] = 'F', 'F', '1', 'F'
	case "d":
		d.scratch[0], d.scratch[1], d.scratch[2], d.scratch[3] = 'F', 'F', 'F', '1'
	default:
		return errors.New("Group has to be in a-d or A-D")
	}

	switch device {
	case 1:
		d.scratch[4], d.scratch[5], d.scratch[6] = '1', 'F', 'F'
	case 2:
		d.scratch[4], d.scratch[5], d.scratch[6] = 'F', '1', 'F'
	case 3:
		d.scratch[4], d.scratch[5], d.scratch[6] = 'F', 'F', '1'
	default:
		return errors.New("Group must be between 1 and 3")
	}

	// unused
	d.scratch[7], d.scratch[8], d.scratch[9] = '0', '0', '0'

	// status
	if status {
		d.scratch[10], d.scratch[11] = '1', '0'
	} else {
		d.scratch[10], d.scratch[11] = '0', '1'
	}

	return nil
}

func (d *Device) triStateToBinary() {
	// both bits are important
	for i, b := range d.scratch {
		switch b {
		case '0':
			d.scratch[i] = 0x0 // 00
		case '1':
			d.scratch[i] = 0x3 // 11
		case 'F':
			d.scratch[i] = 0x1 // 01
		}
	}
}
