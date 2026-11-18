// Package manchester implements Manchester line coding for OOK radio
// modules such as the SYN115 transmitter and the SYN480R receiver.
//
// A zero bit is a high level followed by a low level, a one bit is a low
// level followed by a high level. Bits are sent most significant bit first.
// Each frame starts with an all-zero preamble so the receiver can find the
// first rising edge and sample the half-bits that follow.
//
// The radio hardware is accessed through the Transmitter and Receiver
// interfaces. Any module that keys a carrier or outputs a demodulated level
// can be used with this package.
package manchester // import "tinygo.org/x/drivers/manchester"
