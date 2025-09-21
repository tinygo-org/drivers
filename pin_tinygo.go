//go:build tinygo

// Implementation of the Pin interface for tinygo (baremetal), depends on "machine" package.
// SafePin, SafePinInput and SafePinOutput wrappers must be used in constructors of drivers.

package drivers

import (
	"machine"
)

// MachinePin wraps machine.Pin to ensure correct pin mode is set when Get or Set methods are called.

type MachinePin struct {
	pin     machine.Pin
	mode    machine.PinMode
	modeSet bool
}

func (p *MachinePin) Get() bool {
	if !p.modeSet || p.mode != machine.PinInput {
		p.pin.Configure(machine.PinConfig{Mode: machine.PinInput})
		p.mode = machine.PinInput
	}
	return p.pin.Get()
}

func (p *MachinePin) Set(high bool) {
	if !p.modeSet || p.mode != machine.PinOutput {
		p.pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		p.mode = machine.PinOutput
	}
	p.pin.Set(high)
}

// Wrappers for Pin, PinInput and PinOutput interfaces that convert machine.Pin to safe MachinePin.

func SafePinInput(pin PinInput) PinInput {
	if p, ok := pin.(machine.Pin); ok {
		return &MachinePin{pin: p}
	}
	return pin
}

func SafePinOutput(pin PinOutput) PinOutput {
	if p, ok := pin.(machine.Pin); ok {
		return &MachinePin{pin: p}
	}
	return pin
}

func SafePin(pin Pin) Pin {
	if p, ok := pin.(machine.Pin); ok {
		return &MachinePin{pin: p}
	}
	return pin
}
