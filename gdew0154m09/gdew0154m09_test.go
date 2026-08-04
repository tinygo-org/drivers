package gdew0154m09

import (
	"bytes"
	"errors"
	"image/color"
	"testing"
	"time"
)

type fakeOutputPin struct {
	high    bool
	changes []bool
}

func (p *fakeOutputPin) Set(level bool) {
	p.high = level
	p.changes = append(p.changes, level)
}

type fakeInputPin struct {
	high bool
}

func (p *fakeInputPin) Get() bool {
	return p.high
}

type spiOperation struct {
	command bool
	data    []byte
}

type fakeSPI struct {
	t          *testing.T
	cs         *fakeOutputPin
	dc         *fakeOutputPin
	operations []spiOperation
	failAt     int
	failErr    error
}

func (s *fakeSPI) Transfer(value byte) (byte, error) {
	s.t.Helper()
	if s.cs.high {
		s.t.Error("chip select was high during command transfer")
	}
	if s.dc.high {
		s.t.Error("data/command pin was high during command transfer")
	}
	s.operations = append(s.operations, spiOperation{command: true, data: []byte{value}})
	if len(s.operations) == s.failAt {
		return 0, s.failErr
	}
	return 0, nil
}

func (s *fakeSPI) Tx(write, read []byte) error {
	s.t.Helper()
	if s.cs.high {
		s.t.Error("chip select was high during data transfer")
	}
	if !s.dc.high {
		s.t.Error("data/command pin was low during data transfer")
	}
	data := append([]byte(nil), write...)
	s.operations = append(s.operations, spiOperation{data: data})
	if len(s.operations) == s.failAt {
		return s.failErr
	}
	return nil
}

func newTestDevice(t *testing.T) (*Device, *fakeSPI, *fakeOutputPin, *fakeInputPin) {
	t.Helper()
	cs := &fakeOutputPin{high: true}
	dc := &fakeOutputPin{high: true}
	rst := &fakeOutputPin{high: true}
	busy := &fakeInputPin{high: true}
	bus := &fakeSPI{t: t, cs: cs, dc: dc}
	return New(bus, cs, dc, rst, busy), bus, rst, busy
}

func TestNewStartsWithWhiteBuffers(t *testing.T) {
	device, _, _, _ := newTestDevice(t)
	for index := range device.buffer {
		if device.buffer[index] != 0xff || device.previous[index] != 0xff {
			t.Fatalf("pixel byte %d was not initialized white", index)
		}
	}
}

func TestSize(t *testing.T) {
	device, _, _, _ := newTestDevice(t)
	width, height := device.Size()
	if width != Width || height != Height {
		t.Fatalf("size = %dx%d, want %dx%d", width, height, Width, Height)
	}
}

func TestSetPixel(t *testing.T) {
	device, _, _, _ := newTestDevice(t)
	black := color.RGBA{A: 0xff}
	white := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}

	device.SetPixel(0, 0, black)
	device.SetPixel(199, 199, black)
	if device.buffer[0] != 0x7f {
		t.Fatalf("first byte = %#x, want 0x7f", device.buffer[0])
	}
	if device.buffer[len(device.buffer)-1] != 0xfe {
		t.Fatalf("last byte = %#x, want 0xfe", device.buffer[len(device.buffer)-1])
	}

	device.SetPixel(0, 0, white)
	if device.buffer[0] != 0xff {
		t.Fatalf("first byte after white pixel = %#x, want 0xff", device.buffer[0])
	}
}

func TestSetPixelIgnoresOutOfBoundsCoordinates(t *testing.T) {
	device, _, _, _ := newTestDevice(t)
	device.SetPixel(-1, 0, color.RGBA{A: 0xff})
	device.SetPixel(0, -1, color.RGBA{A: 0xff})
	device.SetPixel(Width, 0, color.RGBA{A: 0xff})
	device.SetPixel(0, Height, color.RGBA{A: 0xff})
	for index, value := range device.buffer {
		if value != 0xff {
			t.Fatalf("byte %d changed to %#x", index, value)
		}
	}
}

func TestConfigure(t *testing.T) {
	device, bus, rst, _ := newTestDevice(t)
	if err := device.Configure(Config{}); err != nil {
		t.Fatal(err)
	}
	if device.busyTimeout != DefaultConfig.BusyTimeout {
		t.Fatalf("busy timeout = %v, want %v", device.busyTimeout, DefaultConfig.BusyTimeout)
	}

	wantReset := []bool{true, false, true}
	if !equalBools(rst.changes, wantReset) {
		t.Fatalf("reset changes = %v, want %v", rst.changes, wantReset)
	}
	wantCommands := []byte{0x00, 0x4d, 0xaa, 0xe9, 0xb6, 0xf3, 0x61, 0x60, 0x50, 0xe3, 0x04}
	if got := commandBytes(bus.operations); !bytes.Equal(got, wantCommands) {
		t.Fatalf("commands = %#v, want %#v", got, wantCommands)
	}

	wantData := [][]byte{
		{0xdf, 0x0e}, {0x55}, {0x0f}, {0x02}, {0x11}, {0x0a},
		{0xc8, 0x00, 0xc8}, {0x00}, {0xd7}, {0x00},
	}
	if got := dataTransfers(bus.operations); !equalByteSlices(got, wantData) {
		t.Fatalf("data transfers = %#v, want %#v", got, wantData)
	}
}

func TestConfigurePropagatesSPIErrors(t *testing.T) {
	wantErr := errors.New("SPI failure")
	for _, operation := range []int{1, 2} {
		t.Run(operationName(operation), func(t *testing.T) {
			device, bus, _, _ := newTestDevice(t)
			bus.failAt = operation
			bus.failErr = wantErr
			if err := device.Configure(DefaultConfig); !errors.Is(err, wantErr) {
				t.Fatalf("error = %v, want %v", err, wantErr)
			}
			if !bus.cs.high {
				t.Error("chip select was not released after SPI error")
			}
		})
	}
}

func TestConfigureRejectsNegativeTimeout(t *testing.T) {
	device, bus, rst, _ := newTestDevice(t)
	err := device.Configure(Config{BusyTimeout: -time.Millisecond})
	if !errors.Is(err, ErrInvalidBusyTimeout) {
		t.Fatalf("error = %v, want %v", err, ErrInvalidBusyTimeout)
	}
	if len(bus.operations) != 0 || len(rst.changes) != 0 {
		t.Fatal("invalid configuration performed hardware I/O")
	}
}

func TestDisplaySendsOldAndNewFrames(t *testing.T) {
	device, bus, _, _ := newTestDevice(t)
	device.SetPixel(0, 0, color.RGBA{A: 0xff})
	if err := device.Display(); err != nil {
		t.Fatal(err)
	}

	if len(bus.operations) != 5 {
		t.Fatalf("operation count = %d, want 5", len(bus.operations))
	}
	assertCommand(t, bus.operations[0], commandDataStartOld)
	if len(bus.operations[1].data) != frameBufferSize || bus.operations[1].data[0] != 0xff {
		t.Fatal("old frame was not initially white")
	}
	assertCommand(t, bus.operations[2], commandDataStartNew)
	if len(bus.operations[3].data) != frameBufferSize || bus.operations[3].data[0] != 0x7f {
		t.Fatal("new frame did not contain the black pixel")
	}
	assertCommand(t, bus.operations[4], commandDisplayRefresh)

	bus.operations = nil
	device.SetPixel(0, 0, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})
	if err := device.Display(); err != nil {
		t.Fatal(err)
	}
	if bus.operations[1].data[0] != 0x7f {
		t.Fatal("second refresh did not send the first frame as old data")
	}
}

func TestDisplayPropagatesSPIErrors(t *testing.T) {
	wantErr := errors.New("SPI failure")
	for operation := 1; operation <= 5; operation++ {
		t.Run(operationName(operation), func(t *testing.T) {
			device, bus, _, _ := newTestDevice(t)
			bus.failAt = operation
			bus.failErr = wantErr
			if err := device.Display(); !errors.Is(err, wantErr) {
				t.Fatalf("error = %v, want %v", err, wantErr)
			}
			if !bus.cs.high {
				t.Error("chip select was not released after SPI error")
			}
		})
	}
}

func TestBusyTimeout(t *testing.T) {
	device, _, _, busy := newTestDevice(t)
	busy.high = false
	device.busyTimeout = time.Millisecond
	if err := device.waitUntilIdle(); !errors.Is(err, ErrBusyTimeout) {
		t.Fatalf("error = %v, want %v", err, ErrBusyTimeout)
	}
}

func TestDeepSleep(t *testing.T) {
	device, bus, _, _ := newTestDevice(t)
	if err := device.DeepSleep(); err != nil {
		t.Fatal(err)
	}
	wantCommands := []byte{commandVCOMDataInterval, commandPowerOff, commandDeepSleep}
	if got := commandBytes(bus.operations); !bytes.Equal(got, wantCommands) {
		t.Fatalf("commands = %#v, want %#v", got, wantCommands)
	}
	wantData := [][]byte{{deepSleepVCOMDataInterval}, {deepSleepCheckCode}}
	if got := dataTransfers(bus.operations); !equalByteSlices(got, wantData) {
		t.Fatalf("data transfers = %#v, want %#v", got, wantData)
	}
}

func assertCommand(t *testing.T, operation spiOperation, command byte) {
	t.Helper()
	if !operation.command || len(operation.data) != 1 || operation.data[0] != command {
		t.Fatalf("operation = %#v, want command %#x", operation, command)
	}
}

func commandBytes(operations []spiOperation) []byte {
	var commands []byte
	for _, operation := range operations {
		if operation.command {
			commands = append(commands, operation.data[0])
		}
	}
	return commands
}

func dataTransfers(operations []spiOperation) [][]byte {
	var transfers [][]byte
	for _, operation := range operations {
		if !operation.command {
			transfers = append(transfers, operation.data)
		}
	}
	return transfers
}

func equalBools(left, right []bool) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func equalByteSlices(left, right [][]byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if !bytes.Equal(left[index], right[index]) {
			return false
		}
	}
	return true
}

func operationName(operation int) string {
	return "operation-" + string(rune('0'+operation))
}
