// Package mcp23017 implements a driver for the MCP23017
// I2C port expander chip. See https://www.microchip.com/wwwproducts/en/MCP23017
// for details of the interface.
//
// It also provides a way of joining several such devices into one logical
// device (see the Devices type).
package mcp23017

import (
	"errors"
	"machine"
)

type i2c interface {
	Tx(addr uint16, w, r []byte) error
}

type Bank bool

const (
	// hwAddressFixed holds the bits of the hardware address
	// that are fixed by the chip. Bits 0-3 (those in hwAddressMask)
	// are user-defined by the A0-A2 pins on the chip.
	hwAddress = uint8(0b010_0000)
	// hwAddressMask holds the bits that are significant in hwAddress.
	hwAddressMask = uint8(0b111_1000)

	BankA = false
	BankB = true
)

type register uint8

const (
	// The following registers all refer to port A (except
	// rIOCON with is port-agnostic).
	// ORing them with portB makes them refer to port B.
	rIODIR        = register(0x00) // I/O direction. 0=output; 1=input.
	rIOPOL        = register(0x02) // Invert input values. 0=normal; 1=inverted.
	rGPINTEN      = register(0x04)
	rDEFVAL       = register(0x06)
	rINTCON       = register(0x08)
	rIOCON        = register(0x0A)
	rGPPU         = register(0x0C) // Pull up; 0=no pull-up; 1=pull-up.
	rINTF         = register(0x0E)
	rINTCAP       = register(0x10)
	rGPIO         = register(0x12) // GPIO pin values.
	rOLAT         = register(0x14)
	registerCount = 0x16

	portB = register(0x1)

	flagIOCON_Mirror = byte(0x1 << 6)
)

// PinCount is the number of GPIO pins available on the chip.
const PinCount = 16

// PinMode represents a possible I/O mode for a pin.
// The zero value represents the default value
// after the chip is reset (input).
type PinMode uint8

const (
	// Input configures a pin as an input.
	Input = PinMode(0)
	// Output configures a pin as an output.
	Output = PinMode(1)

	// Direction is the bit mask of the pin mode representing
	// the I/O direction.
	Direction = PinMode(1)

	// Pullup can be bitwise-or'd with Input
	// to cause the pull-up resistor on the pin to
	// be enabled.
	Pullup = PinMode(2)

	// Invert can be bitwise-or'd with Input to
	// cause the pin value to reflect the inverted
	// value on the pin.
	Invert = PinMode(4)
)

// ErrInvalidHWAddress is returned when the hardware address
// of the device is not valid (only some bits can be set by the
// address pins).
var ErrInvalidHWAddress = errors.New("invalid hardware address")

// New returns a new MCP23017 device at the given I2C address
// on the given bus.
// It returns ErrInvalidHWAddress if the address isn't possible for the device.
//
// By default all pins are configured as inputs.
func NewI2C(bus i2c, address uint8) (*Device, error) {
	if address&hwAddressMask != hwAddress {
		return nil, ErrInvalidHWAddress
	}
	d := &Device{
		bus:  bus,
		addr: address,
	}
	pins, err := d.GetPins()
	if err != nil {
		return nil, errors.New("cannot initialize mcp23017 device at " + hex(address) + ": " + err.Error())
	}
	d.pins = pins
	return d, nil
}

func hex(x uint8) string {
	digits := "0123456789abcdef"
	return "0x" + digits[x>>4:x>>4+1] + digits[x&0xf:x&0xf+1]
}

// Device represents an MCP23017 device.
type Device struct {
	// TODO would it be good to have a mutex here so that independent goroutines
	// could change pins without needing to do the locking themselves?

	// bus holds the reference the I2C bus that the device lives on.
	// It's an interface so that we can write tests for it.
	bus  i2c
	addr uint8
	// pins caches the most recent pin values that have been set.
	// This enables us to change individual pin values without
	// doing a read followed by a write.
	pins Pins

	intrptEn   Pins // interrupt pin enabled
	intrptConf Pins // interrupt pin cofig, cleared is toggle
	intrptDefv Pins // interrupt pin default value
	intrptFns  [PinCount]func(Pin)
	intrptF    Pins // interrupt fired flag
	intrptCap  Pins // interrupt captured value flag
}

// GetPins reads all 16 pins from ports A and B.
func (d *Device) GetPins() (Pins, error) {
	return d.readRegisterAB(rGPIO)
}

// SetPins sets all the pins for which mask is high
// to their respective values in pins.
//
// That is, it does the equivalent of:
//
//	for i := 0; i < PinCount; i++ {
//		if mask.Get(i) {
//			d.Pin(i).Set(pins.Get(i))
//		}
//	}
func (d *Device) SetPins(pins, mask Pins) error {
	if mask == 0 {
		return nil
	}
	newPins := (d.pins &^ mask) | (pins & mask)
	if newPins == d.pins {
		return nil
	}
	err := d.writeRegisterAB(rGPIO, newPins)
	if err != nil {
		return err
	}
	d.pins = newPins
	return nil
}

// TogglePins inverts the values on all pins for
// which mask is high.
func (d *Device) TogglePins(mask Pins) error {
	if mask == 0 {
		return nil
	}
	return d.SetPins(^d.pins, mask)
}

// Pin returns a Pin representing the given pin number (from 0 to 15).
// Pin numbers from 0 to 7 represent port A pins 0 to 7.
// Pin numbers from 8 to 15 represent port B pins 0 to 7.
func (d *Device) Pin(pin int) Pin {
	if pin < 0 || pin >= PinCount {
		panic("pin out of range")
	}
	var mask Pins
	mask.High(pin)
	return Pin{
		dev:  d,
		mask: mask,
		pin:  uint8(pin),
	}
}

// SetAllModes sets the mode of all the pins in a single operation.
// If len(modes) is less than PinCount, all remaining pins
// will be set fo modes[len(modes)-1], or PinMode(0) if
// modes is empty.
//
// If len(modes) is greater than PinCount, the excess entries
// will be ignored.
func (d *Device) SetModes(modes []PinMode) error {
	defaultMode := PinMode(0)
	if len(modes) > 0 {
		defaultMode = modes[len(modes)-1]
	}
	var dir, pullup, invert Pins
	for i := 0; i < PinCount; i++ {
		mode := defaultMode
		if i < len(modes) {
			mode = modes[i]
		}
		if mode&Direction == Input {
			dir.High(i)
		}
		if mode&Pullup != 0 {
			pullup.High(i)
		}
		if mode&Invert != 0 {
			invert.High(i)
		}
	}
	if err := d.writeRegisterAB(rIODIR, dir); err != nil {
		return err
	}
	if err := d.writeRegisterAB(rGPPU, pullup); err != nil {
		return err
	}
	if err := d.writeRegisterAB(rIOPOL, invert); err != nil {
		return err
	}
	return nil
}

// GetModes reads the modes of all the pins into modes.
// It's OK if len(modes) is not PinCount - excess entries
// will be left unset.
func (d *Device) GetModes(modes []PinMode) error {
	dir, err := d.readRegisterAB(rIODIR)
	if err != nil {
		return err
	}
	pullup, err := d.readRegisterAB(rGPPU)
	if err != nil {
		return err
	}
	invert, err := d.readRegisterAB(rIOPOL)
	if err != nil {
		return err
	}
	if len(modes) > PinCount {
		modes = modes[:PinCount]
	}
	for i := range modes {
		mode := Output
		if dir.Get(i) {
			mode = Input
		}
		if pullup.Get(i) {
			mode |= Pullup
		}
		if invert.Get(i) {
			mode |= Invert
		}
		modes[i] = mode
	}
	return nil
}

// SetInterruptMirror enables interrupt mirroring between
// the two banks, allowing you to use a single pin to
// handle interrupt for both ports.
func (d *Device) SetInterruptMirror(status bool) error {
	var buf [2]byte
	err := d.bus.Tx(uint16(d.addr), []byte{uint8(rIOCON)}, buf[:])
	if err != nil {
		return err
	}

	if status {
		buf[0], buf[1] = (buf[0] | flagIOCON_Mirror), (buf[1] | flagIOCON_Mirror)
	} else {
		buf[0], buf[1] = (buf[0] &^ flagIOCON_Mirror), (buf[1] &^ flagIOCON_Mirror)
	}

	return d.bus.Tx(uint16(d.addr), []byte{uint8(rIOCON)}, buf[:])
}

func (d *Device) GetInterruptsEnabled() (Pins, error) {
	return d.readRegisterAB(rGPINTEN)
}

func (d *Device) SetInterruptsEnabled(intEn Pins, mask Pins) error {
	if mask == 0 {
		return nil
	}

	newIntEns := (d.intrptEn &^ mask) | (intEn & mask)
	if newIntEns == d.intrptEn {
		return nil
	}
	err := d.writeRegisterAB(rGPINTEN, newIntEns)
	if err != nil {
		return err
	}
	d.intrptEn = newIntEns

	return nil
}

// SetInterrupts sets all the interrupts covered by mask to call
// the callback function on the given pin change condition. The callback
// function is not actually called during the interrupt handler, but
// from a Go routine.
func (d *Device) SetInterrupts(change machine.PinChange, callback func(Pin), mask Pins) error {
	if mask == 0 {
		return nil
	}

	intCon, err := d.readRegisterAB(rINTCON)
	if err != nil {
		return err
	}

	defVal, err := d.readRegisterAB(rDEFVAL)
	if err != nil {
		return err
	}

	switch change {
	case machine.PinRising:
		defVal &= ^mask // rDEFVAL = 0
		intCon |= mask  // rINTCON = 1
	case machine.PinFalling:
		defVal |= mask // rDEFVAL = 1
		intCon |= mask // rINTCON = 1
	case machine.PinToggle:
		intCon &= ^mask // rINTCON = 0
	default:
		return nil
	}

	if change != machine.PinToggle {
		err = d.writeRegisterAB(rDEFVAL, defVal)
		if err != nil {
			return err
		}
		d.intrptDefv = defVal
	}

	err = d.writeRegisterAB(rINTCON, intCon)
	if err != nil {
		return err
	}
	d.intrptConf = intCon
	for i := range PinCount {
		if ((1 << i) & mask) == 0 {
			continue
		}
		d.intrptFns[i] = callback
	}

	return nil
}

// GetInterruptFn returns an interrupt handler for registering
// on the machine.Pin connected to IntA or IntB. This function
// starts a Go Routine.
func (d *Device) GetInterruptFn(bank Bank) func(machine.Pin) {
	ch := make(chan Bank, 4) // what to do about length here?

	go func() {
		// dispatch
		println("go routine here ", bank)
		for _ = range ch {
			var buf [4]byte
			// We'll read INTF and INTCAP in one go
			if err := d.bus.Tx(uint16(d.addr), []byte{uint8(rINTF)}, buf[:]); err != nil {
				// we got an error, not sure what to do there
				continue
			}

			// We should probably store these in one 32 bit value and
			// read/write as a volatile
			intf := Pins(buf[0]) | (Pins(buf[1]) << 8)
			intcap := Pins(buf[1]) | (Pins(buf[2]) << 8)

			newIntCap := (d.intrptCap &^ intf) | (intcap & intf)
			d.intrptCap = newIntCap

			for i, fn := range d.intrptFns {
				if fn == nil {
					continue
				}
				mask := uint16(1) << i
				if (mask & uint16(intf)) == 0 {
					continue
				}
				fn(Pin{
					mask: Pins(mask),
					pin:  uint8(i),
					dev:  d,
				})
			}
		}
	}()

	return func(p machine.Pin) {
		select {
		case ch <- bank:
		default:
			println("blocked")
		}
	}
}

func (d *Device) writeRegisterAB(r register, val Pins) error {
	w := []byte{uint8(r &^ portB), uint8(val), uint8(val >> 8)}
	return d.bus.Tx(uint16(d.addr), w, nil)
}

func (d *Device) readRegisterAB(r register) (Pins, error) {
	var buf [2]byte
	if err := d.bus.Tx(uint16(d.addr), []byte{uint8(r)}, buf[:]); err != nil {
		return Pins(0), err
	}

	return Pins(buf[0]) | (Pins(buf[1]) << 8), nil
}

// Pin represents a single GPIO pin on the device.
type Pin struct {
	// mask holds the mask of the pin.
	mask Pins
	// pin holds the actual pin number.
	pin uint8
	dev *Device
}

// Set sets the pin to the given value.
func (p Pin) Set(value bool) error {
	// TODO currently this always writes both registers when
	// technically it only needs to write one. We could potentially
	// optimize that.
	if value {
		return p.dev.SetPins(^Pins(0), p.mask)
	} else {
		return p.dev.SetPins(0, p.mask)
	}
}

// High is short for p.Set(true).
func (p Pin) High() error {
	return p.Set(true)
}

// High is short for p.Set(false).
func (p Pin) Low() error {
	return p.Set(false)
}

// Toggle inverts the value output on the pin.
func (p Pin) Toggle() error {
	return p.dev.TogglePins(p.mask)
}

// Get returns the current value of the given pin.
func (p Pin) Get() bool {
	if (p.mask&p.dev.intrptEn) != 0 && (p.mask&p.dev.intrptF) != 0 {
		// clear the captured flag, not needed for toggle pins
		if (p.mask & p.dev.intrptConf) != 0 {
			p.dev.intrptF = (^p.mask & p.dev.intrptF)
		}
		return (p.mask & p.dev.intrptCap) > 0
	}

	// TODO this reads 2 registers when we could read just one.
	pins, err := p.dev.GetPins()
	if err != nil {
		panic(err.Error())
	}
	return pins&p.mask != 0
}

// SetMode configures the pin to the given mode.
func (p Pin) SetMode(mode PinMode) error {
	// We could use a more efficient single-register
	// read/write pattern but setting pin modes isn't an
	// operation that's likely to need to be efficient, so
	// use less code and use Get/SetModes directly.
	modes := make([]PinMode, PinCount)
	if err := p.dev.GetModes(modes); err != nil {
		return err
	}
	modes[p.pin] = mode
	err := p.dev.SetModes(modes)
	if err != nil {
		return err
	}
	err = p.dev.GetModes(modes)
	return nil
}

// SetInterrupt configures the pin to the given mode.
func (p Pin) SetInterrupt(change machine.PinChange, callback func(Pin)) error {
	return p.dev.SetInterrupts(change, callback, p.mask)
}

// SetInterruptEnabled enable interrupts on this pin
func (p Pin) SetInterruptEnabled() error {
	return p.dev.SetInterruptsEnabled(p.mask, p.mask)
}

// GetMode returns the mode of the pin.
func (p Pin) GetMode() (PinMode, error) {
	modes := make([]PinMode, PinCount)
	if err := p.dev.GetModes(modes); err != nil {
		return 0, err
	}
	return modes[p.pin], nil
}

// Configure is a stub NoOp function to match the interface
// for pins expected by some other drivers. The user
// must configure pins manually as required by drivers that
// use the pins.
func (p Pin) Configure(mCfg machine.PinConfig) {
}

// Pins represents a bitmask of pin values.
// Port A values are in bits 0-8 (numbered from least significant bit)
// Port B values are in bits 9-15.
type Pins uint16

// Set sets the value for the given pin.
func (p *Pins) Set(pin int, value bool) {
	if value {
		p.High(pin)
	} else {
		p.Low(pin)
	}
}

// Get returns the value for the given pin.
func (p Pins) Get(pin int) bool {
	return (p & pinMask(pin)) != 0
}

// High is short for p.Set(pin, true).
func (p *Pins) High(pin int) {
	*p |= pinMask(pin)
}

// Low is short for p.Set(pin, false).
func (p *Pins) Low(pin int) {
	*p &^= pinMask(pin)
}

// Toggle inverts the value of the given pin.
func (p *Pins) Toggle(pin int) {
	*p ^= pinMask(pin)
}

func pinMask(pin int) Pins {
	return 1 << pin
}
