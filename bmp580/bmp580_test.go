package bmp580

import (
	"errors"
	"testing"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/tester"
)

// fakeBMP580 behaves like the sensor where the driver depends on it: a soft
// reset reports power-on complete, and a forced measurement fills the data
// registers and goes back to standby after polls reads of ODR_CONFIG.
type fakeBMP580 struct {
	*tester.I2CDevice8
	data      [6]byte
	polls     int
	pending   int
	count     int
	nackReset bool
}

func newFake(t *testing.T) (*tester.I2CBus, *fakeBMP580) {
	bus := tester.NewI2CBus(t)
	dev := &fakeBMP580{I2CDevice8: tester.NewI2CDevice8(t, Address), polls: 3}
	dev.Registers[regChipID] = chipIDBMP580
	dev.Registers[regStatus] = statusNVMReady | 1
	dev.Registers[regODRConfig] = 0x70
	bus.AddDevice(dev)
	return bus, dev
}

func (f *fakeBMP580) Tx(w, r []byte) error {
	if len(w) == 1 && w[0] == regODRConfig && f.pending > 0 {
		if f.pending--; f.pending == 0 {
			f.Registers[regODRConfig] &^= pwrModeMask
			copy(f.Registers[regTempData:], f.data[:])
			f.count++
		}
	}
	if err := f.I2CDevice8.Tx(w, r); err != nil {
		return err
	}
	switch {
	case len(w) == 1 && w[0] == regIntStatus:
		f.Registers[regIntStatus] = 0
	case len(w) == 2 && w[0] == regCmd && w[1] == cmdSoftReset:
		f.Registers[regIntStatus] |= intStatusPOR
		if f.nackReset {
			return errors.New("nack")
		}
	case len(w) == 2 && w[0] == regODRConfig && w[1]&pwrModeMask == pwrModeForced:
		f.pending = f.polls
	}
	return nil
}

func TestConnected(t *testing.T) {
	for _, c := range []struct {
		id   uint8
		want bool
	}{{chipIDBMP580, true}{0x60, false}} {
		bus, dev := newFake(t)
		dev.Registers[regChipID] = c.id
		d := New(bus)
		if got := d.Connected(); got != c.want {
			t.Errorf("chip ID %#x: Connected() = %v, want %v", c.id, got, c.want)
		}
	}
}

func TestConfigure(t *testing.T) {
	bus, dev := newFake(t)
	d := New(bus)
	if err := d.Configure(Config{Pressure: Sampling16X, Temperature: Sampling2X}); err != nil {
		t.Fatal(err)
	}
	if got, want := dev.Registers[regOSRConfig], uint8(osrPressureEnable|4<<3|1); got != want {
		t.Errorf("OSR_CONFIG = %#x, want %#x", got, want)
	}
	if got, want := dev.Registers[regODRConfig], uint8(odrDeepDisable|odrReset1Hz|pwrModeStandby); got != want {
		t.Errorf("ODR_CONFIG = %#x, want %#x", got, want)
	}
}

// The soft reset is not ACKed over I2C, datasheet section 7.36.
func TestConfigureResetNotAcked(t *testing.T) {
	bus, dev := newFake(t)
	dev.nackReset = true
	d := New(bus)
	if err := d.Configure(Config{}); err != nil {
		t.Errorf("Configure() = %v, want nil", err)
	}
}

func TestConfigureNotReady(t *testing.T) {
	bus, dev := newFake(t)
	dev.Registers[regStatus] |= statusNVMError
	d := New(bus)
	if err := d.Configure(Config{}); err != ErrNotReady {
		t.Errorf("Configure() = %v, want ErrNotReady", err)
	}
}

func TestUpdate(t *testing.T) {
	bus, dev := newFake(t)
	d := New(bus)
	if err := d.Configure(Config{}); err != nil {
		t.Fatal(err)
	}
	// 21.5 °C is 21.5*65536 = 0x158000; 101325 Pa is 101325*64 = 0x62F340.
	dev.data = [6]byte{0x00, 0x80, 0x15, 0x40, 0xF3, 0x62}
	if err := d.Update(drivers.Temperature | drivers.Pressure); err != nil {
		t.Fatal(err)
	}
	if d.Temperature() != 21_500 || d.Pressure() != 101_325_000 {
		t.Errorf("Temperature(), Pressure() = %d, %d, want 21500, 101325000", d.Temperature(), d.Pressure())
	}
	if dev.count != 1 {
		t.Errorf("took %d measurements, want 1", dev.count)
	}
	// Nothing to measure.
	if err := d.Update(drivers.Humidity); err != nil || dev.count != 1 {
		t.Errorf("Update(Humidity) = %v and took a measurement", err)
	}
}

func TestReadTemperatureAndPressure(t *testing.T) {
	bus, dev := newFake(t)
	d := New(bus)
	if err := d.Configure(Config{}); err != nil {
		t.Fatal(err)
	}
	// -10.25 °C is the two's complement of 10.25*65536 = 0xA4000.
	dev.data = [6]byte{0x00, 0xC0, 0xF5, 0x40, 0xF3, 0x62}
	if temp, err := d.ReadTemperature(); err != nil || temp != -10_250 {
		t.Errorf("ReadTemperature() = %d, %v, want -10250", temp, err)
	}
	if press, err := d.ReadPressure(); err != nil || press != 101_325_000 {
		t.Errorf("ReadPressure() = %d, %v, want 101325000", press, err)
	}
}

func TestConvert(t *testing.T) {
	for _, c := range []struct {
		data  [6]byte
		temp  int32
		press int32
	}{
		{[6]byte{0, 0, 0, 0, 0, 0}, 0, 0},
		{[6]byte{0x00, 0x80, 0x15, 0x40, 0xF3, 0x62}, 21_500, 101_325_000},
		{[6]byte{0x00, 0xC0, 0xF5, 0, 0, 0}, -10_250, 0},
		// The largest pressure the registers hold: just under 262144 Pa.
		{[6]byte{0, 0, 0, 0xFF, 0xFF, 0xFF}, 0, 262_143_984},
	} {
		temp, press := convert(c.data[:])
		if temp != c.temp || press != c.press {
			t.Errorf("convert(% x) = %d, %d, want %d, %d", c.data, temp, press, c.temp, c.press)
		}
	}
}

func TestTimeout(t *testing.T) {
	bus, dev := newFake(t)
	d := New(bus)
	if err := d.Configure(Config{}); err != nil {
		t.Fatal(err)
	}
	dev.polls = 1 << 30
	if err := d.Update(drivers.Pressure); err != ErrTimeout {
		t.Errorf("Update() = %v, want ErrTimeout", err)
	}
}
