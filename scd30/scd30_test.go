package scd30

import (
	"bytes"
	"errors"
	"math"
	"testing"

	"tinygo.org/x/drivers"
)

type transaction struct {
	write    []byte
	response []byte
	err      error
}

type fakeBus struct {
	t             *testing.T
	transactions  []transaction
	next          int
	wrongAddress  bool
	extraTransfer bool
}

func (b *fakeBus) Tx(address uint16, write, read []byte) error {
	b.t.Helper()
	if address != Address {
		b.wrongAddress = true
		b.t.Errorf("address = %#x, want %#x", address, Address)
		return nil
	}
	if b.next >= len(b.transactions) {
		b.extraTransfer = true
		b.t.Errorf("unexpected transaction: write=%#v read-len=%d", write, len(read))
		return nil
	}

	want := b.transactions[b.next]
	b.next++
	if !bytes.Equal(write, want.write) {
		b.t.Errorf("write = %#v, want %#v", write, want.write)
	}
	if len(read) != len(want.response) {
		b.t.Errorf("read length = %d, want %d", len(read), len(want.response))
	}
	if want.err != nil {
		return want.err
	}
	copy(read, want.response)
	return nil
}

func (b *fakeBus) verify(t *testing.T) {
	t.Helper()
	if b.next != len(b.transactions) {
		t.Errorf("completed %d transactions, want %d", b.next, len(b.transactions))
	}
	if b.wrongAddress {
		t.Error("driver used an unexpected address")
	}
	if b.extraTransfer {
		t.Error("driver performed an unexpected transfer")
	}
}

func newFakeBus(t *testing.T, transactions ...transaction) *fakeBus {
	t.Helper()
	bus := &fakeBus{t: t, transactions: transactions}
	t.Cleanup(func() { bus.verify(t) })
	return bus
}

func TestConfigure(t *testing.T) {
	bus := newFakeBus(t,
		transaction{write: []byte{0x46, 0x00, 0x00, 0x0a, 0x5a}},
		transaction{write: []byte{0x53, 0x06, 0x00, 0x01, 0xb0}},
	)
	err := New(bus).Configure(Config{
		MeasurementInterval:      10,
		AutomaticSelfCalibration: true,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDefaultConfig(t *testing.T) {
	if DefaultConfig.MeasurementInterval != 2 {
		t.Errorf("default interval = %d, want 2", DefaultConfig.MeasurementInterval)
	}
	if DefaultConfig.AutomaticSelfCalibration {
		t.Error("automatic self-calibration should be disabled by default")
	}
}

func TestMeasurementIntervalBounds(t *testing.T) {
	tests := []struct {
		name    string
		seconds uint16
		valid   bool
	}{
		{name: "below minimum", seconds: 1},
		{name: "minimum", seconds: 2, valid: true},
		{name: "maximum", seconds: 1800, valid: true},
		{name: "above maximum", seconds: 1801},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var transactions []transaction
			if test.valid {
				argument := []byte{byte(test.seconds >> 8), byte(test.seconds)}
				transactions = append(transactions, transaction{write: []byte{
					0x46, 0x00, argument[0], argument[1], crc8(argument),
				}})
			}
			bus := newFakeBus(t, transactions...)
			err := New(bus).SetMeasurementInterval(test.seconds)
			if test.valid && err != nil {
				t.Fatal(err)
			}
			if !test.valid && !errors.Is(err, ErrInvalidInterval) {
				t.Fatalf("error = %v, want %v", err, ErrInvalidInterval)
			}
		})
	}
}

func TestAmbientPressureBounds(t *testing.T) {
	tests := []struct {
		pressure uint16
		valid    bool
	}{
		{pressure: 0, valid: true},
		{pressure: 699},
		{pressure: 700, valid: true},
		{pressure: 1400, valid: true},
		{pressure: 1401},
	}

	for _, test := range tests {
		t.Run(stringName(test.pressure), func(t *testing.T) {
			var transactions []transaction
			if test.valid {
				argument := []byte{byte(test.pressure >> 8), byte(test.pressure)}
				transactions = append(transactions, transaction{write: []byte{
					0x00, 0x10, argument[0], argument[1], crc8(argument),
				}})
			}
			bus := newFakeBus(t, transactions...)
			err := New(bus).StartContinuousMeasurement(test.pressure)
			if test.valid && err != nil {
				t.Fatal(err)
			}
			if !test.valid && !errors.Is(err, ErrInvalidAmbientPressure) {
				t.Fatalf("error = %v, want %v", err, ErrInvalidAmbientPressure)
			}
		})
	}
}

func TestStopContinuousMeasurement(t *testing.T) {
	bus := newFakeBus(t, transaction{write: []byte{0x01, 0x04}})
	if err := New(bus).StopContinuousMeasurement(); err != nil {
		t.Fatal(err)
	}
}

func TestDataReady(t *testing.T) {
	for _, test := range []struct {
		name  string
		value uint16
		ready bool
	}{
		{name: "not ready", value: 0},
		{name: "ready", value: 1, ready: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			bus := newFakeBus(t,
				transaction{write: []byte{0x02, 0x02}},
				transaction{response: encodeWord(test.value)},
			)
			ready, err := New(bus).DataReady()
			if err != nil {
				t.Fatal(err)
			}
			if ready != test.ready {
				t.Errorf("ready = %v, want %v", ready, test.ready)
			}
		})
	}
}

func TestDataReadyRejectsBadCRC(t *testing.T) {
	response := encodeWord(1)
	response[2]++
	bus := newFakeBus(t,
		transaction{write: []byte{0x02, 0x02}},
		transaction{response: response},
	)
	_, err := New(bus).DataReady()
	if !errors.Is(err, ErrCRC) {
		t.Fatalf("error = %v, want %v", err, ErrCRC)
	}
}

func TestReadMeasurement(t *testing.T) {
	response := appendFloat(nil, 800.5)
	response = appendFloat(response, 23.25)
	response = appendFloat(response, 48.75)
	bus := newFakeBus(t,
		transaction{write: []byte{0x03, 0x00}},
		transaction{response: response},
	)

	device := New(bus)
	if err := device.ReadMeasurement(); err != nil {
		t.Fatal(err)
	}
	if got := device.CO2(); got != 801 {
		t.Errorf("CO2 = %d, want 801 ppm", got)
	}
	if got := device.Temperature(); got != 23250 {
		t.Errorf("temperature = %d, want 23250 mC", got)
	}
	if got := device.Humidity(); got != 4875 {
		t.Errorf("humidity = %d, want 4875 hundredths of a percent", got)
	}
}

func TestReadMeasurementRoundsNegativeTemperature(t *testing.T) {
	response := appendFloat(nil, 400)
	response = appendFloat(response, -10.1236)
	response = appendFloat(response, 50)
	bus := newFakeBus(t,
		transaction{write: []byte{0x03, 0x00}},
		transaction{response: response},
	)

	device := New(bus)
	if err := device.ReadMeasurement(); err != nil {
		t.Fatal(err)
	}
	if got := device.Temperature(); got != -10124 {
		t.Errorf("temperature = %d, want -10124 mC", got)
	}
}

func TestReadMeasurementRejectsBadCRCWithoutChangingCache(t *testing.T) {
	for corruptWord := 0; corruptWord < 6; corruptWord++ {
		t.Run(stringName(uint16(corruptWord)), func(t *testing.T) {
			response := appendFloat(nil, 800.5)
			response = appendFloat(response, 23.25)
			response = appendFloat(response, 48.75)
			response[corruptWord*3+2]++
			bus := newFakeBus(t,
				transaction{write: []byte{0x03, 0x00}},
				transaction{response: response},
			)

			device := New(bus)
			device.co2 = 500
			device.temperature = 21000
			device.humidity = 4000
			err := device.ReadMeasurement()
			if !errors.Is(err, ErrCRC) {
				t.Fatalf("error = %v, want %v", err, ErrCRC)
			}
			if device.CO2() != 500 || device.Temperature() != 21000 || device.Humidity() != 4000 {
				t.Errorf("cache changed after CRC error: CO2=%d temperature=%d humidity=%d",
					device.CO2(), device.Temperature(), device.Humidity())
			}
		})
	}
}

func TestUpdateIgnoresUnsupportedMeasurements(t *testing.T) {
	bus := newFakeBus(t)
	if err := New(bus).Update(drivers.Pressure); err != nil {
		t.Fatal(err)
	}
}

func TestBusErrorsAreReturned(t *testing.T) {
	wantErr := errors.New("I2C failure")
	for _, test := range []struct {
		name         string
		transactions []transaction
		action       func(*Device) error
	}{
		{
			name:         "command write",
			transactions: []transaction{{write: []byte{0x02, 0x02}, err: wantErr}},
			action: func(device *Device) error {
				_, err := device.DataReady()
				return err
			},
		},
		{
			name: "response read",
			transactions: []transaction{
				{write: []byte{0x02, 0x02}},
				{response: make([]byte, 3), err: wantErr},
			},
			action: func(device *Device) error {
				_, err := device.DataReady()
				return err
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			bus := newFakeBus(t, test.transactions...)
			if err := test.action(New(bus)); !errors.Is(err, wantErr) {
				t.Fatalf("error = %v, want %v", err, wantErr)
			}
		})
	}
}

func TestCRC8(t *testing.T) {
	if got := crc8([]byte{0x00, 0x02}); got != 0xe3 {
		t.Fatalf("crc = %#x, want 0xe3", got)
	}
}

func appendFloat(destination []byte, value float32) []byte {
	bits := math.Float32bits(value)
	destination = append(destination, encodeWord(uint16(bits>>16))...)
	destination = append(destination, encodeWord(uint16(bits))...)
	return destination
}

func encodeWord(value uint16) []byte {
	result := []byte{byte(value >> 8), byte(value), 0}
	result[2] = crc8(result[:2])
	return result
}

func stringName(value uint16) string {
	const digits = "0123456789"
	if value == 0 {
		return "0"
	}
	var buffer [5]byte
	position := len(buffer)
	for value > 0 {
		position--
		buffer[position] = digits[value%10]
		value /= 10
	}
	return string(buffer[position:])
}
