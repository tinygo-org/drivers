package irprotocol // import "tinygo.org/x/drivers/irremote/irprotocol"

import "time"

// NEC protocol reference
// https://www.sbprojects.net/knowledge/ir/nec.php

// NEC timings.
const (
	NECCarrierHz = 38000

	NECUnit        = time.Nanosecond * 562_500
	NECLeadMark    = NECUnit * 16
	NECLeadSpace   = NECUnit * 8
	NECRepeatSpace = NECUnit * 4
	NECBitMark     = NECUnit
	NECZeroSpace   = NECUnit
	NECOneSpace    = NECUnit * 3
	NECTrailMark   = NECUnit

	// NECRepeatPeriod is the time from the start of one frame to the start of
	// the next repeat frame while a button stays down.
	NECRepeatPeriod = NECUnit * 192
)

// SplitRawNECData breaks a raw NEC code into its parts and reports whether the
// command and its inverse agree.
func SplitRawNECData(data uint32) (valid bool, address uint16, command byte) {
	addrLow := byte(data)
	addrHigh := byte(data >> 8)
	command = byte(data >> 16)
	invCmd := byte(data >> 24)
	address = MakeNECAddress(addrLow, addrHigh)
	valid = command == ^invCmd
	return
}

// MakeRawNECData assembles a raw NEC code from an address and a command.
func MakeRawNECData(address uint16, command byte) uint32 {
	addrLow, addrHigh := SplitNECAddress(address)
	return uint32(^command)<<24 | uint32(command)<<16 | uint32(addrHigh)<<8 | uint32(addrLow)
}

// SplitNECAddress splits an NEC address into low and high bytes.
func SplitNECAddress(address uint16) (addrLow, addrHigh byte) {
	addrLow = byte(address)
	addrHigh = byte(address >> 8)
	if addrHigh == 0 {
		addrHigh = ^addrLow
	}
	return addrLow, addrHigh
}

// MakeNECAddress assembles an NEC address from low and high bytes. A high byte
// that is the inverse of the low byte is the 8 bit form, not an extended address.
func MakeNECAddress(addrLow, addrHigh byte) uint16 {
	if addrHigh == ^addrLow {
		return uint16(addrLow)
	}
	return uint16(addrHigh)<<8 | uint16(addrLow)
}
