package sd

import (
	"bytes"
	"errors"
	"testing"
)

type fakeSPI struct {
	out  []byte
	sent []byte
}

func (f *fakeSPI) Transfer(b byte) (byte, error) {
	f.sent = append(f.sent, b)
	if len(f.out) == 0 {
		return 0xff, nil
	}
	r := f.out[0]
	f.out = f.out[1:]
	return r, nil
}

func (f *fakeSPI) Tx(w, r []byte) error {
	if len(w) > 0 && len(r) > 0 && &w[0] == &r[0] {
		return errors.New("w and r share a buffer")
	}
	for i := range r {
		var b byte = 0
		if w != nil {
			b = w[i]
		}
		r[i], _ = f.Transfer(b)
	}
	return nil
}

func TestReadDataSendsOnes(t *testing.T) {
	want := make([]byte, 100)
	for i := range want {
		want[i] = byte(i)
	}
	bus := &fakeSPI{out: append([]byte{tokSTART_BLOCK}, want...)}
	card := NewSPICard(bus, func(bool) {})
	got := make([]byte, len(want))
	if err := card.read_data(got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("got %x, want %x", got, want)
	}
	for i, b := range bus.sent {
		if b != 0xff {
			t.Fatalf("byte %d sent as %#x, want 0xff", i, b)
		}
	}
}
