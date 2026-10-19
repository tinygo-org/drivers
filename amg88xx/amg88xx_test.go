package amg88xx

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

var errBus = errors.New("I2C failure")

type testBus struct {
	t         *testing.T
	registers [256]uint8
	writes    [][]uint8
	fail      bool
}

func (b *testBus) Tx(addr uint16, w, r []byte) error {
	if addr != AddressHigh {
		b.t.Fatalf("address = %#x, want %#x", addr, AddressHigh)
	}
	if b.fail {
		return errBus
	}
	if len(w) == 0 {
		b.t.Fatal("no register written")
	}
	if len(r) > 0 {
		copy(r, b.registers[w[0]:])
		return nil
	}
	b.writes = append(b.writes, append([]uint8(nil), w...))
	copy(b.registers[w[0]:], w[1:])
	return nil
}

func newTestDevice(t *testing.T) (*Device, *testBus) {
	bus := &testBus{t: t}
	dev := New(bus)
	return &dev, bus
}

func TestConfigure(t *testing.T) {
	dev, bus := newTestDevice(t)
	bus.registers[PIXEL_OFFSET+127] = 0x01
	if err := dev.Configure(Config{}); err != nil {
		t.Fatal(err)
	}
	want := [][]uint8{
		{PCTL, NORMAL_MODE},
		{RST, INITIAL_RESET},
		{FPSC, FPS_10},
	}
	if !reflect.DeepEqual(bus.writes, want) {
		t.Fatalf("writes = %#v, want %#v", bus.writes, want)
	}
}

func TestConfigureNoFrame(t *testing.T) {
	dev, _ := newTestDevice(t)
	start := time.Now()
	if err := dev.Configure(Config{}); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed < time.Second {
		t.Fatalf("returned after %v, want at least 1s", elapsed)
	}
}

func TestReadPixels(t *testing.T) {
	dev, bus := newTestDevice(t)
	tests := []struct {
		raw  uint16
		want int32
	}{
		{0x000, 0},
		{0x001, 250},
		{0x064, 25000},
		{0x140, 80000},
		{0x7FF, 511750},
		{0xFFF, -250},
		{0xF38, -50000},
		{0x800, -512000},
	}
	for i := 0; i < 64; i++ {
		raw := tests[i%len(tests)].raw
		bus.registers[PIXEL_OFFSET+2*i] = uint8(raw)
		bus.registers[PIXEL_OFFSET+2*i+1] = uint8(raw >> 8)
	}
	var pixels [64]int32
	if err := dev.ReadPixels(&pixels); err != nil {
		t.Fatal(err)
	}
	for i, got := range pixels {
		if want := tests[i%len(tests)].want; got != want {
			t.Errorf("pixel %d = %d, want %d", i, got, want)
		}
	}
}

func TestReadThermistor(t *testing.T) {
	dev, bus := newTestDevice(t)
	tests := []struct {
		raw  uint16
		want int32
	}{
		{0x000, 0},
		{0x001, 62},
		{0x190, 25000},
		{0x7FF, 127937},
		{0x801, -62},
		{0x8A0, -10000},
	}
	for _, tc := range tests {
		bus.registers[TTHL] = uint8(tc.raw)
		bus.registers[TTHH] = uint8(tc.raw >> 8)
		got, err := dev.ReadThermistor()
		if err != nil || got != tc.want {
			t.Errorf("raw %#x = %d, %v, want %d", tc.raw, got, err, tc.want)
		}
	}
}

func TestSettings(t *testing.T) {
	dev, bus := newTestDevice(t)
	steps := []struct {
		call func() error
		want []uint8
	}{
		{func() error { return dev.SetPCTL(SLEEP_MODE) }, []uint8{PCTL, 0x10}},
		{func() error { return dev.SetFrameRate(FPS_1) }, []uint8{FPSC, 0x01}},
		{func() error { return dev.SetFrameRate(0xFE) }, []uint8{FPSC, 0x00}},
		{func() error { return dev.SetMovingAverageMode(true) }, []uint8{AVE, 0x20}},
		{func() error { return dev.SetMovingAverageMode(false) }, []uint8{AVE, 0x00}},
		{dev.EnableInterrupt, []uint8{INTC, 0x01}},
		{func() error { return dev.SetInterruptMode(ABSOLUTE_VALUE) }, []uint8{INTC, 0x03}},
		{dev.DisableInterrupt, []uint8{INTC, 0x02}},
		{func() error { return dev.SetInterruptMode(DIFFERENCE) }, []uint8{INTC, 0x00}},
		{dev.ClearInterrupt, []uint8{RST, FLAG_RESET}},
		{func() error { return dev.ClearStatus(0xFF) }, []uint8{SCLR, 0x0E}},
	}
	for i, step := range steps {
		bus.writes = nil
		if err := step.call(); err != nil {
			t.Fatal(err)
		}
		if len(bus.writes) != 1 || !reflect.DeepEqual(bus.writes[0], step.want) {
			t.Errorf("step %d writes = %#v, want %#v", i, bus.writes, step.want)
		}
	}
}

func TestInterruptLevels(t *testing.T) {
	dev, bus := newTestDevice(t)
	if err := dev.SetInterruptLevelsHysteresis(30000, -10000, 1000000); err != nil {
		t.Fatal(err)
	}
	want := []uint8{INTHL, 0x78, 0x00, 0xD8, 0x0F, 0xFF, 0x07}
	if len(bus.writes) != 1 || !reflect.DeepEqual(bus.writes[0], want) {
		t.Fatalf("writes = %#v, want %#v", bus.writes, want)
	}

	bus.writes = nil
	if err := dev.SetInterruptLevels(40000, -1000000); err != nil {
		t.Fatal(err)
	}
	want = []uint8{INTHL, 0xA0, 0x00, 0x00, 0x08, 0x98, 0x00}
	if len(bus.writes) != 1 || !reflect.DeepEqual(bus.writes[0], want) {
		t.Fatalf("writes = %#v, want %#v", bus.writes, want)
	}
}

func TestInterruptTable(t *testing.T) {
	dev, bus := newTestDevice(t)
	copy(bus.registers[INT_OFFSET:], []uint8{0x01, 0, 0, 0x80, 0, 0, 0, 0x81})
	var table [8]uint8
	if err := dev.ReadInterruptTable(&table); err != nil {
		t.Fatal(err)
	}
	for pixel := 0; pixel < 64; pixel++ {
		want := pixel == 0 || pixel == 31 || pixel == 56 || pixel == 63
		if got := PixelInterrupt(&table, pixel); got != want {
			t.Errorf("pixel %d = %v, want %v", pixel, got, want)
		}
	}
}

func TestReadStatus(t *testing.T) {
	dev, bus := newTestDevice(t)
	bus.registers[STAT] = STATUS_INTF | STATUS_OVF_THS
	status, err := dev.ReadStatus()
	if err != nil || status != STATUS_INTF|STATUS_OVF_THS {
		t.Fatalf("status = %#x, %v", status, err)
	}
}

func TestErrors(t *testing.T) {
	dev, bus := newTestDevice(t)
	bus.fail = true
	var pixels [64]int32
	var table [8]uint8
	calls := map[string]func() error{
		"Configure":          func() error { return dev.Configure(Config{}) },
		"ReadPixels":         func() error { return dev.ReadPixels(&pixels) },
		"ReadInterruptTable": func() error { return dev.ReadInterruptTable(&table) },
		"SetInterruptLevels": func() error { return dev.SetInterruptLevels(1000, 0) },
		"EnableInterrupt":    dev.EnableInterrupt,
		"ReadStatus": func() error {
			_, err := dev.ReadStatus()
			return err
		},
		"ReadThermistor": func() error {
			_, err := dev.ReadThermistor()
			return err
		},
	}
	for name, call := range calls {
		if err := call(); err != errBus {
			t.Errorf("%s error = %v, want %v", name, err, errBus)
		}
	}
}

func TestNoAllocations(t *testing.T) {
	dev, bus := newTestDevice(t)
	bus.fail = true
	var pixels [64]int32
	var table [8]uint8
	allocs := testing.AllocsPerRun(100, func() {
		dev.ReadPixels(&pixels)
		dev.ReadThermistor()
		dev.ReadStatus()
		dev.ReadInterruptTable(&table)
		dev.SetPCTL(NORMAL_MODE)
		dev.SetFrameRate(FPS_10)
		dev.SetMovingAverageMode(true)
		dev.SetInterruptLevels(30000, 10000)
		dev.SetInterruptMode(ABSOLUTE_VALUE)
		dev.EnableInterrupt()
		dev.ClearInterrupt()
		dev.ClearStatus(STATUS_INTF)
	})
	if allocs != 0 {
		t.Fatalf("allocations = %v, want 0", allocs)
	}
}
