package irremote // import "tinygo.org/x/drivers/irremote"

import (
	"machine"
	"time"
)

// NEC protocol references
// https://www.sbprojects.net/knowledge/ir/nec.php
// https://techdocs.altium.com/display/FPGA/NEC+Infrared+Transmission+Protocol
// https://simple-circuit.com/arduino-nec-remote-control-decoder/

// Data encapsulates the data received by the ReceiverDevice.
type Data struct {
	// Code is the raw IR data received.
	Code uint32
	// Address is the decoded address from the IR data received.
	Address uint16
	// Command is the decoded command from the IR data recieved
	Command uint16
	// Flags provides additional information about the IR data received. See DataFlags
	Flags DataFlags
}

// DataFlags provides bitwise flags representing various information about recieved IR data.
type DataFlags uint16

// Valid values for DataFlags
const (
	// DataFlagIsRepeat set indicates that the IR data is a repeat commmand
	DataFlagIsRepeat DataFlags = 1 << iota
)

// CommandHandler defines the callback function used to provide IR data received by the ReceiverDevice.
type CommandHandler func(data Data)

// nec_ir_state represents the various internal states used to decode the NEC IR protocol commands
type nec_ir_state uint8

// Valid values for nec_ir_state
const (
	lead_pulse_start nec_ir_state = iota // Start receiving IR data, beginning of 9ms lead pulse
	lead_space_start                     // End of 9ms lead pulse, start of 4.5ms space
	lead_space_end                       // End of 4.5ms space, start of 562µs pulse
	bit_read_start                       // End of 562µs pulse, start of 562µs or 1687µs space
	bit_read_end                         // End of 562µs or 1687µs space
	trail_pulse_end                      // End of 562µs trailing pulse
)

// ReceiverDevice is the device for receiving IR commands
type ReceiverDevice struct {
	pin      machine.Pin    // IR input pin.
	ch       CommandHandler // client callback function
	necState nec_ir_state   // internal state machine
	data     Data           // decoded data for client
	lastTime time.Time      // used to track states
	bitIndex int            // tracks which bit (0-31) of necCode is being read
}

// NewReceiver returns a new IR receiver device
func NewReceiver(pin machine.Pin) ReceiverDevice {
	return ReceiverDevice{pin: pin}
}

// Configure configures the input pin for the IR receiver device
func (ir *ReceiverDevice) Configure() {
	// The IR receiver sends logic HIGH when NOT receiving IR, and logic LOW when receiving IR
	ir.pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
}

// SetCommandHandler is used to start or stop receiving IR commands via a callback function (pass nil to stop)
func (ir *ReceiverDevice) SetCommandHandler(ch CommandHandler) {
	ir.ch = ch
	ir.resetStateMachine()
	if ch != nil {
		// Start monitoring IR output pin for changes
		ir.pin.SetInterrupt(machine.PinFalling|machine.PinRising, ir.pinChange)
	} else {
		// Stop monitoring IR output pin for changes
		ir.pin.SetInterrupt(0, nil)
	}
}

// Internal helper function to reset state machine on protocol failure
func (ir *ReceiverDevice) resetStateMachine() {
	ir.data = Data{}
	ir.bitIndex = 0
	ir.necState = lead_pulse_start
}

// Internal pin rising/falling edge interrupt handler
func (ir *ReceiverDevice) pinChange(pin machine.Pin) {
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
		if !ir.pin.Get() {
			// IR is 'on' (pin is pulled high and sent low when IR is received)
			ir.necState = lead_space_start // move to next state
		}
	case lead_space_start:
		if duration > time.Microsecond*9500 || duration < time.Microsecond*8500 {
			// Invalid interval for 9ms lead pulse. Reset
			ir.resetStateMachine()
		} else {
			// 9ms lead pulse detected, move to next state
			ir.necState = lead_space_end
		}
	case lead_space_end:
		if duration > time.Microsecond*5000 || duration < time.Microsecond*1750 {
			// Invalid interval for 4.5ms lead space OR 2.25ms repeat space. Reset
			ir.resetStateMachine()
		} else {
			// 4.5ms lead space OR 2.25ms repeat space detected
			if duration > time.Microsecond*3000 {
				// 4.5ms lead space detected, new code incoming, move to next state
				ir.resetStateMachine()
				ir.necState = bit_read_start
			} else {
				// 2.25ms repeat space detected.
				if ir.data.Code != 0 {
					// Valid repeat code. Invoke client callback with repeat flag set
					ir.data.Flags |= DataFlagIsRepeat
					if ir.ch != nil {
						ir.ch(ir.data)
					}
					ir.necState = lead_pulse_start
				} else {
					// ir.data is not in a valid state for a repeat. Reset
					ir.resetStateMachine()
				}
			}
		}
	case bit_read_start:
		if duration > time.Microsecond*700 || duration < time.Microsecond*400 {
			// Invalid interval for 562.5µs pulse. Reset
			ir.resetStateMachine()
		} else {
			// 562.5µs pulse detected, move to next state
			ir.necState = bit_read_end
		}
	case bit_read_end:
		if duration > time.Microsecond*1800 || duration < time.Microsecond*400 {
			// Invalid interval for 562.5µs space OR 1687.5µs space. Reset
			ir.resetStateMachine()
		} else {
			// 562.5µs OR 1687.5µs space detected
			mask := uint32(1 << ir.bitIndex)
			if duration > time.Microsecond*1000 {
				// 1687.5µs space detected (logic 1) - Set bit
				ir.data.Code |= mask
			} else {
				// 562.5µs space detected (logic 0) - Clear bit
				ir.data.Code &^= mask
			}

			ir.bitIndex++
			if ir.bitIndex > 31 {
				// We've read all bits for this code, move to next state
				ir.necState = trail_pulse_end
			} else {
				// Read next bit
				ir.necState = bit_read_start
			}
		}
	case trail_pulse_end:
		if duration > time.Microsecond*700 || duration < time.Microsecond*400 {
			// Invalid interval for trailing 562.5µs pulse. Reset
			ir.resetStateMachine()
		} else {
			// 562.5µs trailing pulse detected. Decode & validate data
			err := ir.decode()
			if err == irDecodeErrorNone {
				// Valid data, invoke client callback
				if ir.ch != nil {
					ir.ch(ir.data)
				}
				// around we go again. Note: we don't resetStateMachine() since repeat codes are now possible
				ir.necState = lead_pulse_start
			} else {
				ir.resetStateMachine()
			}
		}
	}
}

// Error type for NEC format decoding
type irDecodeError int

// Valid values for irDecodeError
const (
	irDecodeErrorNone             irDecodeError = iota // no error occurred
	irDecodeErrorInverseCheckFail                      // validation of inverse cmd does not match cmd
)

func (ir *ReceiverDevice) decode() irDecodeError {
	// Decode cmd and inverse cmd and perform validation check
	cmd := uint8((ir.data.Code & 0x00ff0000) >> 16)
	invCmd := uint8((ir.data.Code & 0xff000000) >> 24)
	if cmd != ^invCmd {
		// Validation failure. cmd and inverse cmd do not match
		return irDecodeErrorInverseCheckFail
	}
	// cmd validation pass, decode address
	ir.data.Command = uint16(cmd)
	addrLow := uint8(ir.data.Code & 0xff)
	addrHigh := uint8((ir.data.Code & 0xff00) >> 8)
	if addrHigh == ^addrLow {
		// addrHigh is inverse of addrLow. This is not a valid 16-bit address in extended NEC coding
		// since it is indistinguishable from 8-bit address with inverse validation. Use the 8-bit address
		ir.data.Address = uint16(addrLow)
	} else {
		// 16-bit extended NEC address
		ir.data.Address = (uint16(addrHigh) << 8) | uint16(addrLow)
	}
	// Clear repeat flag
	ir.data.Flags &^= DataFlagIsRepeat
	return irDecodeErrorNone
}
