package vl53l3cx

import "testing"

// newAlgo returns algorithm data for one period of bins with no ambient.
func newAlgo(bins ...int32) *algoData {
	a := &algoData{}
	a.init()
	a.bins.binBufferSize = histogramBufferSize
	a.bins.binCount = uint8(len(bins))
	copy(a.bins.bins[:], bins)
	a.binCount = uint8(len(bins))
	a.periodBins = uint8(len(bins))
	return a
}

func setAbove(a *algoData, bins ...uint8) {
	for _, b := range bins {
		a.aboveThreshold[b] = 1
	}
}

func TestBinSequenceConfig(t *testing.T) {
	var d Device
	d.presetHistogram(Long)
	d.binSequenceConfig()
	if d.hist.binSeq != d.histCfg.even {
		t.Errorf("even frame bin sequence = %v, want %v", d.hist.binSeq, d.histCfg.even)
	}
	d.state.rdStreamCount = 1
	d.binSequenceConfig()
	if d.hist.binSeq != d.histCfg.odd {
		t.Errorf("odd frame bin sequence = %v, want %v", d.hist.binSeq, d.histCfg.odd)
	}
	if d.hist.binRep != [maxBinSequence]uint8{1, 1, 1, 1, 1, 1} {
		t.Errorf("bin repeats = %v, want all 1", d.hist.binRep)
	}
}

func TestHistogramMerge(t *testing.T) {
	var d Device
	var near, far [histogramBufferSize]int32
	for i := range near {
		near[i] = 100
		far[i] = 100
	}
	near[8] = 5000
	far[16] = 5000

	d.hist.bins = near
	d.histogramMerge()
	if d.hist.bins != near {
		t.Fatal("first frame should not change the bins")
	}

	d.hist.streamCount = 2
	d.hist.bins = near
	d.histogramMerge()
	if d.hist.bins[8] != 2*near[8] || d.hist.bins[0] != 2*near[0] {
		t.Errorf("merged bins 0 and 8 = %d %d, want %d %d", d.hist.bins[0], d.hist.bins[8], 2*near[0], 2*near[8])
	}
	d.hist.binSeq[0] = 7
	if n := d.histogramMergeCount(); n != 2 {
		t.Errorf("merge count = %d, want 2", n)
	}

	d.hist.streamCount = 4
	d.hist.bins = far
	d.histogramMerge()
	if d.hist.bins != far {
		t.Error("frame after reset should not be merged")
	}
	if d.merge.waitBeforeNext != mergeFrameWaitEvent+1 {
		t.Errorf("wait = %d, want %d", d.merge.waitBeforeNext, mergeFrameWaitEvent+1)
	}
	if n := d.histogramMergeCount(); n != 0 {
		t.Errorf("merge count after reset = %d, want 0", n)
	}

	d.histogramMerge()
	if d.merge.waitBeforeNext != mergeFrameWaitEvent {
		t.Errorf("wait = %d, want %d", d.merge.waitBeforeNext, mergeFrameWaitEvent)
	}
}

func TestCopyAndScaleAmbientInfo(t *testing.T) {
	var h histogramData
	if copyAndScaleAmbientInfo(&zoneHistInfo{}, &h) {
		t.Error("expected false without previous effective SPADs")
	}

	prev := zoneHistInfo{numAmbientBins: 4, effectiveSpads: 100, totalPeriodsElapsed: 9, ambientEventsSum: 4000}
	h = histogramData{effectiveSpads: 200, totalPeriodsElapsed: 9}
	if !copyAndScaleAmbientInfo(&prev, &h) {
		t.Fatal("expected true")
	}
	if h.ambientEventsSum != 8000 || h.ambientPerBin != 2000 {
		t.Errorf("ambient sum %d per bin %d, want 8000 2000", h.ambientEventsSum, h.ambientPerBin)
	}
}

func TestPhaseConsistencyCheck(t *testing.T) {
	tests := []struct {
		name    string
		state   uint8
		objects uint8
		phase   uint16
		want    uint8
	}{
		{"not output data", stateWaitGPHSync, 0, 0, deviceErrorRangeCompleteNoWrapCheck},
		{"no previous targets", stateOutputData, 0, 0, deviceErrorPrevRangeNoTargets},
		{"same phase", stateOutputData, 1, 0x2400 + 0x100, deviceErrorRangeComplete},
		{"moved phase", stateOutputData, 1, 0x4400, deviceErrorPhaseConsistency},
	}
	for _, tt := range tests {
		var d Device
		d.prevObjects = zoneObjects{rdDeviceState: tt.state, activeObjects: tt.objects}
		d.prevObjects.phase[0] = tt.phase
		d.results.activeResults = 1
		d.results.data[0].phase = 0x2400
		d.results.data[0].rangeStatus = deviceErrorRangeCompleteNoWrapCheck
		d.phaseConsistencyCheck()
		if s := d.results.data[0].rangeStatus; s != tt.want {
			t.Errorf("%s: status = %d, want %d", tt.name, s, tt.want)
		}
	}
}

func TestFillMeasurementROIClip(t *testing.T) {
	var d Device
	d.results.deviceStatus = deviceErrorUserROIClip
	var m Measurement
	d.fillMeasurement(&m)
	if m.Targets[0].Status != MinRangeFail {
		t.Errorf("status = %d, want %d", m.Targets[0].Status, MinRangeFail)
	}
}

func TestAverageHistogram(t *testing.T) {
	in := histogramData{
		binBufferSize: histogramBufferSize,
		binSeq:        [maxBinSequence]uint8{7, 0, 1, 1, 2, 2},
	}
	for i := range in.bins {
		in.bins[i] = int32(i)
	}
	var out histogramData
	averageHistogram(&in, &out)

	if out.binSeq != [maxBinSequence]uint8{7, 0, 1, 2, 16, 16} {
		t.Errorf("bin sequence = %v", out.binSeq)
	}
	if out.binRep != [maxBinSequence]uint8{1, 1, 2, 2, 0, 0} {
		t.Errorf("bin repeats = %v", out.binRep)
	}
	if out.binCount != 16 || out.numAmbientBins != 4 {
		t.Errorf("bin count %d ambient bins %d, want 16 4", out.binCount, out.numAmbientBins)
	}
	// Bins 8 and 12 are averaged to 10, bins 16 and 20 to 18.
	want := []int32{0, 1, 2, 3, 4, 5, 6, 7, 10, 11, 12, 13, 18, 19, 20, 21}
	for i, w := range want {
		if out.bins[i] != w {
			t.Errorf("bin %d = %d, want %d", i, out.bins[i], w)
		}
	}
	for i := len(want); i < histogramBufferSize; i++ {
		if out.bins[i] != 0 {
			t.Errorf("bin %d = %d, want 0", i, out.bins[i])
		}
	}
}

func TestCalcZeroDistancePhase(t *testing.T) {
	tests := []struct {
		refPhase   uint16
		vcselStart uint8
		calStart   uint8
		want       uint16
	}{
		{0x0800, 9, 9, 0x0800},
		{0x0000, 0, 9, 11 * 2048},
		{0x0000, 9, 9, 0},
	}
	for _, tt := range tests {
		h := histogramData{vcselPeriod: 9, phasecalRefPhase: tt.refPhase,
			phasecalVcselStart: tt.vcselStart, calVcselStart: tt.calStart}
		calcZeroDistancePhase(&h)
		if h.zeroDistancePhase != tt.want {
			t.Errorf("%+v: zero distance phase = %#x, want %#x", tt, h.zeroDistancePhase, tt.want)
		}
	}
}

func TestAmbientBins(t *testing.T) {
	h := histogramData{
		binBufferSize:  histogramBufferSize,
		binCount:       histogramBufferSize,
		numAmbientBins: 4,
		binSeq:         [maxBinSequence]uint8{7, 0, 1, 2, 3, 4},
		binRep:         [maxBinSequence]uint8{1, 1, 1, 1, 1, 1},
	}
	for i := range h.bins {
		h.bins[i] = int32(10 + i)
	}
	estimateAmbientFromAmbientBins(&h)
	if h.ambientEventsSum != 46 || h.ambientPerBin != 12 {
		t.Errorf("ambient sum %d per bin %d, want 46 12", h.ambientEventsSum, h.ambientPerBin)
	}

	removeAmbientBins(&h)
	if h.binSeq != [maxBinSequence]uint8{0, 1, 2, 3, 4, 16} {
		t.Errorf("bin sequence = %v", h.binSeq)
	}
	if h.binRep != [maxBinSequence]uint8{1, 1, 1, 1, 1, 0} {
		t.Errorf("bin repeats = %v", h.binRep)
	}
	if h.binCount != 20 || h.numAmbientBins != 0 || h.bins[0] != 14 || h.bins[19] != 33 {
		t.Errorf("bin count %d ambient bins %d first %d last %d, want 20 0 14 33",
			h.binCount, h.numAmbientBins, h.bins[0], h.bins[19])
	}
}

func TestFindThresholds(t *testing.T) {
	a := newAlgo(make([]int32, 20)...)
	a.bins.vcselPeriod = 9
	a.bins.binRep = [maxBinSequence]uint8{1, 1, 1, 1, 1, 0}
	a.bins.ambientPerBin = 1000
	a.bins.bins[5] = 1210
	a.bins.bins[6] = 1209
	a.findThresholds()

	// ambient 1015 + sqrt(1015) * 100 / 16 from ST's tuning defaults
	if a.threshold[0] != 1209 || a.periodBins != 20 {
		t.Errorf("threshold %d period bins %d, want 1209 20", a.threshold[0], a.periodBins)
	}
	if a.aboveThreshold[5] != 1 || a.aboveThreshold[6] != 0 {
		t.Errorf("above threshold = %v", a.aboveThreshold)
	}

	a = newAlgo(make([]int32, 20)...)
	a.bins.binRep = [maxBinSequence]uint8{1, 1, 1, 1, 1, 0}
	a.findThresholds()
	if a.threshold[0] != minAmbientThreshEvents {
		t.Errorf("threshold without ambient = %d, want %d", a.threshold[0], minAmbientThreshEvents)
	}
}

func TestFindPulses(t *testing.T) {
	a := newAlgo(make([]int32, 20)...)
	setAbove(a, 3, 4, 5, 10, 11)
	a.findStartBin()
	a.countPulses()
	a.findPulseEdges()

	if a.startBin != 2 || a.pulseCount != 2 {
		t.Fatalf("start bin %d pulses %d, want 2 2", a.startBin, a.pulseCount)
	}
	want := []pulseData{
		{start: 2, first: 3, peak: 0xFF, last: 5, end: 6, width: 3, filterWOI: 5},
		{start: 9, first: 10, peak: 0xFF, last: 11, end: 12, width: 2, filterWOI: 4},
	}
	for i, w := range want {
		if a.pulses[i] != w {
			t.Errorf("pulse %d = %+v, want %+v", i, a.pulses[i], w)
		}
	}
}

func TestFindPulsesWrap(t *testing.T) {
	a := newAlgo(make([]int32, 20)...)
	setAbove(a, 19, 0, 1)
	a.findStartBin()
	a.countPulses()
	a.findPulseEdges()

	if a.startBin != 18 || a.pulseCount != 1 {
		t.Fatalf("start bin %d pulses %d, want 18 1", a.startBin, a.pulseCount)
	}
	p := a.pulses[0]
	if p.start != 18 || p.first != 19 || p.last != 21 || p.end != 22 || p.width != 3 {
		t.Errorf("pulse = %+v, want bins 18 to 22", p)
	}
}

func TestCountPulsesLimit(t *testing.T) {
	a := newAlgo(make([]int32, 20)...)
	for i := uint8(1); i < 20; i += 2 {
		setAbove(a, i)
	}
	a.findStartBin()
	a.countPulses()
	if a.pulseCount != maxPulses {
		t.Errorf("pulses = %d, want %d", a.pulseCount, maxPulses)
	}
}

func TestPulseEventsAndPad(t *testing.T) {
	a := newAlgo(10, 10, 50, 90, 50, 10, 10, 10)
	a.ambientPerBin = 10
	a.pulses[0] = pulseData{start: 1, end: 5}
	a.pulseEvents(0)
	p := a.pulses[0]
	if p.totalEvents != 210 || p.ambientEvents != 50 || p.signalEvents != 160 {
		t.Errorf("total %d ambient %d signal %d, want 210 50 160", p.totalEvents, p.ambientEvents, p.signalEvents)
	}

	var out histogramData
	a.padPulse(0, &out, -1)
	want := []int32{-1, 10, 50, 90, 50, 10, -1, -1}
	for i, w := range want {
		if out.bins[i] != w {
			t.Errorf("padded bin %d = %d, want %d", i, out.bins[i], w)
		}
	}
}

func TestFilterSums(t *testing.T) {
	h := histogramData{binCount: 8, bins: [histogramBufferSize]int32{1, 2, 4, 8, 16, 32, 64, 128}}
	if sa, sb, sc := filterSums(3, 2, &h); sa != 6 || sb != 8 || sc != 48 {
		t.Errorf("filterSums(3, 2) = %d %d %d, want 6 8 48", sa, sb, sc)
	}
	if sa, sb, sc := filterSums(0, 1, &h); sa != 128 || sb != 1 || sc != 2 {
		t.Errorf("filterSums(0, 1) = %d %d %d, want 128 1 2", sa, sb, sc)
	}
}

func TestDetectPeak(t *testing.T) {
	a := newAlgo(0, 0, 0, 20, 40, 20, 0, 0, 0, 0)
	a.pulses[0] = pulseData{start: 2, end: 6, filterWOI: 1, peak: 0xFF}
	a.pulseBins = a.bins
	var f filteredData
	a.filterPulse(0, &f)
	a.detectPeak(0, &f)
	p := a.pulses[0]
	if p.peak != 4 || p.phase != 1024+4*2048 {
		t.Errorf("peak %d phase %#x, want 4 %#x", p.peak, p.phase, 1024+4*2048)
	}
}

func TestMeanPhase(t *testing.T) {
	if p, ok := meanPhase(4, 100, 200, 100, 0, 20); !ok || p != 1024+4*2048 {
		t.Errorf("symmetric phase = %#x %v, want %#x", p, ok, 1024+4*2048)
	}
	if p, _ := meanPhase(4, 50, 200, 150, 0, 20); p <= 1024+4*2048 {
		t.Errorf("phase with later energy = %#x, want above %#x", p, 1024+4*2048)
	}
	if p, _ := meanPhase(4, 150, 200, 50, 0, 20); p >= 1024+4*2048 {
		t.Errorf("phase with earlier energy = %#x, want below %#x", p, 1024+4*2048)
	}
	if p, ok := meanPhase(4, 100, 50, 100, 50, 20); ok || p != maxAllowedPhase {
		t.Errorf("phase without signal = %#x %v, want %#x false", p, ok, maxAllowedPhase)
	}
	if p, _ := meanPhase(9, 100, 200, 100, 0, 8); p != 1024+2048 {
		t.Errorf("wrapped phase = %#x, want %#x", p, 1024+2048)
	}
}

func TestEstimateSigma(t *testing.T) {
	h := histogramData{binCount: 10, fastOscFrequency: 0xBCCC}
	copy(h.bins[:], []int32{0, 0, 0, 20, 40, 10, 0, 0, 0, 0})
	min := uint16(sigmaEstimatorRefMM << 7)
	clean := estimateSigma(4, sigmaEstimatorRefMM, 10, 1, &h, &h)
	if clean < min {
		t.Errorf("sigma = %d, want at least %d", clean, min)
	}

	noisy := h
	noisy.ambientPerBin = 10
	for i := range noisy.bins[:10] {
		noisy.bins[i] += 10
	}
	if s := estimateSigma(4, sigmaEstimatorRefMM, 10, 1, &noisy, &h); s <= clean {
		t.Errorf("sigma with ambient = %d, want above %d", s, clean)
	}

	h.fastOscFrequency = 0
	if s := estimateSigma(4, sigmaEstimatorRefMM, 10, 1, &h, &h); s != 0xFFFF {
		t.Errorf("sigma without oscillator = %#x, want 0xFFFF", s)
	}
}

func TestWindowPhase(t *testing.T) {
	a := newAlgo(0, 0, 100, 0, 0, 0, 0, 0)
	if p := a.windowPhase(1, 3); p != 1024+2*2048 {
		t.Errorf("phase = %#x, want %#x", p, 1024+2*2048)
	}
	if p := a.windowPhase(4, 6); p != maxAllowedPhase {
		t.Errorf("phase without events = %#x, want %#x", p, maxAllowedPhase)
	}

	a = newAlgo(0, 0, 0, 0, 0, 0, 0, 100)
	if p := a.windowPhase(-1, 0); p != 0 {
		t.Errorf("phase before bin 0 = %#x, want 0", p)
	}
}

func TestPulsePhaseRange(t *testing.T) {
	a := newAlgo(0, 0, 10, 40, 80, 40, 10, 0, 0, 0)
	a.pulses[0] = pulseData{start: 1, peak: 4, end: 7, phase: 1024 + 4*2048}
	a.pulsePhaseRange(0)
	p := a.pulses[0]
	if p.phaseLow >= p.phase || p.phaseHigh <= p.phase {
		t.Errorf("phase range %#x %#x %#x, want low < phase < high", p.phaseLow, p.phase, p.phaseHigh)
	}
}

func TestSortPulses(t *testing.T) {
	var a algoData
	a.pulseCount = 3
	a.pulses[0].phase = 300
	a.pulses[1].phase = 100
	a.pulses[2].phase = 200
	a.sortPulses()
	for i, w := range []uint32{100, 200, 300} {
		if a.pulses[i].phase != w {
			t.Errorf("pulse %d phase = %d, want %d", i, a.pulses[i].phase, w)
		}
	}
}

func TestSetRangeData(t *testing.T) {
	h := histogramData{zeroDistancePhase: 0x0800}
	tests := []struct {
		name  string
		phase uint32
		sigma uint16
		want  uint8
	}{
		{"valid", 0x2400, 0, deviceErrorRangeCompleteNoWrapCheck},
		{"sigma", 0x2400, (sigmaThreshMM << 5) + 1, deviceErrorSigmaThresholdCheck},
		{"too far", 0x0800 + 0x8800 + 1, 0, deviceErrorRangePhaseCheck},
	}
	for _, tt := range tests {
		p := pulseData{start: 2, end: 6, phase: tt.phase, sigma: tt.sigma}
		var r rangeData
		setRangeData(0x08, 0x88, sigmaThreshMM, &h, &p, &r)
		if r.rangeStatus != tt.want {
			t.Errorf("%s: status = %d, want %d", tt.name, r.rangeStatus, tt.want)
		}
		if r.woiWidth != 5 || r.zeroDistancePhase != 0x0800 {
			t.Errorf("%s: woi width %d zero phase %#x, want 5 0x800", tt.name, r.woiWidth, r.zeroDistancePhase)
		}
	}
}

func TestCalcRates(t *testing.T) {
	r := rangeData{woiWidth: 5, signalEvents: 10000, ambientEvents: 1000}
	calcRates(0x28, 0xBCCC, 100, 0x2000, &r)
	if r.peakDurationUs == 0 || r.woiDurationUs <= r.peakDurationUs {
		t.Errorf("peak duration %d woi duration %d", r.peakDurationUs, r.woiDurationUs)
	}
	if r.peakSignalRate == 0 || r.avgSignalRate >= r.peakSignalRate || r.ambientRate == 0 {
		t.Errorf("peak %d avg %d ambient %d", r.peakSignalRate, r.avgSignalRate, r.ambientRate)
	}

	r = rangeData{signalEvents: 10000}
	calcRates(0x28, 0, 100, 0x2000, &r)
	if r.peakSignalRate != 0 {
		t.Errorf("rate without oscillator = %d, want 0", r.peakSignalRate)
	}
}

func TestRangeMaths(t *testing.T) {
	if r := rangeMaths(0xBCCC, 0x0800, 0x0800, 2048, 0); r != 0 {
		t.Errorf("range at zero phase = %d, want 0", r)
	}
	// The offset is in quarter millimeters and the gain is 5.11 fixed point.
	if r := rangeMaths(0xBCCC, 0x0800, 0x0800, 2048, 40); r != 10 {
		t.Errorf("range with offset = %d, want 10", r)
	}
	near := rangeMaths(0xBCCC, 0x2400, 0x0800, histGainFactor, 0)
	far := rangeMaths(0xBCCC, 0x4400, 0x0800, histGainFactor, 0)
	if near <= 0 || far <= near {
		t.Errorf("ranges = %d %d, want increasing", near, far)
	}

	var r rangeData
	r.fastOscFrequency = 0xBCCC
	r.zeroDistancePhase = 0x0800
	r.phaseLow, r.phase, r.phaseHigh = 0x2000, 0x2400, 0x2800
	calcRanges(histGainFactor, 0, &r)
	if int32(r.medianRangeMM) != near || r.minRangeMM >= r.medianRangeMM || r.maxRangeMM <= r.medianRangeMM {
		t.Errorf("ranges = %d %d %d", r.minRangeMM, r.medianRangeMM, r.maxRangeMM)
	}
}

func TestDurationMaths(t *testing.T) {
	if d := durationMaths(1<<12, 1<<4, 1, 1<<12); d != 1 {
		t.Errorf("duration = %d, want 1", d)
	}
	if d := durationMaths(1<<12, 1<<4, 2048, 1<<12); d != 2048 {
		t.Errorf("duration = %d, want 2048", d)
	}
}

func TestRatePerSpadMaths(t *testing.T) {
	if v := ratePerSpadMaths(0, 1, 256, 0xFFFF); v != 1 {
		t.Errorf("rate per spad = %d, want 1", v)
	}
	if v := ratePerSpadMaths(6, 10, 0, 0xFFFF); v != 10<<6 {
		t.Errorf("rate without spads = %d, want %d", v, 10<<6)
	}
	if v := ratePerSpadMaths(6, 0xFFFF, 1, 0x1234); v != 0x1234 {
		t.Errorf("clamped rate = %#x, want 0x1234", v)
	}
}

func TestEventsPerSpadMaths(t *testing.T) {
	if v := eventsPerSpadMaths(1000, 0, 100); v != 0 {
		t.Errorf("events without spads = %d, want 0", v)
	}
	if v := eventsPerSpadMaths(1, 256, 0); v != 1000<<11 {
		t.Errorf("events without duration = %d, want %d", v, 1000<<11)
	}
	if v := eventsPerSpadMaths(1, 256, 1<<11); v != 1000 {
		t.Errorf("events per spad = %d, want 1000", v)
	}
}

func TestCalcPLLPeriod(t *testing.T) {
	if p := calcPLLPeriodUs(0); p != 0 {
		t.Errorf("pll period without oscillator = %d, want 0", p)
	}
	if p := calcPLLPeriodUs(0x1000); p != 1<<18 {
		t.Errorf("pll period = %d, want %d", p, 1<<18)
	}
	if mm := calcPLLPeriodMM(0xBCCC); mm == 0 {
		t.Error("pll period in mm should not be 0")
	}
}

func TestCalcMMEffectiveSpads(t *testing.T) {
	var good [32]uint8
	for i := range good {
		good[i] = 0xFF
	}
	// The full 16x16 array has 32 aperture SPADs.
	inner, outer := calcMMEffectiveSpads(0xC7, 0xFF, 0xC7, 0xFF, &good, 0x38)
	if inner != 224*0x100+32*0x38 || outer != 0 {
		t.Errorf("inner %d outer %d, want %d 0", inner, outer, 224*0x100+32*0x38)
	}

	inner, outer = calcMMEffectiveSpads(0xC7, 0x33, 0xC7, 0xFF, &good, 0x38)
	if inner == 0 || outer == 0 || uint32(inner)+uint32(outer) != 224*0x100+32*0x38 {
		t.Errorf("inner %d outer %d, want the full array split", inner, outer)
	}

	good = [32]uint8{}
	if inner, outer = calcMMEffectiveSpads(0xC7, 0xFF, 0xC7, 0xFF, &good, 0x38); inner != 0 || outer != 0 {
		t.Errorf("inner %d outer %d without good SPADs, want 0 0", inner, outer)
	}
}

func TestCombineMM1MM2Offsets(t *testing.T) {
	var good [32]uint8
	for i := range good {
		good[i] = 0xFF
	}
	cal := additionalOffsetCal{mmInnerPeakRate: 0x80, mmOuterPeakRate: 0x180}
	if o := combineMM1MM2Offsets(10, 10, 0xC7, 0x33, 0xC7, 0xFF, &cal, &good, 0x38); o != 40 {
		t.Errorf("offset = %d, want 40", o)
	}

	good = [32]uint8{}
	if o := combineMM1MM2Offsets(10, 10, 0xC7, 0x33, 0xC7, 0xFF, &cal, &good, 0x38); o != 0 {
		t.Errorf("offset without good SPADs = %d, want 0", o)
	}
}

func TestIsApertureLocation(t *testing.T) {
	n := 0
	for row := uint8(0); row < 16; row++ {
		for col := uint8(0); col < 16; col++ {
			if isApertureLocation(row, col) {
				n++
			}
		}
	}
	if n != 32 || !isApertureLocation(0, 2) || !isApertureLocation(2, 0) || isApertureLocation(0, 0) {
		t.Errorf("aperture count = %d, want 32", n)
	}
}

func TestFillMeasurementNegativeDistance(t *testing.T) {
	tests := []struct {
		distance int16
		want     int16
		status   RangeStatus
	}{
		{-10, 0, RangeValid},
		{proxyMinMM - 1, proxyMinMM - 1, RangeInvalid},
	}
	for _, tt := range tests {
		var d Device
		d.results.activeResults = 1
		d.results.data[0].rangeStatus = deviceErrorRangeComplete
		d.results.data[0].medianRangeMM = tt.distance
		var m Measurement
		d.fillMeasurement(&m)
		if m.Targets[0].Distance != tt.want || m.Targets[0].Status != tt.status {
			t.Errorf("distance %d: got %d status %d, want %d status %d",
				tt.distance, m.Targets[0].Distance, m.Targets[0].Status, tt.want, tt.status)
		}
	}
}

func TestConvertStatus(t *testing.T) {
	tests := map[uint8]RangeStatus{
		deviceErrorRangeComplete:            RangeValid,
		deviceErrorRangePhaseCheck:          OutOfBoundsFail,
		deviceErrorSigmaThresholdCheck:      SigmaFail,
		deviceErrorRangeCompleteNoWrapCheck: RangeValidNoWrapCheckFail,
		deviceErrorPhaseConsistency:         WrapTargetFail,
		deviceErrorEventConsistency:         WrapTargetFail,
		deviceErrorPrevRangeNoTargets:       TargetPresentLackOfSignal,
		deviceErrorRangeCompleteMergedPulse: RangeValidMergedPulse,
		deviceErrorNoUpdate:                 NoTarget,
	}
	for in, want := range tests {
		if got := convertStatus(in); got != want {
			t.Errorf("convertStatus(%d) = %d, want %d", in, got, want)
		}
	}
}

func TestIsqrt(t *testing.T) {
	for _, n := range []uint32{0, 1, 2, 3, 4, 15, 16, 17, 1000000, 0xFFFFFFFF} {
		r := isqrt(n)
		if uint64(r)*uint64(r) > uint64(n) || uint64(r+1)*uint64(r+1) <= uint64(n) {
			t.Errorf("isqrt(%d) = %d", n, r)
		}
	}
}

func TestRowCol(t *testing.T) {
	for spad := 0; spad < 256; spad++ {
		row, col := decodeRowCol(uint8(spad))
		if got := encodeRowCol(row, col); got != uint8(spad) {
			t.Errorf("encodeRowCol(decodeRowCol(%d)) = %d", spad, got)
		}
	}
}

func TestDecodeZoneLimits(t *testing.T) {
	xLL, yLL, xUR, yUR := decodeZoneLimits(0xC7, 0xFF)
	if xLL != 0 || yLL != 0 || xUR != 15 || yUR != 15 {
		t.Errorf("full zone = %d %d %d %d, want 0 0 15 15", xLL, yLL, xUR, yUR)
	}
}

func TestRateToKcps(t *testing.T) {
	tests := map[uint16]uint32{0: 0, 1 << 7: 1000, 3 << 6: 1500}
	for in, want := range tests {
		if got := rateToKcps(in); got != want {
			t.Errorf("rateToKcps(%#x) = %d, want %d", in, got, want)
		}
	}
}

func TestRateMaths(t *testing.T) {
	if v := rateMaths(1000, 1000); v != 1<<7 {
		t.Errorf("rateMaths(1000, 1000) = %#x, want %#x", v, 1<<7)
	}
	if v := rateMaths(-5, 1000); v != 0 {
		t.Errorf("rateMaths(-5, 1000) = %#x, want 0", v)
	}
	if v := rateMaths(1<<29, 1); v != 0xFFFF {
		t.Errorf("rateMaths(1<<29, 1) = %#x, want 0xFFFF", v)
	}
}
