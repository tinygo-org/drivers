package irprotocol

import (
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
)

type necTestData struct {
	Code    uint32
	Address uint16
	Command byte
}

func decodeTests(t *testing.T, tests []necTestData, expectedValid bool) {
	c := qt.New(t)

	for _, data := range tests {
		name := fmt.Sprintf("Decode:Code:%08x Addr:%04x Cmd:%02x",
			data.Code, data.Address, data.Command)
		c.Run(name, func(c *qt.C) {
			valid, addr, cmd := SplitRawNECData(data.Code)
			c.Assert(valid, qt.Equals, expectedValid)
			if valid {
				c.Assert(addr, qt.Equals, data.Address)
				c.Assert(cmd, qt.Equals, data.Command)
			}
		})
	}
}

func encodeTests(t *testing.T, tests []necTestData) {
	c := qt.New(t)

	for _, data := range tests {
		name := fmt.Sprintf("Encode:Code:%08x Addr:%04x Cmd:%02x",
			data.Code, data.Address, data.Command)
		c.Run(name, func(c *qt.C) {
			code := MakeRawNECData(data.Address, data.Command)
			c.Assert(code, qt.Equals, data.Code)
		})
	}
}

// TestRawNECDataNonExtendedAddr encodes and decodes codes with 8 bit addresses.
func TestRawNECDataNonExtendedAddr(t *testing.T) {
	tests := []necTestData{
		{Code: 0xFF00FF00, Address: 0x0000, Command: 0x00},
		{Code: 0x00FFFF00, Address: 0x0000, Command: 0xFF},
		{Code: 0xFF0000FF, Address: 0x00FF, Command: 0x00},
		{Code: 0x00FF00FF, Address: 0x00FF, Command: 0xFF},
		{Code: 0xFF00DF20, Address: 0x0020, Command: 0x00},
		{Code: 0xFF0020DF, Address: 0x00DF, Command: 0x00},
		{Code: 0xDF20FF00, Address: 0x0000, Command: 0x20},
		{Code: 0x20DFFF00, Address: 0x0000, Command: 0xDF},
	}
	decodeTests(t, tests, true)
	encodeTests(t, tests)
}

// TestRawNECDataExtendedAddr encodes and decodes codes with 16 bit extended
// addresses.
func TestRawNECDataExtendedAddr(t *testing.T) {
	tests := []necTestData{
		{Code: 0xFF000100, Address: 0x0100, Command: 0x00},
		{Code: 0xFF00FE00, Address: 0xFE00, Command: 0x00},
		{Code: 0xFF00F00D, Address: 0xF00D, Command: 0x00},
	}
	decodeTests(t, tests, true)
	encodeTests(t, tests)
}

// TestSplitRawNECDataInvalidCommand checks that a command that does not match
// its inverse fails validation.
func TestSplitRawNECDataInvalidCommand(t *testing.T) {
	decodeTests(t,
		[]necTestData{
			// One wrong bit in each position of the inverse command.
			{Code: 0x01FFFF00, Address: 0x0000, Command: 0xFF},
			{Code: 0x02FFFF00, Address: 0x0000, Command: 0xFF},
			{Code: 0x04FFFF00, Address: 0x0000, Command: 0xFF},
			{Code: 0x08FFFF00, Address: 0x0000, Command: 0xFF},
			{Code: 0x10FFFF00, Address: 0x0000, Command: 0xFF},
			{Code: 0x20FFFF00, Address: 0x0000, Command: 0xFF},
			{Code: 0x40FFFF00, Address: 0x0000, Command: 0xFF},
			{Code: 0x80FFFF00, Address: 0x0000, Command: 0xFF},
			// One wrong bit in each position of the command.
			{Code: 0xFF01FF00, Address: 0x0000, Command: 0xFF},
			{Code: 0xFF02FF00, Address: 0x0000, Command: 0xFF},
			{Code: 0xFF04FF00, Address: 0x0000, Command: 0xFF},
			{Code: 0xFF08FF00, Address: 0x0000, Command: 0xFF},
			{Code: 0xFF10FF00, Address: 0x0000, Command: 0xFF},
			{Code: 0xFF20FF00, Address: 0x0000, Command: 0xFF},
			{Code: 0xFF40FF00, Address: 0x0000, Command: 0xFF},
			{Code: 0xFF80FF00, Address: 0x0000, Command: 0xFF},
		},
		false)
}
