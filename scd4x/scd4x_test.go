package scd4x

import (
	"testing"
)

type testI2CBus struct{}

func (testI2CBus) Tx(uint16, []byte, []byte) error { return nil }

func TestDefaultI2CAddress(t *testing.T) {
	dev := New(testI2CBus{})
	if dev.Address != uint8(Address) {
		t.Fatalf("Address = %#x, want %#x", dev.Address, Address)
	}
}

func TestCRC8(t *testing.T) {
	if got := crc8([]byte{0xbe, 0xef}); got != 0x92 {
		t.Fatalf("crc8(0xbeef) = %#x, want 0x92", got)
	}
}

func TestConversionsUseFullScale(t *testing.T) {
	dev := &Device{temperature: 0xffff, humidity: 0xffff}

	if got := dev.Temperature(); got != 130000 {
		t.Errorf("Temperature() = %d, want 130000", got)
	}
	if got := dev.Humidity(); got != 10000 {
		t.Errorf("Humidity() = %d, want 10000", got)
	}
}
