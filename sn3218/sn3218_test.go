package sn3218_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"tinygo.org/x/drivers/sn3218"
)

type i2cFunc func(addr uint16, w, r []byte) error

func (f i2cFunc) Tx(addr uint16, w, r []byte) error {
	return f(addr, w, r)
}

func TestNew(t *testing.T) {
	calls := 0
	dev := sn3218.New(i2cFunc(func(uint16, []byte, []byte) error {
		calls++
		return nil
	}))
	if dev.Address != 0x54 {
		t.Fatalf("address = %#x, want 0x54", dev.Address)
	}
	if sn3218.NumChannels != 18 {
		t.Fatalf("channels = %d, want 18", sn3218.NumChannels)
	}
	if calls != 0 {
		t.Fatalf("New made %d transactions", calls)
	}
}

func TestOperations(t *testing.T) {
	var values [sn3218.NumChannels]uint8
	for i := range values {
		values[i] = uint8(i * 15)
	}
	pwm := append([]byte{0x01}, values[:]...)
	tests := []struct {
		name   string
		writes [][]byte
		call   func(*sn3218.Device) error
	}{
		{
			name:   "Configure",
			writes: [][]byte{{0x17, 0xFF}, {0x13, 0x3F, 0x3F, 0x3F}, {0x16, 0xFF}, {0x00, 0x01}},
			call:   (*sn3218.Device).Configure,
		},
		{
			name:   "Enable",
			writes: [][]byte{{0x00, 0x01}},
			call:   (*sn3218.Device).Enable,
		},
		{
			name:   "Disable",
			writes: [][]byte{{0x00, 0x00}},
			call:   (*sn3218.Device).Disable,
		},
		{
			name:   "Reset",
			writes: [][]byte{{0x17, 0xFF}},
			call:   (*sn3218.Device).Reset,
		},
		{
			name:   "EnableChannels",
			writes: [][]byte{{0x13, 0x01, 0x02, 0x04}, {0x16, 0xFF}},
			call:   func(d *sn3218.Device) error { return d.EnableChannels(0x4081) },
		},
		{
			name:   "SetChannels",
			writes: [][]byte{pwm, {0x16, 0xFF}},
			call:   func(d *sn3218.Device) error { return d.SetChannels(values) },
		},
		{
			name:   "SetChannelsOff",
			writes: [][]byte{append([]byte{0x01}, make([]byte, sn3218.NumChannels)...), {0x16, 0xFF}},
			call:   func(d *sn3218.Device) error { return d.SetChannels([sn3218.NumChannels]uint8{}) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("success", func(t *testing.T) {
				checkTransactions(t, 0x54, tt.writes, -1, tt.call)
			})
			for failAt := range tt.writes {
				t.Run(fmt.Sprintf("error%d", failAt), func(t *testing.T) {
					checkTransactions(t, 0x54, tt.writes, failAt, tt.call)
				})
			}
		})
	}
}

func TestEnableChannels(t *testing.T) {
	for channel := 0; channel < sn3218.NumChannels; channel++ {
		t.Run(fmt.Sprintf("channel%d", channel), func(t *testing.T) {
			mask := uint32(1) << channel
			write := []byte{0x13, 0, 0, 0}
			write[1+channel/6] = 1 << (channel % 6)
			checkTransactions(t, 0x54, [][]byte{write, {0x16, 0xFF}}, -1,
				func(d *sn3218.Device) error { return d.EnableChannels(mask) })
		})
	}
	tests := []struct {
		name string
		mask uint32
		data []byte
	}{
		{"none", 0, []byte{0x13, 0, 0, 0}},
		{"all", 0x3FFFF, []byte{0x13, 0x3F, 0x3F, 0x3F}},
		{"high bits ignored", 0xFFFC0000, []byte{0x13, 0, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkTransactions(t, 0x54, [][]byte{tt.data, {0x16, 0xFF}}, -1,
				func(d *sn3218.Device) error { return d.EnableChannels(tt.mask) })
		})
	}
}

func TestAddressOverride(t *testing.T) {
	checkTransactions(t, 0x55,
		[][]byte{{0x17, 0xFF}, {0x13, 0x3F, 0x3F, 0x3F}, {0x16, 0xFF}, {0x00, 0x01}},
		-1, (*sn3218.Device).Configure)
}

func checkTransactions(t *testing.T, address uint16, writes [][]byte, failAt int, call func(*sn3218.Device) error) {
	t.Helper()
	busError := errors.New("I2C write failed")
	calls := 0
	bus := i2cFunc(func(addr uint16, w, r []byte) error {
		t.Helper()
		index := calls
		calls++
		if index >= len(writes) {
			t.Fatalf("unexpected transaction %d: %x", index, w)
		}
		if addr != address || !bytes.Equal(w, writes[index]) || len(r) != 0 {
			t.Fatalf("transaction %d = (%#x, %x, %x), want (%#x, %x, no read)",
				index, addr, w, r, address, writes[index])
		}
		if index == failAt {
			return busError
		}
		return nil
	})
	dev := sn3218.New(bus)
	dev.Address = uint8(address)
	err := call(&dev)
	wantCalls := len(writes)
	if failAt >= 0 {
		wantCalls = failAt + 1
		if err != busError {
			t.Fatalf("error = %v, want %v", err, busError)
		}
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != wantCalls {
		t.Fatalf("transactions = %d, want %d", calls, wantCalls)
	}
}
