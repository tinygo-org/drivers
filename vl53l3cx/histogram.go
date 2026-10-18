package vl53l3cx

import "encoding/binary"

// The histogram processing below is ported from ST's VL53LX API
// (vl53lx_api_core.c, vl53lx_hist_funcs.c, vl53lx_hist_algos_gen3.c and
// vl53lx_hist_algos_gen4.c). Function comments name the ST function.

const (
	histogramBufferSize = 24
	maxBinSequence      = 6
	maxBinSequenceCode  = 15
	maxPulses           = 8
	maxAllowedPhase     = 0xFFFF

	rangingWindowVcselPeriods = 2048
	speedOfLightInAirDiv8     = 299704 >> 3

	// Tuning defaults from ST's API (VL53LX_TUNINGPARM_*_DEFAULT)
	histGainFactor           = 1987
	ambientThreshSigma1      = 100
	ambientThreshEventScaler = 4157
	minAmbientThreshEvents   = 16
	signalTotalEventsLimit   = 100
	sigmaEstimatorRefMM      = 1
	sigmaThreshMM            = 180
	phaseTolerance           = 8
	proxyMinMM               = -30
	mergeMaxSize             = 6
	mergeResetThreshold      = 15000
	mergeFrameWaitEvent      = 6
)

type histogramData struct {
	binStart      uint8
	binBufferSize uint8
	binCount      uint8

	numAmbientBins uint8
	binSeq         [maxBinSequence]uint8
	binRep         [maxBinSequence]uint8
	bins           [histogramBufferSize]int32

	interruptStatus uint8
	rangeStatus     uint8
	streamCount     uint8
	effectiveSpads  uint16

	phasecalRefPhase   uint16
	phasecalVcselStart uint8
	calVcselStart      uint8
	vcselWidth         uint16
	vcselPeriod        uint8
	fastOscFrequency   uint16

	totalPeriodsElapsed uint32
	peakDurationUs      uint32
	minBinValue         int32
	maxBinValue         int32
	zeroDistancePhase   uint16
	numAmbientSamples   uint8
	ambientEventsSum    int32
	ambientPerBin       int32

	roiCentreSpad uint8
	roiXYSize     uint8
}

type pulseData struct {
	start     uint8
	first     uint8
	peak      uint8
	last      uint8
	end       uint8
	width     uint8
	filterWOI uint8

	ambientEvents int32
	totalEvents   int32
	signalEvents  int32

	phaseLow  uint32
	phase     uint32
	phaseHigh uint32
	sigma     uint16
}

type rangeData struct {
	woiWidth            uint8
	vcselWidth          uint16
	fastOscFrequency    uint16
	zeroDistancePhase   uint16
	effectiveSpads      uint16
	totalPeriodsElapsed uint32
	peakDurationUs      uint32
	woiDurationUs       uint32

	ambientEvents uint32
	totalEvents   uint32
	signalEvents  int32

	peakSignalRate  uint16
	avgSignalRate   uint16
	ambientRate     uint16
	totalRatePerSpd uint16
	signalPerSpad   uint32

	sigma     uint16
	phaseLow  uint16
	phase     uint16
	phaseHigh uint16

	minRangeMM    int16
	medianRangeMM int16
	maxRangeMM    int16
	rangeStatus   uint8
}

type rangeResults struct {
	streamCount   uint8
	deviceStatus  uint8
	activeResults uint8
	data          [MaxTargets]rangeData
}

type histogramMerge struct {
	rec            [mergeMaxSize][2][histogramBufferSize]int32
	pos            uint8
	waitBeforeNext uint8
}

// algoData is VL53LX_hist_gen3_algo_private_data_t.
type algoData struct {
	binStart      uint8
	binBufferSize uint8
	binCount      uint8
	periodBins    uint8

	ambientPerBin  int32
	lastThreshold  int32
	aboveThreshold [histogramBufferSize]uint8
	pulseIndex     [histogramBufferSize]uint8
	threshold      [histogramBufferSize]int32

	startBin   uint8
	pulseCount uint8
	pulses     [maxPulses]pulseData

	bins          histogramData
	pulseBins     histogramData
	pulseBinsZero histogramData
}

// filteredData is VL53LX_hist_gen4_algo_filtered_data_t.
type filteredData struct {
	a     [histogramBufferSize]int32
	b     [histogramBufferSize]int32
	c     [histogramBufferSize]int32
	diff1 [histogramBufferSize]int32
	diff2 [histogramBufferSize]int32
	valid [histogramBufferSize]uint8
}

// deviceResults is the histogram path of VL53LX_get_device_results.
func (d *Device) deviceResults() error {
	ok, err := d.histogramBinData()
	if err != nil {
		return err
	}
	h := &d.hist
	if ok && h.numAmbientBins == 0 {
		ok = copyAndScaleAmbientInfo(&d.prevHist, h)
	}

	if ok {
		mergeNb := d.histogramMergeCount()
		if mergeNb == 0 {
			mergeNb = 1
		}

		d.rangeOffsetMM = combineMM1MM2Offsets(
			d.customerMMInnerOffset(), d.customerMMOuterOffset(),
			d.modeROICentreSpad, d.modeROIXYSize,
			h.roiCentreSpad, h.roiXYSize,
			&d.addOffCal, &d.goodSpads, uint16(d.generalCfg.dssApertureAttenuation))

		d.processHistogram()

		if mergeNb > 1 {
			n := uint32(mergeNb)
			for i := range d.results.data {
				r := &d.results.data[i]
				r.ambientEvents /= n
				r.totalEvents /= n
				r.signalEvents /= int32(mergeNb)
				r.peakSignalRate /= uint16(mergeNb)
				r.avgSignalRate /= uint16(mergeNb)
				r.ambientRate /= uint16(mergeNb)
				r.signalPerSpad /= n
			}
		}

		d.phaseConsistencyCheck()

		d.prevHist = zoneHistInfo{
			numAmbientBins:      h.numAmbientBins,
			effectiveSpads:      h.effectiveSpads,
			totalPeriodsElapsed: h.totalPeriodsElapsed,
			ambientEventsSum:    h.ambientEventsSum,
		}

		p := &d.prevObjects
		p.rdDeviceState = d.state.rdDeviceState
		p.activeObjects = d.results.activeResults
		for i := uint8(0); i < d.results.activeResults; i++ {
			p.phase[i] = d.results.data[i].phase
		}
	}
	return nil
}

// histogramBinData is VL53LX_get_histogram_bin_data. It returns false when
// the device reports a hardware or ROI error.
func (d *Device) histogramBinData() (bool, error) {
	b := d.buf[2 : 2+histogramBinDataSize]
	if err := d.readBytes(regResultInterruptStatus, b); err != nil {
		return false, err
	}
	h := &d.hist
	h.interruptStatus = b[0]
	h.rangeStatus = b[1]
	h.streamCount = b[3]
	h.effectiveSpads = binary.BigEndian.Uint16(b[4:])
	h.phasecalRefPhase = binary.BigEndian.Uint16(b[regPhasecalReferencePhase-regResultInterruptStatus:])
	h.phasecalVcselStart = b[regPhasecalVcselStart-regResultInterruptStatus]

	b[regHistogramBin23_0-regResultInterruptStatus] =
		b[regHistogramBin23_0MSB-regResultInterruptStatus]<<2 + b[regHistogramBin23_0LSB-regResultInterruptStatus]

	p := b[regHistogramBin0_2-regResultInterruptStatus:]
	for i := range h.bins {
		h.bins[i] = int32(uint32(p[0])<<16 | uint32(p[1])<<8 | uint32(p[2]))
		p = p[3:]
	}

	if h.streamCount == 0 {
		d.merge = histogramMerge{}
	}
	d.histogramMerge()

	h.binStart = 0
	h.binBufferSize = histogramBufferSize
	h.binCount = histogramBufferSize
	h.calVcselStart = d.generalCfg.calVcselStart
	h.vcselWidth = uint16(d.generalCfg.vcselWidth)<<4 + uint16(d.staticCfg.vcselPulseWidthOffset)
	h.fastOscFrequency = d.fastOscFrequency()

	d.binSequenceConfig()

	var encodedTimeout uint16
	if d.state.rdTimingStatus == 0 {
		encodedTimeout = d.timingCfg.rangeTimeoutA
		h.vcselPeriod = d.timingCfg.vcselPeriodA
	} else {
		encodedTimeout = d.timingCfg.rangeTimeoutB
		h.vcselPeriod = d.timingCfg.vcselPeriodB
	}

	h.numAmbientBins = 0
	for i := range h.binSeq {
		if h.binSeq[i]&0x07 == 0x07 {
			h.numAmbientBins += 4
		}
	}

	h.totalPeriodsElapsed = decodeTimeout(encodedTimeout)
	h.peakDurationUs = durationMaths(calcPLLPeriodUs(h.fastOscFrequency), uint32(h.vcselWidth),
		rangingWindowVcselPeriods, h.totalPeriodsElapsed+1)

	calcZeroDistancePhase(h)
	estimateAmbientFromAmbientBins(h)

	h.roiCentreSpad = d.savedROI
	h.roiXYSize = d.savedROISize

	d.results.deviceStatus = deviceErrorNoUpdate
	switch s := h.rangeStatus & 0x1F; s {
	case deviceErrorVcselContinuityFail, deviceErrorVcselWatchdogFail, deviceErrorNoVHVValueFound,
		deviceErrorUserROIClip, deviceErrorMultClipFail:
		d.results.deviceStatus = s
		return false, nil
	}
	return true, nil
}

// binSequenceConfig is VL53LX_hist_get_bin_sequence_config. The ambient
// thresholds are always 0xFFFF in the presets.
func (d *Device) binSequenceConfig() {
	h := &d.hist
	c := &d.histCfg
	const ambThresh = 1024 * 0xFFFF
	if d.state.rdStreamCount&0x01 == 0 {
		h.binSeq = c.even
	} else {
		h.binSeq = c.odd
		if h.ambientEventsSum == ambThresh {
			// The mid ambient odd sequence stores bins 3 and 4 swapped.
			h.binSeq[3], h.binSeq[4] = c.odd[4], c.odd[3]
		}
	}
	for i := range h.binRep {
		h.binRep[i] = 1
	}
}

// histogramMerge is vl53lx_histo_merge.
func (d *Device) histogramMerge() {
	m := &d.merge
	h := &d.hist
	if m.waitBeforeNext != 0 {
		m.waitBeforeNext--
		return
	}

	timing := 1 - h.streamCount%2
	highIndex := histogramBufferSize - int(timing)*4
	var prevPos uint8
	if m.pos > 0 {
		prevPos = m.pos - 1
	} else {
		prevPos = mergeMaxSize - 1
	}

	var stddev int32
	if m.rec[prevPos][timing][4] > 0 {
		var totalPrev, totalCur int32
		for bin := int(timing) * 4; bin < highIndex; bin++ {
			totalPrev += m.rec[prevPos][timing][bin]
			totalCur += h.bins[bin]
		}
		if totalPrev != 0 && totalCur != 0 {
			for bin := int(timing) * 4; bin < highIndex; bin++ {
				prev := m.rec[prevPos][timing][bin] * 1000 / totalPrev
				cur := h.bins[bin] * 1000 / totalCur
				stddev += (prev - cur) * (prev - cur)
			}
		}
	}

	reset := false
	if stddev >= mergeResetThreshold {
		m.rec = [mergeMaxSize][2][histogramBufferSize]int32{}
		m.pos = 0
		reset = true
		if timing == 0 {
			m.waitBeforeNext = mergeFrameWaitEvent
		} else {
			m.waitBeforeNext = mergeFrameWaitEvent + 1
		}
	} else {
		m.rec[m.pos][timing] = h.bins
	}

	if m.pos == mergeMaxSize-1 && timing == 1 {
		m.pos = 0
	} else if timing == 1 {
		m.pos++
	}

	if !(reset && timing == 0) && m.waitBeforeNext == 0 {
		for bin := range h.bins {
			h.bins[bin] = 0
			for i := 0; i < mergeMaxSize; i++ {
				h.bins[bin] += m.rec[i][timing][bin]
			}
		}
	}
}

// histogramMergeCount is VL53LX_compute_histo_merge_nb.
func (d *Device) histogramMergeCount() uint8 {
	timing := 0
	if d.hist.binSeq[0] == 7 {
		timing = 1
	}
	var sum uint8
	for i := 0; i < mergeMaxSize; i++ {
		if d.merge.rec[i][timing][7] > 0 {
			sum++
		}
	}
	return sum
}

// copyAndScaleAmbientInfo is VL53LX_hist_copy_and_scale_ambient_info.
func copyAndScaleAmbientInfo(prev *zoneHistInfo, h *histogramData) bool {
	if prev.effectiveSpads == 0 {
		return false
	}
	if prev.numAmbientBins > 0 && h.numAmbientBins == 0 {
		tmpo := (1 + int64(h.totalPeriodsElapsed)) * int64(h.effectiveSpads)
		tmpi := (1 + int64(prev.totalPeriodsElapsed)) * int64(prev.effectiveSpads)
		evts := tmpo*int64(prev.ambientEventsSum) + tmpi/2
		if tmpi != 0 {
			evts /= tmpi
		}
		h.ambientEventsSum = int32(evts)
		h.ambientPerBin = (h.ambientEventsSum + int32(prev.numAmbientBins)/2) / int32(prev.numAmbientBins)
	}
	return true
}

// phaseConsistencyCheck is VL53LX_hist_phase_consistency_check. With ST's
// default tuning the event and merged pulse checks always pass.
func (d *Device) phaseConsistencyCheck() {
	prev := &d.prevObjects
	if prev.rdDeviceState != stateOutputData {
		return
	}
	const tolerance = uint16(phaseTolerance) << 8
	for i := uint8(0); i < d.results.activeResults; i++ {
		r := &d.results.data[i]
		if r.rangeStatus != deviceErrorRangeComplete && r.rangeStatus != deviceErrorRangeCompleteNoWrapCheck {
			continue
		}
		if prev.activeObjects == 0 {
			r.rangeStatus = deviceErrorPrevRangeNoTargets
		} else {
			r.rangeStatus = deviceErrorPhaseConsistency
		}
		for p := uint8(0); p < prev.activeObjects; p++ {
			var delta uint16
			if r.phase > prev.phase[p] {
				delta = r.phase - prev.phase[p]
			} else {
				delta = prev.phase[p] - r.phase
			}
			if delta < tolerance {
				r.rangeStatus = deviceErrorRangeComplete
			}
		}
	}
}

// fillMeasurement is SetMeasurementData and SetTargetData from ST's API.
func (d *Device) fillMeasurement(m *Measurement) {
	res := &d.results
	m.StreamCount = res.streamCount
	m.NumTargets = res.activeResults
	n := int(res.activeResults)
	if n < 1 {
		n = 1
	}
	for i := 0; i < n; i++ {
		r := &res.data[i]
		t := &m.Targets[i]
		t.SignalRate = rateToKcps(r.peakSignalRate)
		t.AmbientRate = rateToKcps(r.ambientRate)
		t.Distance = r.medianRangeMM
		t.DistanceMin = r.minRangeMM
		t.DistanceMax = r.maxRangeMM

		switch res.deviceStatus {
		case deviceErrorMultClipFail, deviceErrorVcselWatchdogFail, deviceErrorVcselContinuityFail,
			deviceErrorNoVHVValueFound:
			t.Status = HardwareFail
		case deviceErrorUserROIClip:
			t.Status = MinRangeFail
		default:
			t.Status = RangeValid
		}

		if t.Status == RangeValid && res.activeResults == 0 {
			t.Status = NoTarget
			t.SignalRate = 0
			t.Distance = 8191
			t.DistanceMin = 8191
			t.DistanceMax = 8191
		}
		if t.Status == RangeValid {
			t.Status = convertStatus(r.rangeStatus & 0x1F)
		}
		if t.Status == RangeValid && t.Distance < 0 {
			if t.Distance < proxyMinMM {
				t.Status = RangeInvalid
			} else {
				t.Distance = 0
			}
		}
	}
}

// rateToKcps converts a 9.7 fixed point MCPS rate.
func rateToKcps(rate uint16) uint32 {
	return (uint32(rate)*1000 + 64) >> 7
}

// convertStatus is ConvertStatusHisto from ST's API.
func convertStatus(s uint8) RangeStatus {
	switch s {
	case deviceErrorRangePhaseCheck:
		return OutOfBoundsFail
	case deviceErrorSigmaThresholdCheck:
		return SigmaFail
	case deviceErrorRangeCompleteNoWrapCheck:
		return RangeValidNoWrapCheckFail
	case deviceErrorPhaseConsistency, deviceErrorEventConsistency:
		return WrapTargetFail
	case deviceErrorPrevRangeNoTargets:
		return TargetPresentLackOfSignal
	case deviceErrorRangeCompleteMergedPulse:
		return RangeValidMergedPulse
	case deviceErrorRangeComplete:
		return RangeValid
	}
	return NoTarget
}

// processHistogram is VL53LX_hist_process_data with crosstalk compensation
// disabled followed by VL53LX_f_025, the gen4 algorithm.
func (d *Device) processHistogram() {
	averageHistogram(&d.hist, &d.averaged)

	a := &d.algo
	res := &d.results
	a.init()
	a.bins = d.averaged
	pB := &a.bins

	res.streamCount = pB.streamCount
	res.activeResults = 0

	calcZeroDistancePhase(pB)
	estimateAmbientFromAmbientBins(pB)
	removeAmbientBins(pB)

	a.findThresholds()
	a.findStartBin()
	a.countPulses()
	a.findPulseEdges()

	for p := uint8(0); p < a.pulseCount; p++ {
		pulse := &a.pulses[p]
		a.pulseEvents(p)
		a.padPulse(p, &a.pulseBins, pB.ambientPerBin)
		a.padPulse(p, &a.pulseBinsZero, 0)
		a.filterPulse(p, &d.filtered)
		a.detectPeak(p, &d.filtered)
		pulse.sigma = estimateSigma(pulse.peak, sigmaEstimatorRefMM, a.periodBins, pulse.filterWOI,
			&a.pulseBins, &a.pulseBinsZero)
		a.pulsePhaseRange(p)
	}

	a.sortPulses()

	for p := uint8(0); p < a.pulseCount; p++ {
		pulse := &a.pulses[p]
		if res.activeResults >= MaxTargets {
			continue
		}
		if pulse.signalEvents > signalTotalEventsLimit && pulse.peak < 0xFF {
			r := &res.data[res.activeResults]
			setRangeData(d.validPhaseLow, d.validPhaseHigh, sigmaThreshMM, pB, pulse, r)
			calcRates(pB.vcselWidth, pB.fastOscFrequency, pB.totalPeriodsElapsed, pB.effectiveSpads, r)
			calcRanges(histGainFactor, d.rangeOffsetMM, r)
			res.activeResults++
		}
	}
}

// init is VL53LX_f_003.
func (a *algoData) init() {
	a.binBufferSize = histogramBufferSize
	a.binStart = 0
	a.binCount = 0
	a.ambientPerBin = 0
	a.lastThreshold = 0
	a.aboveThreshold = [histogramBufferSize]uint8{}
	a.pulseIndex = [histogramBufferSize]uint8{}
	a.startBin = 0
	a.pulseCount = 0
}

// averageHistogram is VL53LX_f_031. It averages repeated bin sequences.
func averageHistogram(in, out *histogramData) {
	var initialIndex [maxBinSequenceCode + 1]uint8
	var repeatCount [maxBinSequenceCode + 1]uint8

	*out = *in
	out.binCount = 0
	for i := range out.binSeq {
		out.binSeq[i] = maxBinSequenceCode + 1
	}
	for i := 0; i < int(out.binBufferSize); i++ {
		out.bins[i] = 0
	}

	var seqLength uint8
	for lc := 0; lc < maxBinSequence; lc++ {
		cfg := in.binSeq[lc]
		if repeatCount[cfg] == 0 {
			initialIndex[cfg] = seqLength * 4
			out.binSeq[seqLength] = cfg
			seqLength++
		}
		repeatCount[cfg]++
		idx := initialIndex[cfg]
		for i := uint8(0); i < 4; i++ {
			out.bins[idx+i] += in.bins[lc*4+int(i)]
		}
	}

	for lc := range out.binSeq {
		cfg := out.binSeq[lc]
		if cfg <= maxBinSequenceCode {
			out.binRep[lc] = repeatCount[cfg]
		} else {
			out.binRep[lc] = 0
		}
	}
	out.binCount = seqLength * 4

	for lc := range repeatCount {
		n := int32(repeatCount[lc])
		if n > 0 {
			idx := initialIndex[lc]
			for i := uint8(0); i < 4; i++ {
				out.bins[idx+i] += n / 2
				out.bins[idx+i] /= n
			}
		}
	}

	out.numAmbientBins = 0
	if repeatCount[7] > 0 || repeatCount[15] > 0 {
		out.numAmbientBins = 4
	}
}

// calcZeroDistancePhase is VL53LX_hist_calc_zero_distance_phase.
func calcZeroDistancePhase(h *histogramData) {
	period := 2048 * uint32(decodeVcselPeriod(h.vcselPeriod))
	phase := period
	phase += uint32(h.phasecalRefPhase)
	phase += 2048 * uint32(h.phasecalVcselStart)
	phase -= 2048 * uint32(h.calVcselStart)
	h.zeroDistancePhase = uint16(phase % period)
}

// estimateAmbientFromAmbientBins is
// VL53LX_hist_estimate_ambient_from_ambient_bins.
func estimateAmbientFromAmbientBins(h *histogramData) {
	if h.numAmbientBins == 0 {
		return
	}
	h.numAmbientSamples = h.numAmbientBins
	h.ambientEventsSum = 0
	for bin := uint8(0); bin < h.numAmbientBins; bin++ {
		h.ambientEventsSum += h.bins[bin]
	}
	h.ambientPerBin = (h.ambientEventsSum + int32(h.numAmbientBins)/2) / int32(h.numAmbientBins)
}

// removeAmbientBins is VL53LX_hist_remove_ambient_bins.
func removeAmbientBins(h *histogramData) {
	if h.binSeq[0]&0x07 == 0x07 {
		i := 0
		for lc := 0; lc < maxBinSequence; lc++ {
			if h.binSeq[lc]&0x07 != 0x07 {
				h.binSeq[i] = h.binSeq[lc]
				h.binRep[i] = h.binRep[lc]
				i++
			}
		}
		for lc := i; lc < maxBinSequence; lc++ {
			h.binSeq[lc] = maxBinSequenceCode + 1
			h.binRep[lc] = 0
		}
	}
	if h.numAmbientBins > 0 {
		n := h.numAmbientBins
		for bin := n; bin < h.binBufferSize; bin++ {
			h.bins[bin-n] = h.bins[bin]
		}
		h.binCount -= n
		h.numAmbientBins = 0
	}
}

// findThresholds is VL53LX_f_006 without crosstalk.
func (a *algoData) findThresholds() {
	h := &a.bins
	a.binBufferSize = h.binBufferSize
	a.binStart = h.binStart
	a.binCount = h.binCount
	a.ambientPerBin = h.ambientPerBin
	a.periodBins = decodeVcselPeriod(h.vcselPeriod)

	ambEvents := int32((int64(h.ambientPerBin)*ambientThreshEventScaler + 2048) / 4096)

	for lb := uint8(0); lb < h.binCount; lb++ {
		samples := int32(h.binRep[lb>>2])
		if samples > 0 {
			t := int32(isqrt(uint32(samples * ambEvents)))
			t += samples / 2
			t /= samples
			t *= ambientThreshSigma1
			t += 8
			t /= 16
			t += ambEvents
			if t < minAmbientThreshEvents {
				t = minAmbientThreshEvents
			}
			a.threshold[lb] = t
			a.lastThreshold = t
		}
	}

	for lb := h.binStart; lb < h.binCount; lb++ {
		if h.bins[lb] > a.threshold[lb] {
			a.aboveThreshold[lb] = 1
		} else {
			a.aboveThreshold[lb] = 0
		}
	}
}

// findStartBin is VL53LX_f_007.
func (a *algoData) findStartBin() {
	a.startBin = 0
	found := false
	for i := uint8(0); i < a.periodBins; i++ {
		j := (i + 1) % a.periodBins
		if i < a.binCount && j < a.binCount {
			if a.aboveThreshold[i] == 0 && a.aboveThreshold[j] == 1 && !found {
				a.startBin = i
				found = true
			}
		}
	}
}

// countPulses is VL53LX_f_008.
func (a *algoData) countPulses() {
	for lb := a.startBin; lb < a.startBin+a.periodBins; lb++ {
		i := lb % a.periodBins
		j := (lb + 1) % a.periodBins
		if i < a.binCount && j < a.binCount {
			if a.aboveThreshold[i] == 0 && a.aboveThreshold[j] == 1 {
				a.pulseCount++
			}
			if a.aboveThreshold[i] > 0 {
				a.pulseIndex[i] = a.pulseCount
			} else {
				a.pulseIndex[i] = 0
			}
		}
	}
	if a.pulseCount > maxPulses {
		a.pulseCount = maxPulses
	}
}

// findPulseEdges is VL53LX_f_009.
func (a *algoData) findPulseEdges() {
	maxHalfWidth := (a.periodBins - 1) >> 1
	for blb := a.startBin; blb < a.startBin+a.periodBins; blb++ {
		i := blb % a.periodBins
		j := (blb + 1) % a.periodBins
		if i >= a.binCount || j >= a.binCount {
			continue
		}
		if a.pulseIndex[i] == 0 && a.pulseIndex[j] > 0 {
			n := a.pulseIndex[j] - 1
			if n < maxPulses {
				p := &a.pulses[n]
				p.start = blb
				p.first = blb + 1
				p.peak = 0xFF
				p.last = 0
				p.end = 0
			}
		}
		if a.pulseIndex[i] > 0 && a.pulseIndex[j] == 0 {
			n := a.pulseIndex[i] - 1
			if n < maxPulses {
				p := &a.pulses[n]
				p.last = blb
				p.end = blb + 1
				p.width = (p.last + 1) - p.first
				p.filterWOI = (p.end + 1) - p.start
				if p.filterWOI > maxHalfWidth {
					p.filterWOI = maxHalfWidth
				}
			}
		}
	}
}

// pulseEvents is VL53LX_f_010.
func (a *algoData) pulseEvents(n uint8) {
	p := &a.pulses[n]
	p.totalEvents = 0
	p.ambientEvents = 0
	for lb := p.start; lb <= p.end; lb++ {
		i := lb % a.periodBins
		p.totalEvents += a.bins.bins[i]
		p.ambientEvents += a.ambientPerBin
	}
	p.signalEvents = p.totalEvents - p.ambientEvents
}

// padPulse is VL53LX_f_011. It copies the bins and replaces the bins
// outside the pulse with the pad value.
func (a *algoData) padPulse(n uint8, out *histogramData, pad int32) {
	p := &a.pulses[n]
	*out = a.bins
	for lb := a.startBin; lb < a.startBin+a.periodBins; lb++ {
		if lb < p.start || lb > p.end {
			i := lb % a.periodBins
			if i < out.binCount {
				out.bins[i] = pad
			}
		}
	}
}

// filterSums is VL53LX_f_022.
func filterSums(bin, woi uint8, h *histogramData) (sa, sb, sc int32) {
	sb = h.bins[bin]
	for w := uint8(0); w < (woi<<1)+1; w++ {
		j := ((bin + w + h.binCount) - woi) % h.binCount
		if w < woi {
			sa += h.bins[j]
		} else if w > woi {
			sc += h.bins[j]
		}
	}
	return
}

// filterPulse is VL53LX_f_026.
func (a *algoData) filterPulse(n uint8, f *filteredData) {
	p := &a.pulses[n]
	for lb := p.start; lb <= p.end; lb++ {
		i := lb % a.periodBins
		sa, sb, sc := filterSums(i, p.filterWOI, &a.pulseBins)
		f.a[i] = sa
		f.b[i] = sb
		f.c[i] = sc
		f.diff1[i] = (sa + sb) - (sc + a.ambientPerBin)
		f.diff2[i] = (sb + sc) - (sa + a.ambientPerBin)
	}
}

// detectPeak is VL53LX_f_027.
func (a *algoData) detectPeak(n uint8, f *filteredData) {
	p := &a.pulses[n]
	for lb := p.start; lb < p.end; lb++ {
		i := lb % a.periodBins
		j := (lb + 1) % a.periodBins
		if i >= a.binCount || j >= a.binCount {
			continue
		}
		switch {
		case f.diff1[i] == 0 && f.diff2[i] == 0:
			f.valid[i] = 0
		case f.diff1[i] >= 0 && f.diff2[i] >= 0:
			f.valid[i] = 1
		case f.diff1[i] < 0 && f.diff2[i] >= 0 && f.diff1[j] >= 0 && f.diff2[j] < 0:
			f.valid[i] = 1
		default:
			f.valid[i] = 0
		}
		if f.valid[i] > 0 {
			p.peak = lb
			var ok bool
			p.phase, ok = meanPhase(lb, f.a[i], f.b[i], f.c[i], a.ambientPerBin, a.periodBins)
			if !ok {
				f.valid[i] = 0
			}
		}
	}
}

// meanPhase is VL53LX_f_028 without crosstalk.
func meanPhase(bin uint8, sa, sb, sc, ambient int32, periodBins uint8) (uint32, bool) {
	num := 4096 * (int64(sc) - int64(sa))
	bMinusAmb := 2 * 4096 * (int64(sb) - int64(ambient))
	if bMinusAmb == 0 {
		return maxAllowedPhase, false
	}
	phase := (4096*num + bMinusAmb/2) / bMinusAmb
	phase += 2048
	phase += 4096 * int64(bin)
	phase = (phase + 1) / 2
	if phase < 0 {
		phase = 0
	}
	if phase > maxAllowedPhase {
		phase = maxAllowedPhase
	}
	phase %= int64(periodBins) * 2048
	return uint32(phase), true
}

// estimateSigma is VL53LX_f_014 and VL53LX_f_023 without crosstalk.
func estimateSigma(bin, sigmaRefMM, periodBins, woi uint8, ap, zp *histogramData) uint16 {
	i := bin % periodBins
	aZp, _, cZp := filterSums(i, woi, zp)
	sa, sb, sc := filterSums(i, woi, ap)

	const (
		d003 = 0xFFFFFF
		d004 = 0xFFFFFFFFFFFFFF
		d005 = 0x7FFFFFFFFF
		d006 = 0x7FFFFFFFFFFFFFFF
		d007 = 0xFFFFFFFF
	)

	freq := ap.fastOscFrequency
	if freq == 0 {
		return 0xFFFF
	}
	pllPeriodMM := uint64(calcPLLPeriodMM(freq))
	amb := uint64(uint32(ap.ambientPerBin))
	b := uint64(uint32(sb))
	va := uint64(uint32(sa))
	vc := uint64(uint32(sc))

	var bMinusAmb, diff uint64
	if amb > b {
		bMinusAmb = amb - b
	} else {
		bMinusAmb = b - amb
	}
	if va > vc {
		diff = va - vc
	} else {
		diff = vc - va
	}
	if bMinusAmb == 0 {
		return 0xFFFF
	}

	tmp0 := b + amb
	if tmp0 > d003 {
		tmp0 = d003
	}
	tmp1 := diff * diff
	tmp1 <<= 8
	if tmp1 > d004 {
		tmp1 = d004
	}
	tmp1 /= bMinusAmb
	tmp1 /= bMinusAmb
	if tmp1 > d005 {
		tmp1 = d005
	}
	tmp0 = tmp1 * tmp0

	tmp1 = uint64(uint32(cZp)) + uint64(uint32(aZp))
	if tmp1 > d003 {
		tmp1 = d003
	}
	tmp1 <<= 8
	tmp0 += tmp1
	if tmp0 > d006 {
		tmp0 = d006
	}
	if tmp0 > d007 {
		tmp0 /= bMinusAmb
		tmp0 *= pllPeriodMM
	} else {
		tmp0 *= pllPeriodMM
		tmp0 /= bMinusAmb
	}
	if tmp0 > d006 {
		tmp0 = d006
	}
	if tmp0 > d007 {
		tmp0 /= bMinusAmb
		tmp0 /= 4
		tmp0 *= pllPeriodMM
	} else {
		tmp0 *= pllPeriodMM
		tmp0 /= bMinusAmb
		tmp0 /= 4
	}
	if tmp0 > d006 {
		tmp0 = d006
	}
	tmp0 >>= 2
	if tmp0 > d007 {
		tmp0 = d007
	}
	tmp1 = uint64(sigmaRefMM) << 7
	tmp1 *= tmp1
	tmp0 += tmp1
	if tmp0 > d007 {
		tmp0 = d007
	}
	return uint16(isqrt(uint32(tmp0)))
}

// pulsePhaseRange is VL53LX_f_015 with clipping enabled.
func (a *algoData) pulsePhaseRange(n uint8) {
	p := &a.pulses[n]
	i := int16(p.peak % a.periodBins)
	start := i + int16(p.start) - int16(p.peak)
	end := i + int16(p.end) - int16(p.peak)
	width := end - start
	if width > 3 {
		width = 3
	}
	p.phaseLow = a.windowPhase(start, start+width)
	p.phaseHigh = a.windowPhase(end-width, end)
	if p.phaseLow > p.phaseHigh {
		p.phaseLow, p.phaseHigh = p.phaseHigh, p.phaseLow
	}
	if p.phase < p.phaseLow {
		p.phaseLow = p.phase
	}
	if p.phase > p.phaseHigh {
		p.phaseHigh = p.phase
	}
}

// windowPhase is VL53LX_f_020 with clipping enabled.
func (a *algoData) windowPhase(from, to int16) uint32 {
	var eventSum, weightedSum int64
	period := int16(a.periodBins)
	for lb := from; lb <= to; lb++ {
		var i int16
		if lb < 0 {
			i = lb + period
		} else {
			i = lb % period
		}
		v := int64(a.bins.bins[i]) - int64(a.bins.ambientPerBin)
		if v < 0 {
			v = 0
		}
		eventSum += v
		weightedSum += v * (1024 + 2048*int64(lb))
	}
	if eventSum <= 0 {
		return maxAllowedPhase
	}
	weightedSum += eventSum / 2
	weightedSum /= eventSum
	if weightedSum < 0 {
		weightedSum = 0
	}
	return uint32(weightedSum)
}

// sortPulses is VL53LX_f_016 ordered by increasing distance.
func (a *algoData) sortPulses() {
	if a.pulseCount <= 1 {
		return
	}
	for swapped := true; swapped; {
		swapped = false
		for i := uint8(1); i < a.pulseCount; i++ {
			if a.pulses[i-1].phase > a.pulses[i].phase {
				a.pulses[i-1], a.pulses[i] = a.pulses[i], a.pulses[i-1]
				swapped = true
			}
		}
	}
}

// setRangeData is VL53LX_f_017.
func setRangeData(validPhaseLow, validPhaseHigh uint8, sigmaThresh uint16, h *histogramData, p *pulseData, r *rangeData) {
	r.woiWidth = (p.end + 1) - p.start
	r.zeroDistancePhase = h.zeroDistancePhase
	r.sigma = p.sigma
	r.phaseLow = uint16(p.phaseLow)
	r.phase = uint16(p.phase)
	r.phaseHigh = uint16(p.phaseHigh)
	r.totalEvents = uint32(p.totalEvents)
	r.signalEvents = p.signalEvents
	r.ambientEvents = uint32(p.ambientEvents)
	r.totalPeriodsElapsed = h.totalPeriodsElapsed
	r.rangeStatus = deviceErrorRangeCompleteNoWrapCheck

	if sigmaThresh > 0 && uint32(p.sigma) > uint32(sigmaThresh)<<5 {
		r.rangeStatus = deviceErrorSigmaThresholdCheck
	}

	lower := uint16(validPhaseLow) << 8
	if lower < r.zeroDistancePhase {
		lower = r.zeroDistancePhase - lower
	} else {
		lower = 0
	}
	upper := uint16(validPhaseHigh) << 8
	upper += h.zeroDistancePhase
	if r.phase < lower || r.phase > upper {
		r.rangeStatus = deviceErrorRangePhaseCheck
	}
}

// calcRates is VL53LX_f_018.
func calcRates(vcselWidth, fastOscFrequency uint16, totalPeriodsElapsed uint32, effectiveSpads uint16, r *rangeData) {
	r.vcselWidth = vcselWidth
	r.fastOscFrequency = fastOscFrequency
	r.totalPeriodsElapsed = totalPeriodsElapsed
	r.effectiveSpads = effectiveSpads
	if fastOscFrequency == 0 || totalPeriodsElapsed == 0 {
		return
	}
	pll := calcPLLPeriodUs(fastOscFrequency)
	periods := totalPeriodsElapsed + 1
	r.peakDurationUs = durationMaths(pll, uint32(vcselWidth), rangingWindowVcselPeriods, periods)
	r.woiDurationUs = durationMaths(pll, uint32(r.woiWidth)<<4, rangingWindowVcselPeriods, periods)
	r.peakSignalRate = rateMaths(r.signalEvents, r.peakDurationUs)
	r.avgSignalRate = rateMaths(r.signalEvents, r.woiDurationUs)
	r.ambientRate = rateMaths(int32(r.ambientEvents), r.woiDurationUs)
	total := uint32(r.peakSignalRate) + uint32(r.ambientRate)
	r.totalRatePerSpd = ratePerSpadMaths(0x06, total, effectiveSpads, 0xFFFF)
	r.signalPerSpad = eventsPerSpadMaths(r.signalEvents, effectiveSpads, r.peakDurationUs)
}

// calcRanges is VL53LX_f_019.
func calcRanges(gainFactor uint16, offsetMM int16, r *rangeData) {
	r.minRangeMM = int16(rangeMaths(r.fastOscFrequency, r.phaseLow, r.zeroDistancePhase, int32(gainFactor), int32(offsetMM)))
	r.medianRangeMM = int16(rangeMaths(r.fastOscFrequency, r.phase, r.zeroDistancePhase, int32(gainFactor), int32(offsetMM)))
	r.maxRangeMM = int16(rangeMaths(r.fastOscFrequency, r.phaseHigh, r.zeroDistancePhase, int32(gainFactor), int32(offsetMM)))
}

// durationMaths is VL53LX_duration_maths.
func durationMaths(pllPeriodUs, vcselParmPclks, windowVclks, elapsedMclks uint32) uint32 {
	d := windowVclks * pllPeriodUs
	d >>= 12
	tmp := uint64(d)
	d = elapsedMclks * vcselParmPclks
	d >>= 4
	tmp *= uint64(d)
	tmp >>= 12
	if tmp > 0xFFFFFFFF {
		tmp = 0xFFFFFFFF
	}
	return uint32(tmp)
}

// rateMaths is VL53LX_rate_maths. It returns a 9.7 fixed point MCPS rate.
func rateMaths(events int32, timeUs uint32) uint16 {
	const (
		totalCountMax      = 1<<29 - 1
		totalCountResThres = 1 << 24
	)
	var tmp uint32
	if events > totalCountMax {
		tmp = totalCountMax
	} else if events > 0 {
		tmp = uint32(events)
	}
	fracBits := uint32(7)
	if events > totalCountResThres {
		fracBits = 3
	}
	if timeUs > 0 {
		tmp = ((tmp << fracBits) + timeUs/2) / timeUs
	}
	if events > totalCountResThres {
		tmp <<= 4
	}
	if tmp > 0xFFFF {
		tmp = 0xFFFF
	}
	return uint16(tmp)
}

// ratePerSpadMaths is VL53LX_rate_per_spad_maths.
func ratePerSpadMaths(fracBits, peakCountRate uint32, numSpads uint16, maxOutput uint32) uint16 {
	var tmp uint32
	if numSpads > 0 {
		tmp = (peakCountRate << 8) << fracBits
		tmp = (tmp + uint32(numSpads)/2) / uint32(numSpads)
	} else {
		tmp = peakCountRate << fracBits
	}
	if tmp > maxOutput {
		tmp = maxOutput
	}
	return uint16(tmp)
}

// eventsPerSpadMaths is VL53LX_events_per_spad_maths.
func eventsPerSpadMaths(events int32, numSpads uint16, duration uint32) uint32 {
	var total uint64
	dividend := uint64(events) * 1000 * 256
	if numSpads != 0 {
		total = dividend / uint64(numSpads)
	}
	if duration > 0 {
		return uint32(((total << 11) + uint64(duration)/2) / uint64(duration))
	}
	return uint32(total << 11)
}

// rangeMaths is VL53LX_range_maths with no fractional bits.
func rangeMaths(fastOscFrequency, phase, zeroDistancePhase uint16, gainFactor, offsetMM int32) int32 {
	pll := calcPLLPeriodUs(fastOscFrequency)
	tmp := int64(phase) - int64(zeroDistancePhase)
	tmp *= int64(pll)
	tmp /= 1 << 9
	tmp *= speedOfLightInAirDiv8
	tmp /= 1 << 22

	rangeMM := int32(tmp) + offsetMM
	rangeMM *= gainFactor
	rangeMM += 0x0400
	rangeMM /= 0x0800

	r10 := rangeMM * 10 / 4
	if r10%10 < 5 {
		return int32(int16(r10 / 10))
	}
	return int32(int16(r10/10 + 1))
}

// calcPLLPeriodMM is VL53LX_calc_pll_period_mm.
func calcPLLPeriodMM(fastOscFrequency uint16) uint32 {
	pll := calcPLLPeriodUs(fastOscFrequency)
	mm := speedOfLightInAirDiv8 * (pll >> 2)
	return (mm + 1<<15) >> 16
}

// isqrt is VL53LX_isqrt.
func isqrt(num uint32) uint32 {
	var res uint32
	bit := uint32(1) << 30
	for bit > num {
		bit >>= 2
	}
	for bit != 0 {
		if num >= res+bit {
			num -= res + bit
			res = (res >> 1) + bit
		} else {
			res >>= 1
		}
		bit >>= 2
	}
	return res
}

// combineMM1MM2Offsets is VL53LX_hist_combine_mm1_mm2_offsets.
func combineMM1MM2Offsets(mm1OffsetMM, mm2OffsetMM int16, mmROICentre, mmROISize, zoneCentre, zoneSize uint8,
	cal *additionalOffsetCal, goodSpads *[32]uint8, apertureAttenuation uint16) int16 {
	maxInner, maxOuter := calcMMEffectiveSpads(mmROICentre, mmROISize, 0xC7, 0xFF, goodSpads, apertureAttenuation)
	if maxInner == 0 || maxOuter == 0 {
		return 0
	}
	inner, outer := calcMMEffectiveSpads(mmROICentre, mmROISize, zoneCentre, zoneSize, goodSpads, apertureAttenuation)

	mm1Rate := uint32(cal.mmInnerPeakRate) * uint32(inner) / uint32(maxInner)
	mm2Rate := uint32(cal.mmOuterPeakRate) * uint32(outer) / uint32(maxOuter)

	tmp0 := int32(mm1OffsetMM)*int32(mm1Rate) + int32(mm2OffsetMM)*int32(mm2Rate)
	tmp1 := int32(mm1Rate) + int32(mm2Rate)
	if tmp1 != 0 {
		tmp0 = (tmp0 * 4) / tmp1
	}
	return int16(tmp0)
}

// calcMMEffectiveSpads is VL53LX_calc_mm_effective_spads.
func calcMMEffectiveSpads(mmROICentre, mmROISize, zoneCentre, zoneSize uint8, goodSpads *[32]uint8,
	apertureAttenuation uint16) (inner, outer uint16) {
	mmXLL, mmYLL, mmXUR, mmYUR := decodeZoneLimits(mmROICentre, mmROISize)
	zXLL, zYLL, zXUR, zYUR := decodeZoneLimits(zoneCentre, zoneSize)

	for y := zYLL; y <= zYUR; y++ {
		for x := zXLL; x <= zXUR; x++ {
			spad := encodeRowCol(uint8(y), uint8(x))
			if goodSpads[spad>>3]&(1<<(spad&0x07)) == 0 {
				continue
			}
			att := uint16(0x0100)
			if isApertureLocation(uint8(y), uint8(x)) {
				att = apertureAttenuation
			}
			if x >= mmXLL && x <= mmXUR && y >= mmYLL && y <= mmYUR {
				inner += att
			} else {
				outer += att
			}
		}
	}
	return
}

func encodeRowCol(row, col uint8) uint8 {
	if row > 7 {
		return 128 + (col << 3) + (15 - row)
	}
	return ((15 - col) << 3) + row
}

func decodeRowCol(spad uint8) (row, col uint8) {
	if spad > 127 {
		return 8 + ((255 - spad) & 0x07), (spad - 128) >> 3
	}
	return spad & 0x07, (127 - spad) >> 3
}

// decodeZoneLimits is VL53LX_decode_zone_limits.
func decodeZoneLimits(centre, size uint8) (xLL, yLL, xUR, yUR int16) {
	yCentre, xCentre := decodeRowCol(centre)
	height := int16(size >> 4)
	width := int16(size & 0x0F)

	xLL = int16(xCentre) - (width+1)/2
	if xLL < 0 {
		xLL = 0
	}
	xUR = xLL + width
	if xUR > 15 {
		xUR = 15
	}
	yLL = int16(yCentre) - (height+1)/2
	if yLL < 0 {
		yLL = 0
	}
	yUR = yLL + height
	if yUR > 15 {
		yUR = 15
	}
	return
}

func isApertureLocation(row, col uint8) bool {
	r := row % 4
	c := col % 4
	return (r == 0 && c == 2) || (r == 2 && c == 0)
}
