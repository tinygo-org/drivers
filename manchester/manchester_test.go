package manchester

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

func TestEncode(t *testing.T) {
	for _, tc := range []struct {
		data byte
		want []bool
	}{
		{0x00, repeatPair(true, false, 8)},
		{0x01, append(repeatPair(true, false, 7), false, true)},
		{0x80, append([]bool{false, true}, repeatPair(true, false, 7)...)},
		{0xff, repeatPair(false, true, 8)},
	} {
		got := Encode([]byte{tc.data})
		if !equal(got, tc.want) {
			t.Errorf("Encode(%#02x) = %v, want %v", tc.data, got, tc.want)
		}
	}
}

func TestDecode(t *testing.T) {
	var data []byte
	for i := range 256 {
		data = append(data, byte(i))
	}
	got, err := Decode(Encode(data))
	if err != nil {
		t.Fatalf("Decode returned an error: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("round trip changed the data")
	}
}

func TestDecodeFrameLength(t *testing.T) {
	if _, err := Decode([]bool{true, false, true}); !errors.Is(err, ErrFrameLength) {
		t.Fatalf("Decode with 3 levels returned %v, want %v", err, ErrFrameLength)
	}
}

func TestDecodeSymbol(t *testing.T) {
	wave := Encode([]byte{0})
	wave[0] = false
	if _, err := Decode(wave); !errors.Is(err, ErrSymbol) {
		t.Fatalf("Decode with an invalid symbol returned %v, want %v", err, ErrSymbol)
	}
}

func TestEncoderWrite(t *testing.T) {
	tx := &fakeTransmitter{}
	enc := NewEncoder(tx, Config{BitRate: 100000, Preamble: 1})

	n, err := enc.Write([]byte{0x80})
	if err != nil {
		t.Fatalf("Write returned an error: %v", err)
	}
	if n != 1 {
		t.Fatalf("Write returned %d, want 1", n)
	}

	want := append(Encode([]byte{0x00}), Encode([]byte{0x80})...)
	want = append(want, false)
	if !equal(tx.levels, want) {
		t.Fatalf("Write sent %v, want %v", tx.levels, want)
	}
}

func TestDecoderRead(t *testing.T) {
	data := []byte{0x00, 0xa5, 0xff}
	wave := append(Encode(make([]byte, 2)), Encode(data)...)
	rx := &fakeReceiver{levels: wave, delay: 3}

	dec := NewDecoder(rx, Config{BitRate: 100000, Preamble: 2, Timeout: time.Second})
	got := make([]byte, len(data))
	n, err := dec.Read(got)
	if err != nil {
		t.Fatalf("Read returned an error: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Read returned %d, want %d", n, len(data))
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("Read returned %x, want %x", got, data)
	}
}

func TestDecoderSettling(t *testing.T) {
	data := []byte{0xa5}
	// The first two preamble bytes are disturbed, the last two are checked.
	wave := append(Encode([]byte{0x00, 0x40}), Encode(make([]byte, 2))...)
	wave = append(wave, Encode(data)...)
	dec := NewDecoder(&fakeReceiver{levels: wave, delay: 3}, Config{BitRate: 100000, Preamble: 4, Timeout: time.Second})

	got := make([]byte, len(data))
	if _, err := dec.Read(got); err != nil {
		t.Fatalf("Read returned an error: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("Read returned %x, want %x", got, data)
	}
}

func TestDecoderReadEmpty(t *testing.T) {
	dec := NewDecoder(&fakeReceiver{}, Config{})
	if n, err := dec.Read(nil); n != 0 || err != nil {
		t.Fatalf("Read(nil) returned %d, %v", n, err)
	}
}

func TestDecoderTimeout(t *testing.T) {
	dec := NewDecoder(&fakeReceiver{}, Config{BitRate: 100000, Timeout: time.Millisecond})
	n, err := dec.Read(make([]byte, 1))
	if n != 0 || !errors.Is(err, ErrTimeout) {
		t.Fatalf("Read returned %d, %v, want 0, %v", n, err, ErrTimeout)
	}
}

func TestDecoderBadPreamble(t *testing.T) {
	wave := append(Encode([]byte{0x00, 0xff}), Encode([]byte{0xa5})...)
	dec := NewDecoder(&fakeReceiver{levels: wave, delay: 3}, Config{BitRate: 100000, Preamble: 2, Timeout: 5 * time.Millisecond})
	n, err := dec.Read(make([]byte, 1))
	if n != 0 || !errors.Is(err, ErrTimeout) {
		t.Fatalf("Read returned %d, %v, want 0, %v", n, err, ErrTimeout)
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := Config{}
	if got := cfg.bitPeriod(); got != time.Second/DefaultBitRate {
		t.Errorf("bitPeriod() = %v, want %v", got, time.Second/DefaultBitRate)
	}
	if got := cfg.preambleBytes(); got != DefaultPreamble {
		t.Errorf("preambleBytes() = %d, want %d", got, DefaultPreamble)
	}

	cfg = Config{BitRate: 1000, Preamble: 4}
	if got := cfg.bitPeriod(); got != time.Millisecond {
		t.Errorf("bitPeriod() = %v, want %v", got, time.Millisecond)
	}
	if got := cfg.preambleBytes(); got != 4 {
		t.Errorf("preambleBytes() = %d, want 4", got)
	}
}

type fakeTransmitter struct {
	levels []bool
}

func (tx *fakeTransmitter) Set(high bool) {
	tx.levels = append(tx.levels, high)
}

type fakeReceiver struct {
	levels []bool
	delay  int
	calls  int
}

func (rx *fakeReceiver) Get() bool {
	if rx.calls < rx.delay {
		rx.calls++
		return false
	}
	i := rx.calls - rx.delay
	rx.calls++
	if i < len(rx.levels) {
		return rx.levels[i]
	}
	return false
}

func repeatPair(first, second bool, n int) []bool {
	wave := make([]bool, 0, 2*n)
	for range n {
		wave = append(wave, first, second)
	}
	return wave
}

func equal(a, b []bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
