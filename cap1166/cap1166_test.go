package cap1166

import (
	"errors"
	"reflect"
	"testing"

	"tinygo.org/x/drivers/tester"
)

type registerWrite struct {
	register uint8
	value    uint8
}

type testBus struct {
	bus    *tester.I2CBus
	calls  int
	failAt int
	err    error
	writes []registerWrite
}

func (b *testBus) Tx(addr uint16, w, r []byte) error {
	b.calls++
	if b.calls == b.failAt {
		return b.err
	}
	if len(w) == 2 && len(r) == 0 {
		b.writes = append(b.writes, registerWrite{w[0], w[1]})
	}
	return b.bus.Tx(addr, w, r)
}

func newTestDevice(t *testing.T) (Device, *tester.I2CDevice8, *testBus) {
	t.Helper()
	bus := &testBus{bus: tester.NewI2CBus(t), err: errors.New("I2C failure")}
	chip := bus.bus.NewDevice(DefaultAddress)
	chip.Registers[regProductID] = ProductID
	chip.Registers[regMultiTouchConf] = 0xF5
	chip.Registers[regSensitivity] = 0x8F
	return New(bus, DefaultAddress), chip, bus
}

func TestConfigure(t *testing.T) {
	dev, chip, bus := newTestDevice(t)
	if dev.Address != DefaultAddress {
		t.Fatalf("address = %#x, want %#x", dev.Address, DefaultAddress)
	}
	if err := dev.Configure(); err != nil {
		t.Fatal(err)
	}
	want := []registerWrite{
		{regInputEnable, 0x3F},
		{regInterruptEnable, 0x3F},
		{regRepeatEnable, 0x00},
		{regMultiTouchConf, 0x75},
		{regSamplingConfig, 0x08},
		{regSensitivity, 0xEF},
		{regGeneralConfig, 0x38},
		{regConfiguration2, 0x60},
	}
	if !reflect.DeepEqual(bus.writes, want) {
		t.Fatalf("writes = %#v, want %#v", bus.writes, want)
	}
	for _, write := range want {
		if got := chip.Registers[write.register]; got != write.value {
			t.Errorf("register %#x = %#x, want %#x", write.register, got, write.value)
		}
	}
}

func TestProductID(t *testing.T) {
	dev, chip, bus := newTestDevice(t)
	if !dev.Connected() {
		t.Fatal("device not connected")
	}
	chip.Registers[regProductID] = 0x50
	if dev.Connected() {
		t.Fatal("wrong product ID accepted")
	}
	if err := dev.Configure(); err == nil || err.Error() != "cap1166: unexpected product id 0x50" {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bus.writes) != 0 {
		t.Fatal("configuration wrote registers after wrong product ID")
	}
	bus.failAt = bus.calls + 1
	if dev.Connected() {
		t.Fatal("read error accepted")
	}
}

func TestTouchInputs(t *testing.T) {
	dev, chip, _ := newTestDevice(t)
	for mask := 0; mask < 64; mask++ {
		chip.Registers[regInputStatus] = uint8(mask) | 0xC0
		got, err := dev.InputStatus()
		if err != nil || got != uint8(mask) {
			t.Fatalf("status = %#x, %v, want %#x", got, err, mask)
		}
		for channel := uint8(0); channel < NumInputs; channel++ {
			pressed, err := dev.Pressed(channel)
			if err != nil || pressed != (mask&(1<<channel) != 0) {
				t.Fatalf("mask %#x channel %d = %v, %v", mask, channel, pressed, err)
			}
		}
	}
	for _, value := range []uint8{0x00, 0x01, 0xFE, 0xFF} {
		chip.Registers[regMainControl] = value
		if err := dev.ClearInterrupt(); err != nil {
			t.Fatal(err)
		}
		if got := chip.Registers[regMainControl]; got != value&^1 {
			t.Fatalf("main control = %#x, want %#x", got, value&^1)
		}
	}
}

func TestTouchSettings(t *testing.T) {
	dev, chip, bus := newTestDevice(t)
	for _, enable := range []bool{true, false} {
		if err := dev.EnableMultitouch(enable); err != nil {
			t.Fatal(err)
		}
		want := uint8(0xF5)
		if enable {
			want = 0x75
		}
		if got := chip.Registers[regMultiTouchConf]; got != want {
			t.Fatalf("multitouch = %#x, want %#x", got, want)
		}
	}
	for bits, multiplier := range []uint8{128, 64, 32, 16, 8, 4, 2, 1} {
		if err := dev.SetSensitivity(multiplier); err != nil {
			t.Fatal(err)
		}
		if got, want := chip.Registers[regSensitivity], uint8(0x8F)|uint8(bits<<4); got != want {
			t.Fatalf("sensitivity %d = %#x, want %#x", multiplier, got, want)
		}
	}
	for _, multiplier := range []uint8{0, 3, 255} {
		calls := bus.calls
		if err := dev.SetSensitivity(multiplier); err == nil {
			t.Fatalf("invalid sensitivity %d accepted", multiplier)
		}
		if bus.calls != calls {
			t.Fatal("invalid sensitivity used I2C")
		}
	}
	for _, mask := range []uint8{0, 0x15, 0x3F} {
		if err := dev.EnableRepeat(mask); err != nil {
			t.Fatal(err)
		}
		if got := chip.Registers[regRepeatEnable]; got != mask {
			t.Fatalf("repeat = %#x, want %#x", got, mask)
		}
	}
}

func TestGraph(t *testing.T) {
	for _, tt := range []struct {
		name       string
		percentage float32
		polarity   uint8
		state      uint8
		duty       uint8
	}{
		{"below zero", -1, 0, 0, 0},
		{"off", 0, 0, 0, 0},
		{"partial first LED", 0.125, 0, 0x20, 0xC0},
		{"quarter", 0.25, 0x20, 0x10, 0x80},
		{"half", 0.5, 0x38, 0, 0},
		{"three quarters", 0.75, 0x3C, 0x02, 0x80},
		{"on", 1, 0x3F, 0, 0},
		{"above one", 2, 0x3F, 0, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dev, _, bus := newTestDevice(t)
			if err := dev.SetGraph(tt.percentage); err != nil {
				t.Fatal(err)
			}
			want := []registerWrite{
				{regLEDDirectRamp, 0},
				{regLEDBehaviour1, 0},
				{regLEDBehaviour2, 0},
				{regLEDDirectDuty, tt.duty},
				{regLEDPolarity, tt.polarity},
				{regLEDOutputCon, tt.state},
			}
			if !reflect.DeepEqual(bus.writes, want) {
				t.Fatalf("writes = %#v, want %#v", bus.writes, want)
			}
			bus.writes = nil
			if err := dev.GraphOff(); err != nil {
				t.Fatal(err)
			}
			want = []registerWrite{{regLEDPolarity, 0}, {regLEDOutputCon, 0}}
			if !reflect.DeepEqual(bus.writes, want) {
				t.Fatalf("off writes = %#v, want %#v", bus.writes, want)
			}
		})
	}
}

func TestTransactionErrors(t *testing.T) {
	for _, tt := range []struct {
		name  string
		calls int
		run   func(*Device) error
	}{
		{"configure", 11, (*Device).Configure},
		{"input status", 1, func(d *Device) error { _, err := d.InputStatus(); return err }},
		{"pressed", 1, func(d *Device) error { _, err := d.Pressed(Up); return err }},
		{"clear interrupt", 2, (*Device).ClearInterrupt},
		{"repeat", 1, func(d *Device) error { return d.EnableRepeat(0x3F) }},
		{"multitouch", 2, func(d *Device) error { return d.EnableMultitouch(true) }},
		{"sensitivity", 2, func(d *Device) error { return d.SetSensitivity(2) }},
		{"graph", 6, func(d *Device) error { return d.SetGraph(0.25) }},
		{"graph off", 2, (*Device).GraphOff},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dev, _, bus := newTestDevice(t)
			if err := tt.run(&dev); err != nil {
				t.Fatal(err)
			}
			if bus.calls != tt.calls {
				t.Fatalf("calls = %d, want %d", bus.calls, tt.calls)
			}
			for failAt := 1; failAt <= tt.calls; failAt++ {
				dev, _, bus := newTestDevice(t)
				bus.failAt = failAt
				if err := tt.run(&dev); !errors.Is(err, bus.err) {
					t.Fatalf("transaction %d error = %v, want %v", failAt, err, bus.err)
				}
				if bus.calls != failAt {
					t.Fatalf("calls after error = %d, want %d", bus.calls, failAt)
				}
			}
		})
	}
}
