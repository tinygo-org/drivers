package mcp23017

import "tinygo.org/x/drivers"

// All is a convenience value that represents all pins high (or all mask bits one).
var All = PinSlice{0xffff}

// Devices holds a slice of devices that can be treated as one
// contiguous set of devices. Earlier entries in the slice have
// lower-numbered pins, so index 0 holds pins 0-7, index 1 holds
// pins 8-15, etc.
type Devices []*Device

// NewI2CDevices returns a Devices slice holding the Device values
// for all the given addresses on the given bus.
// When more than one bus is in use, create the slice yourself.
func NewI2CDevices(bus drivers.I2C, addrs ...uint8) (Devices, error) {
	devs := make(Devices, len(addrs))
	for i, addr := range addrs {
		dev, err := NewI2C(bus, addr)
		if err != nil {
			// TODO return a more informative error.
			return nil, err
		}
		devs[i] = dev
	}
	return devs, nil
}

// SetModes sets the pin modes of all the pins on all the devices in devs.
// If there are less entries in modes than there are pins, the
// last entry is replicated to all of them (or PinMode(0) if modes
// is empty).
func (devs Devices) SetModes(modes []PinMode) error {
	var defaultModes []PinMode
	if len(modes) > 0 {
		defaultModes = modes[len(modes)-1:]
	}
	for i, dev := range devs {
		pinStart := i * PinCount
		var devModes []PinMode
		if pinStart < len(modes) {
			devModes = modes[pinStart:]
		} else {
			devModes = defaultModes
		}
		if err := dev.SetModes(devModes); err != nil {
			return err
		}
	}
	return nil
}

// GetModes gets the pin modes from the devices.
// It's OK if modes isn't the same length as all the pins:
// extra entries will be left unchanged.
func (devs Devices) GetModes(modes []PinMode) error {
	for i, dev := range devs {
		pinStart := i * PinCount
		if pinStart >= len(modes) {
			break
		}
		if err := dev.GetModes(modes[pinStart:]); err != nil {
			return err
		}
	}
	return nil
}

// Pin returns the pin for the given number.
func (devs Devices) Pin(pin int) Pin {
	if pin < 0 || pin >= len(devs)*PinCount {
		panic("pin out of range")
	}
	return devs[pin/PinCount].Pin(pin % PinCount)
}

// GetPins returns pin values for all the pins.
func (devs Devices) GetPins(pins PinSlice) error {
	for i, dev := range devs {
		if i >= len(pins) {
			break
		}
		devPins, err := dev.GetPins()
		if err != nil {
			return err
		}
		pins[i] = devPins
	}
	return nil
}

// SetPins sets all the pins for which mask is high
// to their respective values in pins.
//
// That is, it does the equivalent of:
//
//	for i := 0; i < PinCount*len(devs); i++ {
//		if mask.Get(i) {
//			d.Pin(i).Set(pins.Get(i))
//		}
//	}
func (devs Devices) SetPins(pins, mask PinSlice) error {
	defaultPins := pins.extra()
	defaultMask := mask.extra()
	for i, dev := range devs {
		devPins := defaultPins
		if i < len(pins) {
			devPins = pins[i]
		}
		devMask := defaultMask
		if i < len(mask) {
			devMask = mask[i]
		}
		if err := dev.SetPins(devPins, devMask); err != nil {
			return err
		}
	}
	return nil
}

// TogglePins inverts the values on all pins for
// which mask is high.
func (devs Devices) TogglePins(mask PinSlice) error {
	defaultMask := mask.extra()
	for i, dev := range devs {
		devMask := defaultMask
		if i < len(mask) {
			devMask = mask[i]
		}
		if err := dev.TogglePins(devMask); err != nil {
			return err
		}
	}
	return nil
}

// PinSlice represents an arbitrary nunber of pins, each element corresponding
// to the pins for one device. The value of the highest numbered pin in the
// slice is extended to all other pins beyond the end of the slice.
type PinSlice []Pins

// Get returns the value for the given pin. If the length of pins is too short
// for the pin number, the value of the highest available pin is returned.
// That is, the highest numbered pin in the last element of pins
// is effectively replicated to all other elements.
//
// This means that PinSlice{} means "all pins high" and
// PinSlice{0xffff} means "all pins low".
func (pins PinSlice) Get(i int) bool {
	if len(pins) == 0 || i < 0 {
		return false
	}
	if i >= len(pins)*PinCount {
		return pins[len(pins)-1].Get(PinCount - 1)
	}
	return pins[i/PinCount].Get(i % PinCount)
}

// Set sets the value for the given pin.
func (pins PinSlice) Set(i int, value bool) {
	pins[i/PinCount].Set(i%PinCount, value)
}

// High is short for p.Set(pin, true).
func (pins PinSlice) High(pin int) {
	pins[pin/PinCount].High(pin % PinCount)
}

// High is short for p.Set(pin, false).
func (pins PinSlice) Low(pin int) {
	pins[pin/PinCount].Low(pin % PinCount)
}

// Toggle inverts the value of the given pin.
func (pins PinSlice) Toggle(pin int) {
	pins[pin/PinCount].Toggle(pin % PinCount)
}

// Ensure checks that pins has enough space to store
// at least length pins. If it does, it returns pins unchanged.
// Otherwise, it returns pins with elements appended as needed,
// populating additonal elements by replicating the
// highest pin (mirroring the behavior of PinSlice.Get).
func (pins PinSlice) Ensure(length int) PinSlice {
	if length == 0 {
		return pins
	}
	n := length/PinCount + 1
	if len(pins) >= n {
		return pins
	}
	// TODO we could potentially make use of additional
	// extra capacity in pins when available instead
	// of allocating a new slice always.
	newPins := make(PinSlice, n)
	copy(newPins, pins)
	if extend := pins.extra(); extend != 0 {
		for i := len(pins); i < n; i++ {
			newPins[i] = extend
		}
	}
	return newPins
}

// extra returns the value of implied extra elements beyond
// the end of pins.
func (pins PinSlice) extra() Pins {
	if len(pins) == 0 || !pins[len(pins)-1].Get(PinCount-1) {
		return 0
	}
	return ^Pins(0)
}
