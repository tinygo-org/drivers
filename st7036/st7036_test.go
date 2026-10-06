package st7036

import (
	"io"
	"reflect"
	"testing"
)

var _ io.Writer = (*Device)(nil)

type mockSPI struct {
	bytes []byte
	cs    *mockPin
	rs    *mockPin
	modes []bool
}

func (s *mockSPI) Transfer(b byte) (byte, error) {
	if s.cs.high {
		panic("chip select is high during transfer")
	}
	s.bytes = append(s.bytes, b)
	s.modes = append(s.modes, s.rs.high)
	return 0, nil
}

func (s *mockSPI) Tx(w, r []byte) error {
	panic("unexpected SPI Tx")
}

type mockPin struct {
	high   bool
	levels []bool
}

func (p *mockPin) High() {
	p.high = true
	p.levels = append(p.levels, true)
}

func (p *mockPin) Low() {
	p.high = false
	p.levels = append(p.levels, false)
}

func newTestDevice() (Device, *mockSPI) {
	cs := &mockPin{high: true}
	rs := &mockPin{}
	spi := &mockSPI{cs: cs, rs: rs}
	return Device{
		bus:                    spi,
		cs:                     cs,
		rs:                     rs,
		instructionSetTemplate: defaultInstructionSetTemplate,
	}, spi
}

func assertBytes(t *testing.T, spi *mockSPI, want []byte) {
	t.Helper()
	if !reflect.DeepEqual(spi.bytes, want) {
		t.Fatalf("SPI bytes = %x, want %x", spi.bytes, want)
	}
}

func TestConfigure(t *testing.T) {
	for _, tc := range []struct {
		name    string
		config  Config
		columns uint8
		rows    uint8
		offsets []uint8
	}{
		{"default", Config{}, 16, 3, []uint8{0, 0x10, 0x20}},
		{"one row", Config{Rows: 1, Columns: 8}, 8, 1, []uint8{0}},
		{"two rows", Config{Rows: 2, Columns: 20}, 20, 2, []uint8{0, 0x40}},
		{"three rows", Config{Rows: 3, Columns: 16}, 16, 3, []uint8{0, 0x10, 0x20}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, spi := newTestDevice()
			if err := d.Configure(tc.config); err != nil {
				t.Fatal(err)
			}
			columns, rows := d.Size()
			if columns != tc.columns || rows != tc.rows {
				t.Fatalf("Size() = (%d, %d), want (%d, %d)", columns, rows, tc.columns, tc.rows)
			}
			if !reflect.DeepEqual(d.rowOffsets, tc.offsets) {
				t.Fatalf("row offsets = %x, want %x", d.rowOffsets, tc.offsets)
			}
			assertBytes(t, spi, []byte{
				0x38, 0x0c,
				0x38, 0x06,
				0x39, 0x15,
				0x39, 0x56,
				0x39, 0x6b,
				0x39, 0x78,
				0x38, 0x01,
				0x38, 0x80,
			})
		})
	}
	d, spi := newTestDevice()
	if err := d.Configure(Config{Rows: 4}); err != errInvalidPosition {
		t.Fatalf("Configure() error = %v, want %v", err, errInvalidPosition)
	}
	if len(spi.bytes) != 0 {
		t.Fatal("invalid configuration sent SPI bytes")
	}
}

func TestContrast(t *testing.T) {
	for _, contrast := range []uint8{0, 15, 16, 40, 63} {
		d, spi := newTestDevice()
		if err := d.SetContrast(contrast); err != nil {
			t.Fatal(err)
		}
		assertBytes(t, spi, []byte{
			0x39, 0x54 | (contrast >> 4),
			0x39, 0x6b,
			0x39, 0x70 | (contrast & 0x0f),
		})
	}
	for _, contrast := range []uint8{64, 255} {
		d, spi := newTestDevice()
		if err := d.SetContrast(contrast); err != errInvalidContrast {
			t.Fatalf("SetContrast(%d) error = %v, want %v", contrast, err, errInvalidContrast)
		}
		if len(spi.bytes) != 0 {
			t.Fatal("invalid contrast sent SPI bytes")
		}
	}
}

func TestCursor(t *testing.T) {
	for _, rows := range []uint8{1, 2, 3} {
		d, spi := newTestDevice()
		if err := d.Configure(Config{Rows: rows}); err != nil {
			t.Fatal(err)
		}
		for row := uint8(0); row < rows; row++ {
			spi.bytes = nil
			if err := d.SetCursorPosition(15, row); err != nil {
				t.Fatal(err)
			}
			offsets := []uint8{0, 0x10, 0x20}
			if rows == 2 {
				offsets[1] = 0x40
			}
			assertBytes(t, spi, []byte{0x38, 0x80 | (offsets[row] + 15)})
		}
		for _, position := range [][2]uint8{{16, 0}, {0, rows}, {255, 255}} {
			spi.bytes = nil
			if err := d.SetCursorPosition(position[0], position[1]); err != errInvalidPosition {
				t.Fatalf("SetCursorPosition(%v) error = %v", position, err)
			}
			if len(spi.bytes) != 0 {
				t.Fatal("invalid position sent SPI bytes")
			}
		}
		spi.bytes = nil
		if err := d.Clear(); err != nil {
			t.Fatal(err)
		}
		assertBytes(t, spi, []byte{0x38, 0x01, 0x38, 0x80})
	}
}

func TestDisplayMode(t *testing.T) {
	for mode := uint8(0); mode < 8; mode++ {
		d, spi := newTestDevice()
		d.SetDisplayMode(mode&4 != 0, mode&2 != 0, mode&1 != 0)
		assertBytes(t, spi, []byte{0x38, 0x08 | mode})
	}
}

func TestCommandFraming(t *testing.T) {
	d, spi := newTestDevice()
	d.SetCursorOffset(0x20)
	d.SetBias(1)
	assertBytes(t, spi, []byte{0x38, 0xa0, 0x39, 0x15})
	for i, mode := range spi.modes {
		if mode {
			t.Fatalf("command byte %d sent in data mode", i)
		}
	}
	wantLevels := []bool{false, true, false, true, false, true, false, true}
	if !reflect.DeepEqual(spi.cs.levels, wantLevels) {
		t.Fatalf("chip-select levels = %v, want %v", spi.cs.levels, wantLevels)
	}
}

func TestWrite(t *testing.T) {
	d, spi := newTestDevice()
	data := []byte{'L', 'C', 'D', 0, 7}
	n, err := d.Write(data)
	if err != nil || n != len(data) {
		t.Fatalf("Write() = (%d, %v), want (%d, nil)", n, err, len(data))
	}
	assertBytes(t, spi, data)
	for i, mode := range spi.modes {
		if !mode {
			t.Fatalf("byte %d sent in command mode", i)
		}
	}
	if !spi.cs.high {
		t.Fatal("chip select is low after Write")
	}
	wantLevels := []bool{false, true, false, true, false, true, false, true, false, true}
	if !reflect.DeepEqual(spi.cs.levels, wantLevels) {
		t.Fatalf("chip-select levels = %v, want %v", spi.cs.levels, wantLevels)
	}
	spi.bytes = nil
	n, err = d.Write(nil)
	if n != 0 || err != nil || len(spi.bytes) != 0 {
		t.Fatalf("Write(nil) = (%d, %v), SPI bytes = %x", n, err, spi.bytes)
	}
}

func TestReset(t *testing.T) {
	d, spi := newTestDevice()
	d.Reset()
	reset := &mockPin{high: true}
	d.reset = reset
	d.Reset()
	if !reflect.DeepEqual(reset.levels, []bool{false, true}) {
		t.Fatalf("reset levels = %v, want [false true]", reset.levels)
	}
	if len(spi.bytes) != 0 {
		t.Fatal("Reset sent SPI bytes")
	}
}

func TestCreateChar(t *testing.T) {
	for _, slot := range []uint8{0, 7} {
		d, spi := newTestDevice()
		if err := d.Configure(Config{}); err != nil {
			t.Fatal(err)
		}
		spi.bytes = nil
		spi.modes = nil
		charMap := [8]byte{1, 2, 4, 8, 16, 8, 4, 2}
		if err := d.CreateChar(slot, charMap); err != nil {
			t.Fatal(err)
		}
		var want []byte
		for i, value := range charMap {
			want = append(want, 0x38, 0x40|(slot*8+uint8(i)), value)
		}
		want = append(want, 0x38, 0x80)
		assertBytes(t, spi, want)
		for i, mode := range spi.modes {
			wantData := i < 24 && i%3 == 2
			if mode != wantData {
				t.Fatalf("byte %d data mode = %v, want %v", i, mode, wantData)
			}
		}
	}
	d, spi := newTestDevice()
	if err := d.CreateChar(8, [8]byte{}); err != errInvalidCharSlot {
		t.Fatalf("CreateChar() error = %v, want %v", err, errInvalidCharSlot)
	}
	if len(spi.bytes) != 0 {
		t.Fatal("invalid character slot sent SPI bytes")
	}
}
