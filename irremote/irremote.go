package irremote // import "tinygo.org/x/drivers/irremote"

import (
	"machine"
	"time"
)

// NEC protocol references
// https://www.sbprojects.net/knowledge/ir/nec.php
// https://techdocs.altium.com/display/FPGA/NEC+Infrared+Transmission+Protocol
// https://simple-circuit.com/arduino-nec-remote-control-decoder/

// Client callback function for IR commands
// 'code' is the raw NEC format code. 'addr' & 'cmd' are the NEC decoded address and command codes respectively
// See the NEC protocol reference links above for more information
type IRCallback func(code uint32, addr uint16, cmd uint8, repeat bool)

// Public API for an IR receiver device
type IRReceiverDevice interface {
	// Configure pins for the IR receiver module
	Configure()
	// Start/Stop receiving IR callbacks (pass nil to stop)
	Callback(cb IRCallback)
}

// Return a new IR receiver device
func New(pin machine.Pin) IRReceiverDevice {
	return &irReceiver{pin: pin}
}

// Internal NEC IR protocol states
const (
	lead_pulse_start = iota // Start receiving IR data, beginning of 9ms lead pulse
	lead_space_start        // End of 9ms lead pulse, start of 4.5ms space
	lead_space_end          // End of 4.5ms space, start of 562µs pulse
	bit_read_start          // End of 562µs pulse, start of 562µs or 1687µs space
	bit_read_end            // End of 562µs or 1687µs space
)

// Internal IR device struct
type irReceiver struct {
	pin      machine.Pin // IR output pin.
	cb       IRCallback  // client callback
	necState int         // state machine
	necCode  uint32      // data read
	lastTime time.Time   // used to track states
	bitIndex int         // tracks which bit of necCode is being read
}

func (ir *irReceiver) Configure() {
	// The IR receiver sends logic HIGH when NOT receiving IR, and logic LOW when receiving IR
	ir.pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
}

func (ir *irReceiver) Callback(cb IRCallback) {
	ir.cb = cb
	ir.necState = lead_pulse_start
	if cb != nil {
		// Start monitoring IR output pin for changes
		ir.pin.SetInterrupt(machine.PinFalling|machine.PinRising, ir.pinChange)
	} else {
		// Stop monitoring IR output pin for changes
		ir.pin.SetInterrupt(0, nil)
	}
}

// Internal pin rising/falling edge interrupt handler
func (ir *irReceiver) pinChange(pin machine.Pin) {
	/* Currently TinyGo is sending machine.NoPin (0xff) for all pins, at least on RP2040
	if pin != ir.pin {
		return // This is not the pin you're looking for
	}
	*/
	now := time.Now()
	duration := now.Sub(ir.lastTime)
	ir.lastTime = now
	switch ir.necState {
	case lead_pulse_start:
		ir.necCode = 0
		ir.bitIndex = 0
		if !ir.pin.Get() {
			// IR is 'on' (pin is pulled high and sent low when IR is received)
			ir.necState = lead_space_start // move to next state
		}
	case lead_space_start:
		if duration > time.Microsecond*9500 || duration < time.Microsecond*8500 {
			// Invalid interval for lead 9ms pulse. Reset
			ir.necState = lead_pulse_start
		} else {
			// Lead pulse detected, move to next state
			ir.necState = lead_space_end
		}
	case lead_space_end:
		if duration > time.Microsecond*5000 || duration < time.Microsecond*4000 {
			// Invalid interval for 4.5ms lead space. Reset
			ir.necState = lead_pulse_start
		} else {
			// Lead space detected, move to next state
			ir.necState = bit_read_start
		}
	case bit_read_start:
		if duration > time.Microsecond*700 || duration < time.Microsecond*400 {
			// Invalid interval for 562.5µs pulse. Reset
			ir.necState = lead_pulse_start
		} else {
			// 562.5µs pulse detected, move to next state
			ir.necState = bit_read_end
		}
	case bit_read_end:
		if duration > time.Microsecond*1800 || duration < time.Microsecond*400 {
			// Invalid interval for 562.5µs OR 1687.5µs space. Reset
			ir.necState = lead_pulse_start
		} else {
			// 562.5µs OR 1687.5µs space detected
			mask := uint32((1 << (31 - ir.bitIndex)))
			if duration > time.Microsecond*1000 {
				// 562.5µs space detected (logic 1) - Set bit
				ir.necCode |= mask
			} else {
				// 1687.5µs space detected (logic 0) - Clear bit
				ir.necCode &^= mask
			}

			ir.bitIndex++
			if ir.bitIndex > 31 {
				// We've read all bits for this code. around we go again
				ir.necState = lead_pulse_start
				// Decode (address, command) & validate
				addr, cmd, err := ir.decode()
				if irDecodeErrorNone == err {
					// valid addr & cmd. Inovke client callback function
					// TODO: repeat codes. Always send 'false' for now
					ir.cb(ir.necCode, addr, cmd, false)
				}
			} else {
				// Read next bit
				ir.necState = bit_read_start
			}
		}
	}
}

// Error type for NEC format decoding
type irDecodeError int

const (
	irDecodeErrorNone             = iota // no error occurred
	irDecodeErrorInverseCheckFail        // validation of inverse cmd does not match cmd
)

func (ir *irReceiver) decode() (addr uint16, cmd uint8, err irDecodeError) {
	addr, cmd, err = 0, 0, irDecodeErrorNone
	// Decode cmd and inverse cmd and perform validation check
	cmd = uint8((ir.necCode & 0xff00) >> 8)
	invCmd := uint8(ir.necCode & 0xff)
	if cmd != ^invCmd {
		// Validation failure. cmd and inverse cmd do not match
		err = irDecodeErrorInverseCheckFail
		return
	}
	// cmd validation pass, decode address
	addrLow := uint8((ir.necCode & 0xff000000) >> 24)
	addrHigh := uint8((ir.necCode & 0x00ff0000) >> 16)
	if addrHigh == ^addrLow {
		// addrHigh is inverse of addrLow. This is not a valid 16-bit address in extended NEC coding
		// since it is indistinguishable from 8-bit address with inverse validation. Use the 8-bit address
		addr = uint16(addrLow)
	} else {
		// 16-bit extended NEC address
		addr = (uint16(addrHigh) << 8) | uint16(addrLow)
	}
	return
}
