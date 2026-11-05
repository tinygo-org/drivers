package epd2in7gray

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"image/color"
	"strings"
	"testing"
	"time"

	"tinygo.org/x/drivers"
)

type fakeOutputPin struct {
	high    bool
	changes []bool
}

func (p *fakeOutputPin) Set(high bool) {
	p.high = high
	p.changes = append(p.changes, high)
}

type fakeInputPin struct {
	high   bool
	levels []bool
	reads  int
	bus    *fakeSPI
}

func (p *fakeInputPin) Get() bool {
	p.bus.t.Helper()
	last := p.bus.operations[len(p.bus.operations)-1]
	if !last.command || !bytes.Equal(last.data, []byte{0x71}) {
		p.bus.t.Fatal("BUSY was read without GET_STATUS")
	}
	p.reads++
	if len(p.levels) != 0 {
		high := p.levels[0]
		p.levels = p.levels[1:]
		return high
	}
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
	transfers  int
	failByte   int
	failErr    error
}

func (s *fakeSPI) Transfer(value byte) (byte, error) {
	s.t.Helper()
	if s.cs.high {
		s.t.Fatal("CS is high during SPI transfer")
	}
	if len(s.cs.changes) > 1 && !s.cs.changes[len(s.cs.changes)-2] {
		s.t.Fatal("CS was not released between SPI bytes")
	}
	return 0, s.record(spiOperation{command: !s.dc.high, data: []byte{value}})
}

func (s *fakeSPI) Tx(write, read []byte) error {
	s.t.Helper()
	s.t.Fatal("bulk SPI transfers do not match the vendor byte framing")
	return nil
}

func (s *fakeSPI) record(operation spiOperation) error {
	s.transfers++
	if !operation.command && len(s.operations) > 0 && !s.operations[len(s.operations)-1].command {
		last := &s.operations[len(s.operations)-1]
		last.data = append(last.data, operation.data...)
	} else {
		s.operations = append(s.operations, operation)
	}
	if len(s.operations) == s.failAt || s.transfers == s.failByte {
		return s.failErr
	}
	return nil
}

func newTestDevice(t *testing.T) (*Device, *fakeSPI, *fakeOutputPin, *fakeInputPin) {
	t.Helper()
	cs := &fakeOutputPin{high: true}
	dc := &fakeOutputPin{high: true}
	rst := &fakeOutputPin{high: true}
	bus := &fakeSPI{t: t, cs: cs, dc: dc}
	busy := &fakeInputPin{high: true, bus: bus}
	return New(bus, cs, dc, rst, busy), bus, rst, busy
}

func registerWrites(t *testing.T, operations []spiOperation) []registerSetting {
	t.Helper()
	var writes []registerSetting
	for _, operation := range operations {
		if operation.command {
			writes = append(writes, registerSetting{command: operation.data[0]})
		} else {
			if len(writes) == 0 {
				t.Fatal("data before first command")
			}
			last := &writes[len(writes)-1]
			last.data = append(last.data, operation.data...)
		}
	}
	return writes
}

func assertWrites(t *testing.T, operations []spiOperation, want []registerSetting) {
	t.Helper()
	got := registerWrites(t, operations)
	if len(got) != len(want) {
		t.Fatalf("got %d register writes, want %d\n%#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].command != want[i].command || !bytes.Equal(got[i].data, want[i].data) {
			t.Fatalf("write %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestNew(t *testing.T) {
	d, bus, rst, busy := newTestDevice(t)
	if w, h := d.Size(); w != 176 || h != 264 {
		t.Fatalf("size = %dx%d, want 176x264", w, h)
	}
	if len(d.buffer) != 11616 {
		t.Fatalf("framebuffer length = %d, want 11616", len(d.buffer))
	}
	if !bytes.Equal(d.buffer, bytes.Repeat([]byte{0xff}, 11616)) {
		t.Fatal("framebuffer is not white")
	}
	if len(bus.operations)+len(bus.cs.changes)+len(bus.dc.changes)+len(rst.changes)+busy.reads != 0 {
		t.Fatal("New performed hardware I/O")
	}
}

func TestSetPixelPacking(t *testing.T) {
	d, bus, _, _ := newTestDevice(t)
	for x := int16(0); x < 8; x++ {
		shade := uint8(x%4) * 85
		d.SetPixel(x, 0, color.RGBA{shade, shade, shade, 255})
	}
	if !bytes.Equal(d.buffer[:3], []byte{0x1b, 0x1b, 0xff}) {
		t.Fatalf("first bytes = %x, want 1b1bff", d.buffer[:3])
	}
	d.SetPixel(3, 0, color.RGBA{85, 85, 85, 255})
	if d.buffer[0] != 0x19 {
		t.Fatalf("overwritten byte = %x, want 19", d.buffer[0])
	}
	d.SetPixel(175, 0, color.RGBA{})
	d.SetPixel(0, 1, color.RGBA{})
	d.SetPixel(175, 263, color.RGBA{})
	if d.buffer[43] != 0xfc || d.buffer[44] != 0x3f || d.buffer[11615] != 0xfc {
		t.Fatal("row or framebuffer boundary pixel was misplaced")
	}
	if len(bus.operations) != 0 {
		t.Fatal("SetPixel performed I/O")
	}
	d.ClearBuffer()
	if !bytes.Equal(d.buffer, bytes.Repeat([]byte{0xff}, 11616)) {
		t.Fatal("ClearBuffer left non-white pixels")
	}
}

func TestSetPixelLuminance(t *testing.T) {
	d, _, _, _ := newTestDevice(t)
	for value := 0; value < 256; value++ {
		c := uint8(value)
		d.SetPixel(0, 0, color.RGBA{c, c, c, 255})
		if got := d.buffer[0] >> 6; got != c/64 {
			t.Fatalf("gray %d = shade %d, want %d", c, got, c/64)
		}
	}
	for _, tc := range []struct {
		color color.RGBA
		shade byte
	}{
		{color.RGBA{255, 0, 0, 255}, 1},
		{color.RGBA{0, 255, 0, 255}, 2},
		{color.RGBA{0, 0, 255, 255}, 0},
		{color.RGBA{85, 85, 85, 0}, 1},
		{color.RGBA{170, 170, 170, 0}, 2},
	} {
		d.SetPixel(0, 0, tc.color)
		if got := d.buffer[0] >> 6; got != tc.shade {
			t.Fatalf("color %v = shade %d, want %d", tc.color, got, tc.shade)
		}
	}
}

func TestRotationAndClipping(t *testing.T) {
	for _, tc := range []struct {
		rotation drivers.Rotation
		w, h     int16
		corners  [4][2]int16
		point    [2]int16
	}{
		{drivers.Rotation0, 176, 264, [4][2]int16{{0, 0}, {175, 0}, {0, 263}, {175, 263}}, [2]int16{3, 5}},
		{drivers.Rotation90, 264, 176, [4][2]int16{{175, 0}, {175, 263}, {0, 0}, {0, 263}}, [2]int16{170, 3}},
		{drivers.Rotation180, 176, 264, [4][2]int16{{175, 263}, {0, 263}, {175, 0}, {0, 0}}, [2]int16{172, 258}},
		{drivers.Rotation270, 264, 176, [4][2]int16{{0, 263}, {0, 0}, {175, 263}, {175, 0}}, [2]int16{5, 260}},
	} {
		t.Run(fmt.Sprint(tc.rotation), func(t *testing.T) {
			d, _, _, _ := newTestDevice(t)
			before := bytes.Clone(d.buffer)
			if err := d.SetRotation(tc.rotation); err != nil {
				t.Fatal(err)
			}
			if d.Rotation() != tc.rotation || !bytes.Equal(d.buffer, before) {
				t.Fatal("SetRotation changed pixels or did not retain rotation")
			}
			if w, h := d.Size(); w != tc.w || h != tc.h {
				t.Fatalf("size = %dx%d, want %dx%d", w, h, tc.w, tc.h)
			}
			points := [][2]int16{{0, 0}, {tc.w - 1, 0}, {0, tc.h - 1}, {tc.w - 1, tc.h - 1}, {3, 5}}
			expected := append(tc.corners[:], tc.point)
			for i, point := range points {
				d.SetPixel(point[0], point[1], color.RGBA{})
				x, y := expected[i][0], expected[i][1]
				index := int(y)*44 + int(x)/4
				before[index] &^= 3 << uint(6-x%4*2)
			}
			if !bytes.Equal(d.buffer, before) {
				t.Fatal("rotated pixels do not match expected physical coordinates")
			}
			for _, point := range [][2]int16{
				{-1, 0}, {0, -1}, {tc.w, 0}, {0, tc.h},
				{-32768, 1}, {1, -32768}, {32767, 1}, {1, 32767},
			} {
				d.SetPixel(point[0], point[1], color.RGBA{})
			}
			if !bytes.Equal(d.buffer, before) {
				t.Fatal("out-of-bounds pixel modified framebuffer")
			}
			if err := d.SetRotation(drivers.Rotation0Mirror); !errors.Is(err, ErrInvalidRotation) {
				t.Fatalf("invalid rotation error = %v", err)
			}
			if d.Rotation() != tc.rotation {
				t.Fatal("invalid rotation changed the drawing orientation")
			}
		})
	}
}

func TestConfigure(t *testing.T) {
	d, bus, rst, _ := newTestDevice(t)
	var stages []string
	d.SetPixel(0, 0, color.RGBA{})
	if err := d.Configure(Config{OnInitStage: func(stage string) {
		stages = append(stages, stage)
	}}); err != nil {
		t.Fatal(err)
	}
	if d.busyTimeout != 30*time.Second || d.logicalWidth != 176 {
		t.Fatal("incorrect default timeout or framebuffer stride")
	}
	if !bytes.Equal(d.buffer, bytes.Repeat([]byte{0xff}, 11616)) {
		t.Fatal("Configure did not clear the framebuffer")
	}
	if fmt.Sprint(rst.changes) != "[true false true]" {
		t.Fatalf("reset transitions = %v", rst.changes)
	}
	wantStages := []string{
		"reset", "power setting", "booster soft start",
		"power optimization 0x60", "power optimization 0x89",
		"power optimization 0x90", "power optimization 0x93",
		"power optimization 0xa0", "power optimization 0xa1",
		"power optimization 0x73", "partial refresh disabled", "power on",
		"panel setting 0xbf", "PLL setting 0x90", "resolution",
		"VCOM DC setting 0x12", "data interval 0x97",
	}
	if strings.Join(stages, "\n") != strings.Join(wantStages, "\n") {
		t.Fatalf("init stages = %v, want %v", stages, wantStages)
	}
	assertWrites(t, bus.operations, []registerSetting{
		{0x01, []byte{0x03, 0x00, 0x2b, 0x2b}},
		{0x06, []byte{0x07, 0x07, 0x17}},
		{0xf8, []byte{0x60, 0xa5}},
		{0xf8, []byte{0x89, 0xa5}},
		{0xf8, []byte{0x90, 0x00}},
		{0xf8, []byte{0x93, 0x2a}},
		{0xf8, []byte{0xa0, 0xa5}},
		{0xf8, []byte{0xa1, 0x00}},
		{0xf8, []byte{0x73, 0x41}},
		{0x16, []byte{0x00}},
		{0x04, nil},
		{0x71, nil},
		{0x00, []byte{0xbf}},
		{0x30, []byte{0x90}},
		{0x61, []byte{0x00, 0xb0, 0x01, 0x08}},
		{0x82, []byte{0x12}},
		{0x50, []byte{0x97}},
	})
}

func TestConfigureRejectsInvalidConfigBeforeIO(t *testing.T) {
	for _, tc := range []struct {
		cfg Config
		err error
	}{
		{Config{Width: -8}, ErrInvalidDimensions},
		{Config{Width: 184}, ErrInvalidDimensions},
		{Config{Width: 175}, ErrInvalidDimensions},
		{Config{Height: -1}, ErrInvalidDimensions},
		{Config{Height: 265}, ErrInvalidDimensions},
		{Config{LogicalWidth: -8}, ErrInvalidDimensions},
		{Config{LogicalWidth: 168}, ErrInvalidDimensions},
		{Config{LogicalWidth: 177}, ErrInvalidDimensions},
		{Config{Rotation: drivers.Rotation0Mirror}, ErrInvalidRotation},
		{Config{Rotation: 255}, ErrInvalidRotation},
		{Config{BusyTimeout: -time.Second}, ErrInvalidBusyTimeout},
	} {
		t.Run(fmt.Sprint(tc.cfg), func(t *testing.T) {
			d, bus, rst, _ := newTestDevice(t)
			before := &d.buffer[0]
			if err := d.Configure(tc.cfg); !errors.Is(err, tc.err) {
				t.Fatalf("error = %v, want %v", err, tc.err)
			}
			if len(bus.operations)+len(rst.changes)+len(bus.cs.changes) != 0 {
				t.Fatal("invalid configuration performed I/O")
			}
			if &d.buffer[0] != before || d.width != 176 || d.height != 264 {
				t.Fatal("invalid configuration changed driver state")
			}
		})
	}
}

func TestDisplayBitplanesAndLUT(t *testing.T) {
	d, bus, _, _ := newTestDevice(t)
	for x := int16(0); x < 8; x++ {
		c := uint8(x%4) * 85
		d.SetPixel(x, 0, color.RGBA{c, c, c, 255})
	}
	d.SetPixel(175, 263, color.RGBA{170, 170, 170, 255})
	if err := d.Display(); err != nil {
		t.Fatal(err)
	}
	high := bytes.Repeat([]byte{0xff}, 5808)
	low := bytes.Clone(high)
	high[0] = 0x33
	low[0] = 0x55
	low[5807] = 0xfe
	assertWrites(t, bus.operations, expectedRefresh(t, high, low))
}

func expectedRefresh(t *testing.T, high, low []byte) []registerSetting {
	t.Helper()
	decode := func(s string) []byte {
		t.Helper()
		data, err := hex.DecodeString(strings.ReplaceAll(s, " ", ""))
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	zeros := strings.Repeat("00", 18)
	ww := decode("400a00000001 901414000001 10140a000001 a01301000001 " + zeros)
	return []registerSetting{
		{0x10, high},
		{0x13, low},
		{0x20, decode("0000 000a00000001 601414000001 001400000001 00130a010001 " + zeros)},
		{0x21, ww},
		{0x22, decode("400a00000001 901414000001 00140a000001 990c01030401 " + zeros)},
		{0x23, decode("400a00000001 901414000001 00140a000001 990b04040101 " + zeros)},
		{0x24, decode("800a00000001 901414000001 20140a000001 501301000001 " + zeros)},
		{0x25, ww},
		{0x12, nil},
		{0x71, nil},
	}
}

func TestPackPlaneExhaustive(t *testing.T) {
	for packed := 0; packed < 65536; packed++ {
		var high, low byte
		for p := 0; p < 8; p++ {
			shade := (packed >> uint(14-2*p)) & 3
			if shade >= 2 {
				high |= 0x80 >> uint(p)
			}
			if shade%2 == 1 {
				low |= 0x80 >> uint(p)
			}
		}
		first, second := byte(packed>>8), byte(packed)
		if got := packPlane(first, second, 1); got != high {
			t.Fatalf("high plane of %04x = %02x, want %02x", packed, got, high)
		}
		if got := packPlane(first, second, 0); got != low {
			t.Fatalf("low plane of %04x = %02x, want %02x", packed, got, low)
		}
	}
}

func TestLogicalStride(t *testing.T) {
	d, bus, _, _ := newTestDevice(t)
	if err := d.Configure(Config{
		Width: 16, Height: 3, LogicalWidth: 24,
		Rotation: drivers.Rotation270, BusyTimeout: 7 * time.Second,
	}); err != nil {
		t.Fatal(err)
	}
	if len(d.buffer) != 18 || d.busyTimeout != 7*time.Second {
		t.Fatal("incorrect custom framebuffer length or timeout")
	}
	if w, h := d.Size(); w != 3 || h != 16 {
		t.Fatalf("padded, rotated size = %dx%d, want 3x16", w, h)
	}
	writes := registerWrites(t, bus.operations)
	if !bytes.Equal(writes[14].data, []byte{0x00, 0x10, 0x00, 0x03}) {
		t.Fatalf("resolution setting = %x", writes[14].data)
	}
	d.SetPixel(2, 15, color.RGBA{})
	if d.buffer[3] != 0xfc {
		t.Fatal("rotated pixel did not use visible width and logical stride")
	}
	if err := d.SetRotation(drivers.Rotation0); err != nil {
		t.Fatal(err)
	}
	for y := int16(0); y < 3; y++ {
		for x := int16(0); x < 16; x++ {
			c := uint8(255)
			if y == 0 {
				c = uint8(x%4) * 85
			} else if y == 2 {
				c = 0
			}
			d.SetPixel(x, y, color.RGBA{c, c, c, 255})
		}
		d.buffer[int(y)*6+4] = 0
		d.buffer[int(y)*6+5] = 0
	}
	before := bytes.Clone(d.buffer)
	d.SetPixel(16, 0, color.RGBA{})
	if !bytes.Equal(before, d.buffer) {
		t.Fatal("padding was exposed as a visible pixel")
	}
	bus.operations = nil
	if err := d.Display(); err != nil {
		t.Fatal(err)
	}
	assertWrites(t, bus.operations, expectedRefresh(t,
		[]byte{0x33, 0x33, 0xff, 0xff, 0x00, 0x00},
		[]byte{0x55, 0x55, 0xff, 0xff, 0x00, 0x00}))
}

func TestPlaneExactlyFillsScratchBuffer(t *testing.T) {
	d, bus, _, _ := newTestDevice(t)
	if err := d.Configure(Config{Width: 8, Height: 64}); err != nil {
		t.Fatal(err)
	}
	bus.operations = nil
	bus.transfers = 0
	if err := d.sendPlane(0x10, 1, false); err != nil {
		t.Fatal(err)
	}
	if len(bus.operations) != 2 {
		t.Fatalf("operation count = %d, want command and one data group", len(bus.operations))
	}
	if bus.transfers != 65 {
		t.Fatalf("SPI byte count = %d, want 65", bus.transfers)
	}
	assertWrites(t, bus.operations, []registerSetting{
		{0x10, bytes.Repeat([]byte{0xff}, 64)},
	})
}

func TestReconfigureAfterSleep(t *testing.T) {
	d, bus, rst, _ := newTestDevice(t)
	if err := d.Configure(Config{Width: 8, Height: 64, Rotation: drivers.Rotation90}); err != nil {
		t.Fatal(err)
	}
	d.SetPixel(0, 0, color.RGBA{})
	if err := d.DeepSleep(); err != nil {
		t.Fatal(err)
	}
	bus.operations = nil
	rst.changes = nil
	if err := d.Configure(Config{}); err != nil {
		t.Fatal(err)
	}
	if w, h := d.Size(); w != 176 || h != 264 {
		t.Fatalf("reconfigured size = %dx%d, want 176x264", w, h)
	}
	if d.Rotation() != drivers.Rotation0 || len(d.buffer) != 11616 {
		t.Fatal("reconfiguration did not restore default geometry")
	}
	if !bytes.Equal(d.buffer, bytes.Repeat([]byte{0xff}, 11616)) {
		t.Fatal("reconfiguration did not clear the framebuffer")
	}
	if fmt.Sprint(rst.changes) != "[true false true]" {
		t.Fatalf("wake reset transitions = %v", rst.changes)
	}
	writes := registerWrites(t, bus.operations)
	if writes[0].command != 0x01 || writes[len(writes)-1].command != 0x50 {
		t.Fatal("wake did not replay controller configuration")
	}
	bus.operations = nil
	if err := d.Display(); err != nil {
		t.Fatal(err)
	}
	white := bytes.Repeat([]byte{0xff}, 5808)
	assertWrites(t, bus.operations, expectedRefresh(t, white, white))
}

func TestClearDisplayPreservesBuffer(t *testing.T) {
	for _, afterImage := range []bool{false, true} {
		t.Run(fmt.Sprint(afterImage), func(t *testing.T) {
			d, bus, _, _ := newTestDevice(t)
			if err := d.Configure(Config{}); err != nil {
				t.Fatal(err)
			}
			d.SetPixel(0, 0, color.RGBA{})
			before := bytes.Clone(d.buffer)
			if afterImage {
				if err := d.Display(); err != nil {
					t.Fatal(err)
				}
			}
			bus.operations = nil
			if err := d.ClearDisplay(); err != nil {
				t.Fatal(err)
			}
			white := bytes.Repeat([]byte{0xff}, 5808)
			assertWrites(t, bus.operations, expectedRefresh(t, white, white))
			if !bytes.Equal(d.buffer, before) {
				t.Fatal("ClearDisplay modified the framebuffer")
			}
			bus.operations = nil
			if err := d.Display(); err != nil {
				t.Fatal(err)
			}
			image := bytes.Clone(white)
			image[0] = 0x7f
			assertWrites(t, bus.operations, expectedRefresh(t, image, image))
		})
	}
}

func TestBusyRelease(t *testing.T) {
	d, bus, _, busy := newTestDevice(t)
	busy.levels = []bool{false, false, true}
	if err := d.waitUntilIdle(); err != nil {
		t.Fatal(err)
	}
	if busy.reads != 3 {
		t.Fatalf("BUSY reads = %d, want 3", busy.reads)
	}
	assertWrites(t, bus.operations, []registerSetting{{0x71, nil}, {0x71, nil}, {0x71, nil}})
}

func TestBusyTimeout(t *testing.T) {
	d, bus, _, busy := newTestDevice(t)
	busy.high = false
	if err := d.Configure(Config{BusyTimeout: time.Millisecond}); !errors.Is(err, ErrBusyTimeout) {
		t.Fatalf("Configure error = %v, want busy timeout", err)
	}
	if busy.reads == 0 || !bus.cs.high {
		t.Fatal("timeout did not poll BUSY or release CS")
	}
	for _, operation := range bus.operations {
		if operation.command && operation.data[0] == 0x00 {
			t.Fatal("Configure continued after power-on timeout")
		}
	}
	for _, operation := range []struct {
		name string
		run  func() error
	}{
		{"Display", d.Display},
		{"ClearDisplay", d.ClearDisplay},
		{"DeepSleep", d.DeepSleep},
	} {
		t.Run(operation.name, func(t *testing.T) {
			if err := operation.run(); !errors.Is(err, ErrBusyTimeout) {
				t.Fatalf("error = %v, want busy timeout", err)
			}
		})
	}
}

func TestConfigureSPIErrors(t *testing.T) {
	for _, failAt := range []int{1, 2, 21, 22, 27, 28, 32} {
		t.Run(fmt.Sprint(failAt), func(t *testing.T) {
			d, bus, _, _ := newTestDevice(t)
			bus.failAt = failAt
			bus.failErr = errors.New("SPI failure")
			if err := d.Configure(Config{}); !errors.Is(err, bus.failErr) {
				t.Fatalf("error = %v, want SPI failure", err)
			}
			if !bus.cs.high || len(bus.operations) != failAt {
				t.Fatal("Configure failed to release CS or continued after SPI error")
			}
		})
	}
}

func TestDisplaySPIErrors(t *testing.T) {
	d, bus, _, _ := newTestDevice(t)
	if err := d.Display(); err != nil {
		t.Fatal(err)
	}
	count := len(bus.operations)
	for failAt := 1; failAt <= count; failAt++ {
		t.Run(fmt.Sprint(failAt), func(t *testing.T) {
			d, bus, _, _ := newTestDevice(t)
			bus.failAt = failAt
			bus.failErr = errors.New("SPI failure")
			if err := d.Display(); !errors.Is(err, bus.failErr) {
				t.Fatalf("error = %v, want SPI failure", err)
			}
			if !bus.cs.high || len(bus.operations) != failAt {
				t.Fatal("Display failed to release CS or continued after SPI error")
			}
		})
	}
}

func TestPlaneSPIErrorWithinData(t *testing.T) {
	for _, failByte := range []int{2, 33, 65, 66, 5809} {
		t.Run(fmt.Sprint(failByte), func(t *testing.T) {
			d, bus, _, _ := newTestDevice(t)
			bus.failByte = failByte
			bus.failErr = errors.New("SPI failure")
			if err := d.sendPlane(0x10, 1, false); !errors.Is(err, bus.failErr) {
				t.Fatalf("error = %v, want SPI failure", err)
			}
			if !bus.cs.high || bus.transfers != failByte {
				t.Fatal("plane transfer failed to release CS or continued after SPI error")
			}
		})
	}
}

func TestDeepSleep(t *testing.T) {
	d, bus, _, _ := newTestDevice(t)
	if err := d.DeepSleep(); err != nil {
		t.Fatal(err)
	}
	assertWrites(t, bus.operations, []registerSetting{
		{0x50, []byte{0xf7}},
		{0x02, nil},
		{0x71, nil},
		{0x07, []byte{0xa5}},
	})
	count := len(bus.operations)
	for failAt := 1; failAt <= count; failAt++ {
		t.Run(fmt.Sprint(failAt), func(t *testing.T) {
			d, bus, _, _ := newTestDevice(t)
			bus.failAt = failAt
			bus.failErr = errors.New("SPI failure")
			if err := d.DeepSleep(); !errors.Is(err, bus.failErr) {
				t.Fatalf("error = %v, want SPI failure", err)
			}
			if !bus.cs.high || len(bus.operations) != failAt {
				t.Fatal("DeepSleep failed to release CS or continued after SPI error")
			}
		})
	}
}
