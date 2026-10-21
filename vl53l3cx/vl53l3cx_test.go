package vl53l3cx

import (
	"testing"

	"tinygo.org/x/drivers/tester"
)

func newTestDevice(t *testing.T) (*Device, *tester.I2CDeviceAddr16, *tester.I2CBus) {
	t.Helper()
	bus := tester.NewI2CBus(t)
	chip := tester.NewI2CDeviceAddr16(t, Address)
	bus.AddDevice(chip)

	r := &chip.Registers
	r[regIdentificationModelID] = modelID >> 8
	r[regIdentificationModelID+1] = modelID & 0xFF
	r[regFirmwareSystemStatus] = 0x01
	r[regResultOscCalibrateVal] = 0x01
	r[regResultOscCalibrateVal+1] = 0x51
	r[regI2CSlaveDeviceAddress] = Address
	r[regI2CSlaveDeviceAddress+5] = 0xBC
	r[regI2CSlaveDeviceAddress+6] = 0xCC
	for i := 15; i < 47; i++ {
		r[regIdentificationModelID+i] = 0xFF
	}
	r[regIdentificationModelID+47] = 0xC7
	r[regIdentificationModelID+48] = 0xFF
	return New(bus), chip, bus
}

func configure(t *testing.T, cfg Config) (*Device, *tester.I2CDeviceAddr16) {
	t.Helper()
	dev, chip, _ := newTestDevice(t)
	if err := dev.Configure(cfg); err != nil {
		t.Fatal(err)
	}
	return dev, chip
}

// setHistogram fills the result registers that histogramBinData reads.
func setHistogram(chip *tester.I2CDeviceAddr16, streamCount uint8, bins [histogramBufferSize]int32) {
	r := &chip.Registers
	r[regResultInterruptStatus] = 0x01
	r[regResultInterruptStatus+1] = deviceErrorRangeComplete
	r[regResultInterruptStatus+3] = streamCount
	r[regResultInterruptStatus+4] = 0x20
	r[regResultInterruptStatus+5] = 0x00
	for i, v := range bins {
		p := regHistogramBin0_2 + 3*i
		r[p] = uint8(v >> 16)
		r[p+1] = uint8(v >> 8)
		r[p+2] = uint8(v)
	}
	r[regPhasecalReferencePhase] = 0x08
	r[regPhasecalReferencePhase+1] = 0x00
	r[regPhasecalVcselStart] = 0x09
	r[regGPIOTioHVStatus] = 0x00
}

func pulseHistogram(peak int) [histogramBufferSize]int32 {
	var bins [histogramBufferSize]int32
	for i := range bins {
		bins[i] = 1000
	}
	bins[peak-1] += 20000
	bins[peak] += 40000
	bins[peak+1] += 20000
	return bins
}

func TestConnected(t *testing.T) {
	dev, chip, _ := newTestDevice(t)
	if !dev.Connected() {
		t.Fatal("expected device to be connected")
	}
	chip.Registers[regIdentificationModelID] = 0
	if dev.Connected() {
		t.Fatal("expected device to not be connected")
	}
}

func TestConfigureNotConnected(t *testing.T) {
	dev, chip, _ := newTestDevice(t)
	chip.Registers[regIdentificationModelID] = 0
	if err := dev.Configure(Config{}); err != errNotConnected {
		t.Fatalf("err = %v, want %v", err, errNotConnected)
	}
}

func TestConfigureTimeout(t *testing.T) {
	dev, chip, _ := newTestDevice(t)
	chip.Registers[regFirmwareSystemStatus] = 0
	if err := dev.Configure(Config{}); err != errTimeout {
		t.Fatalf("err = %v, want %v", err, errTimeout)
	}
}

func TestConfigure(t *testing.T) {
	dev, chip := configure(t, Config{})
	if v := chip.Registers[regSoftReset]; v != 0x01 {
		t.Errorf("soft reset = %#x, want 0x01", v)
	}
	if dev.oscCalibrateVal != 0x0151 {
		t.Errorf("osc calibrate = %#x, want 0x0151", dev.oscCalibrateVal)
	}
	if f := dev.fastOscFrequency(); f != 0xBCCC {
		t.Errorf("fast osc frequency = %#x, want 0xBCCC", f)
	}
	if dev.addOffCal.mmInnerPeakRate != 0x0080 || dev.addOffCal.mmOuterPeakRate != 0x0180 {
		t.Errorf("offset cal peak rates = %#x %#x, want default", dev.addOffCal.mmInnerPeakRate, dev.addOffCal.mmOuterPeakRate)
	}
	if dev.validPhaseHigh != 0x88 {
		t.Errorf("valid phase high = %#x, want Long preset", dev.validPhaseHigh)
	}
	assertTimingBudget(t, dev, defaultTimingBudget)
}

func TestConfigureCalibrationData(t *testing.T) {
	dev, chip, _ := newTestDevice(t)
	r := &chip.Registers
	r[regIdentificationModelID+15] = 0x0F
	r[regIdentificationModelID+47] = 0x99
	r[regIdentificationModelID+48] = 0x33
	copy(r[regNVMCtrlDataOut:], []uint8{0x00, 0x10, 0x00, 0x20})
	if err := dev.Configure(Config{}); err != nil {
		t.Fatal(err)
	}
	if dev.goodSpads[0] != 0x0F || dev.goodSpads[31] != 0xFF {
		t.Errorf("good SPADs = %#x %#x, want 0x0F 0xFF", dev.goodSpads[0], dev.goodSpads[31])
	}
	if dev.modeROICentreSpad != 0x99 || dev.modeROIXYSize != 0x33 {
		t.Errorf("mode ROI = %#x %#x, want 0x99 0x33", dev.modeROICentreSpad, dev.modeROIXYSize)
	}
	want := additionalOffsetCal{0x10, 0x20, 0x10, 0x20}
	if dev.addOffCal != want {
		t.Errorf("offset cal = %+v, want %+v", dev.addOffCal, want)
	}
}

func TestConfigureSettings(t *testing.T) {
	dev, _ := configure(t, Config{Mode: Short, TimingBudget: 100000})
	if dev.validPhaseHigh != 0x28 {
		t.Errorf("valid phase high = %#x, want Short preset", dev.validPhaseHigh)
	}
	assertTimingBudget(t, dev, 100000)
}

func TestConfigureLowOscFrequency(t *testing.T) {
	dev, chip, _ := newTestDevice(t)
	chip.Registers[regI2CSlaveDeviceAddress+5] = 0x00
	chip.Registers[regI2CSlaveDeviceAddress+6] = 0x10
	if err := dev.Configure(Config{}); err != nil {
		t.Fatal(err)
	}
	if f := dev.fastOscFrequency(); f != 0xBCCC {
		t.Errorf("fast osc frequency = %#x, want 0xBCCC", f)
	}
}

// assertTimingBudget allows for rounding in the timeout encoding.
func assertTimingBudget(t *testing.T, dev *Device, want uint32) {
	t.Helper()
	got := dev.TimingBudget()
	if got+want/50 < want || got > want+want/50 {
		t.Errorf("timing budget = %d, want about %d", got, want)
	}
}

func TestSetTimingBudget(t *testing.T) {
	dev, _ := configure(t, Config{})
	for _, us := range []uint32{0, timingGuardUs, maxTimingBudgetUs + timingGuardUs + 1} {
		if err := dev.SetTimingBudget(us); err != errInvalidParams {
			t.Errorf("SetTimingBudget(%d) err = %v, want %v", us, err, errInvalidParams)
		}
	}
	for _, us := range []uint32{20000, 50000, 200000} {
		if err := dev.SetTimingBudget(us); err != nil {
			t.Fatal(err)
		}
		assertTimingBudget(t, dev, us)
	}
}

func TestSetBeforeConfigure(t *testing.T) {
	dev, _, _ := newTestDevice(t)
	if err := dev.SetTimingBudget(50000); err != errNotConfigured {
		t.Errorf("SetTimingBudget err = %v, want %v", err, errNotConfigured)
	}
	if err := dev.SetDistanceMode(Short); err != errNotConfigured {
		t.Errorf("SetDistanceMode err = %v, want %v", err, errNotConfigured)
	}
}

func TestSetDistanceMode(t *testing.T) {
	dev, _ := configure(t, Config{TimingBudget: 50000})
	for _, mode := range []DistanceMode{0, Long + 1} {
		if err := dev.SetDistanceMode(mode); err != errInvalidParams {
			t.Errorf("SetDistanceMode(%d) err = %v, want %v", mode, err, errInvalidParams)
		}
	}
	tests := []struct {
		mode      DistanceMode
		phaseHigh uint8
		vcselA    uint8
	}{
		{Short, 0x28, 0x03},
		{Medium, 0x48, 0x05},
		{Long, 0x88, 0x09},
	}
	for _, tt := range tests {
		if err := dev.SetDistanceMode(tt.mode); err != nil {
			t.Fatal(err)
		}
		if dev.validPhaseHigh != tt.phaseHigh || dev.timingCfg.vcselPeriodA != tt.vcselA {
			t.Errorf("mode %d: phase high %#x vcsel %#x, want %#x %#x",
				tt.mode, dev.validPhaseHigh, dev.timingCfg.vcselPeriodA, tt.phaseHigh, tt.vcselA)
		}
		assertTimingBudget(t, dev, 50000)
	}
}

func TestSetAddress(t *testing.T) {
	dev, chip, bus := newTestDevice(t)
	if err := dev.SetAddress(0x30); err != nil {
		t.Fatal(err)
	}
	if v := chip.Registers[regI2CSlaveDeviceAddress]; v != 0x30 {
		t.Fatalf("address register = %#x, want 0x30", v)
	}

	moved := tester.NewI2CDeviceAddr16(t, 0x30)
	moved.Registers[regIdentificationModelID] = modelID >> 8
	moved.Registers[regIdentificationModelID+1] = modelID & 0xFF
	bus.AddDevice(moved)
	chip.Registers[regIdentificationModelID] = 0
	if !dev.Connected() {
		t.Fatal("expected device to be connected at the new address")
	}
}

func TestDataReady(t *testing.T) {
	dev, chip, _ := newTestDevice(t)
	chip.Registers[regGPIOTioHVStatus] = 0x01
	if ready, err := dev.DataReady(); err != nil || ready {
		t.Fatalf("DataReady() = %v, %v, want false", ready, err)
	}
	chip.Registers[regGPIOTioHVStatus] = 0x00
	if ready, err := dev.DataReady(); err != nil || !ready {
		t.Fatalf("DataReady() = %v, %v, want true", ready, err)
	}
}

func TestStartStopRanging(t *testing.T) {
	dev, chip := configure(t, Config{})
	if err := dev.StartRanging(); err != nil {
		t.Fatal(err)
	}
	r := &chip.Registers
	if r[regPatchCtrl] != 0x01 || r[regFirmwareEnable] != 0x01 {
		t.Errorf("patch ctrl %#x firmware enable %#x, want 0x01 0x01", r[regPatchCtrl], r[regFirmwareEnable])
	}
	if v := r[regPatchOffset0]; v != 0x29 {
		t.Errorf("patch offset = %#x, want 0x29", v)
	}
	if v := r[regI2CSlaveDeviceAddress]; v != Address {
		t.Errorf("address register = %#x, want %#x", v, Address)
	}
	if v := r[regPowerManagementGo1+4] & measurementModeMask; v != measurementModeBack2B {
		t.Errorf("mode start = %#x, want back to back", v)
	}

	if err := dev.StopRanging(); err != nil {
		t.Fatal(err)
	}
	if v := r[regPowerManagementGo1+4] & measurementModeMask; v != measurementModeAbort {
		t.Errorf("mode start = %#x, want abort", v)
	}
	if r[regPatchCtrl] != 0x00 || r[regPowerManagementGo1] != 0x00 {
		t.Errorf("patch ctrl %#x power force %#x, want 0x00 0x00", r[regPatchCtrl], r[regPowerManagementGo1])
	}
}

func TestReadTarget(t *testing.T) {
	dev, chip := configure(t, Config{})
	if err := dev.StartRanging(); err != nil {
		t.Fatal(err)
	}

	var m Measurement
	setHistogram(chip, 0, pulseHistogram(8))
	if err := dev.Read(&m); err != nil {
		t.Fatal(err)
	}
	if m.NumTargets != 1 {
		t.Fatalf("targets = %d, want 1", m.NumTargets)
	}
	near := m.Targets[0]
	if near.Status != RangeValidNoWrapCheckFail {
		t.Errorf("first frame status = %d, want %d", near.Status, RangeValidNoWrapCheckFail)
	}
	if near.Distance <= 0 || near.DistanceMin > near.Distance || near.DistanceMax < near.Distance {
		t.Errorf("distance = %d [%d, %d]", near.Distance, near.DistanceMin, near.DistanceMax)
	}
	if near.SignalRate == 0 || near.AmbientRate == 0 {
		t.Errorf("signal rate = %d ambient rate = %d, want non zero", near.SignalRate, near.AmbientRate)
	}

	// The phase consistency check starts on the third frame after start.
	// Odd frames have no ambient bins, so the target is 4 bins earlier.
	frames := []struct {
		stream uint8
		peak   int
		status RangeStatus
	}{
		{0, 8, RangeValidNoWrapCheckFail},
		{1, 4, RangeValid},
		{2, 8, RangeValid},
		{3, 4, RangeValid},
		{4, 12, WrapTargetFail},
	}
	for i, f := range frames {
		setHistogram(chip, f.stream, pulseHistogram(f.peak))
		if err := dev.Read(&m); err != nil {
			t.Fatal(err)
		}
		if m.NumTargets != 1 || m.Targets[0].Status != f.status {
			t.Fatalf("frame %d: targets = %d status = %d, want 1 target with status %d",
				i+1, m.NumTargets, m.Targets[0].Status, f.status)
		}
		if f.status == RangeValid {
			if diff := m.Targets[0].Distance - near.Distance; diff < -20 || diff > 20 {
				t.Errorf("frame %d distance = %d, want about %d", i+1, m.Targets[0].Distance, near.Distance)
			}
		}
	}

	dev, chip = configure(t, Config{})
	if err := dev.StartRanging(); err != nil {
		t.Fatal(err)
	}
	setHistogram(chip, 0, pulseHistogram(12))
	if err := dev.Read(&m); err != nil {
		t.Fatal(err)
	}
	if m.NumTargets != 1 || m.Targets[0].Distance <= near.Distance {
		t.Errorf("later pulse: targets = %d distance = %d, want 1 target beyond %d",
			m.NumTargets, m.Targets[0].Distance, near.Distance)
	}
}

func TestReadNoTarget(t *testing.T) {
	dev, chip := configure(t, Config{})
	if err := dev.StartRanging(); err != nil {
		t.Fatal(err)
	}
	var bins [histogramBufferSize]int32
	for i := range bins {
		bins[i] = 1000
	}
	setHistogram(chip, 0, bins)

	var m Measurement
	if err := dev.Read(&m); err != nil {
		t.Fatal(err)
	}
	if m.NumTargets != 0 || m.Targets[0].Status != NoTarget || m.Targets[0].Distance != 8191 {
		t.Errorf("targets = %d status = %d distance = %d, want no target",
			m.NumTargets, m.Targets[0].Status, m.Targets[0].Distance)
	}
}

func TestReadNoisyFrame(t *testing.T) {
	dev, chip := configure(t, Config{})
	if err := dev.StartRanging(); err != nil {
		t.Fatal(err)
	}
	// This frame has a pulse with no valid peak.
	setHistogram(chip, 0, [histogramBufferSize]int32{
		2047, 3947, 3287, 2888, 3790, 1015, 1541, 1408, 3387, 3831, 1429, 1356,
		2737, 1631, 2485, 2026, 3413, 2090, 3194, 2563, 1433, 1147, 2078, 2324,
	})

	var m Measurement
	if err := dev.Read(&m); err != nil {
		t.Fatal(err)
	}
}

func TestReadHardwareFail(t *testing.T) {
	dev, chip := configure(t, Config{})
	if err := dev.StartRanging(); err != nil {
		t.Fatal(err)
	}
	setHistogram(chip, 0, pulseHistogram(8))
	chip.Registers[regResultInterruptStatus+1] = deviceErrorVcselWatchdogFail

	var m Measurement
	if err := dev.Read(&m); err != nil {
		t.Fatal(err)
	}
	if m.Targets[0].Status != HardwareFail {
		t.Errorf("status = %d, want %d", m.Targets[0].Status, HardwareFail)
	}
}

func TestTimeoutEncoding(t *testing.T) {
	if encodeTimeout(0) != 0 {
		t.Errorf("encodeTimeout(0) = %#x, want 0", encodeTimeout(0))
	}
	for _, mclks := range []uint32{1, 2, 255, 256, 257, 1000, 0x1FF00} {
		enc := encodeTimeout(mclks)
		dec := decodeTimeout(enc)
		if dec > mclks || mclks-dec >= mclks/128+1 {
			t.Errorf("decodeTimeout(encodeTimeout(%d)) = %d", mclks, dec)
		}
	}
	if v := encodeTimeout(256); v != 0x00FF {
		t.Errorf("encodeTimeout(256) = %#x, want 0x00FF", v)
	}
	if v := encodeTimeout(257); v != 0x0180 {
		t.Errorf("encodeTimeout(257) = %#x, want 0x0180", v)
	}
}
