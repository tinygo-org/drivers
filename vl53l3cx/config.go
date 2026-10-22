package vl53l3cx

import "encoding/binary"

type staticConfig struct {
	dssTargetRate          uint16
	gpioHVMuxCtrl          uint8
	gpioTioHVStatus        uint8
	spadSelPswidth         uint8
	vcselPulseWidthOffset  uint8
	sigmaEstPulseWidth     uint8
	sigmaEstAmbientWidth   uint8
	sigmaRefMM             uint8
	xtalkValidHeight       uint8
	spare0                 uint8
	spare1                 uint8
	rangeIgnoreThreshold   uint16
	rangeIgnoreValidHeight uint8
	rangeMinClip           uint8
	consistencyTolerance   uint8
	spare2                 uint8
	resetStagesMSB         uint8
	resetStagesLSB         uint8
}

func (c *staticConfig) encode(b []byte) {
	binary.BigEndian.PutUint16(b[0:], c.dssTargetRate)
	b[12] = c.gpioHVMuxCtrl & 0x1F
	b[13] = c.gpioTioHVStatus & 0x3
	b[15] = c.spadSelPswidth & 0x7
	b[16] = c.vcselPulseWidthOffset & 0x1F
	b[18] = c.sigmaEstPulseWidth
	b[19] = c.sigmaEstAmbientWidth
	b[20] = c.sigmaRefMM
	b[21] = c.xtalkValidHeight
	b[22] = c.spare0
	b[23] = c.spare1
	binary.BigEndian.PutUint16(b[24:], c.rangeIgnoreThreshold)
	b[26] = c.rangeIgnoreValidHeight
	b[27] = c.rangeMinClip
	b[28] = c.consistencyTolerance & 0xF
	b[29] = c.spare2
	b[30] = c.resetStagesMSB & 0xF
	b[31] = c.resetStagesLSB
}

type generalConfig struct {
	interruptConfigGPIO    uint8
	calVcselStart          uint8
	vcselWidth             uint8
	phasecalTimeoutMacrop  uint8
	phasecalTarget         uint8
	dssROIModeControl      uint8
	dssManualEffSpads      uint16
	dssApertureAttenuation uint8
	dssMaxSpadsLimit       uint8
	dssMinSpadsLimit       uint8
}

func (c *generalConfig) encode(b []byte) {
	b[2] = c.interruptConfigGPIO
	b[3] = c.calVcselStart & 0x7F
	b[6] = c.vcselWidth & 0x7F
	b[7] = c.phasecalTimeoutMacrop
	b[8] = c.phasecalTarget
	b[11] = c.dssROIModeControl & 0x7
	binary.BigEndian.PutUint16(b[16:], c.dssManualEffSpads)
	b[19] = c.dssApertureAttenuation
	b[20] = c.dssMaxSpadsLimit
	b[21] = c.dssMinSpadsLimit
}

type timingConfig struct {
	mmTimeoutA             uint16
	mmTimeoutB             uint16
	rangeTimeoutA          uint16
	rangeTimeoutB          uint16
	vcselPeriodA           uint8
	vcselPeriodB           uint8
	sigmaThresh            uint16
	minCountRateRtnLimit   uint16
	validPhaseLow          uint8
	validPhaseHigh         uint8
	intermeasurementPeriod uint32
}

func (c *timingConfig) encode(b []byte) {
	b[0] = uint8(c.mmTimeoutA>>8) & 0xF
	b[1] = uint8(c.mmTimeoutA)
	b[2] = uint8(c.mmTimeoutB>>8) & 0xF
	b[3] = uint8(c.mmTimeoutB)
	b[4] = uint8(c.rangeTimeoutA>>8) & 0xF
	b[5] = uint8(c.rangeTimeoutA)
	b[6] = c.vcselPeriodA & 0x3F
	b[7] = uint8(c.rangeTimeoutB>>8) & 0xF
	b[8] = uint8(c.rangeTimeoutB)
	b[9] = c.vcselPeriodB & 0x3F
	binary.BigEndian.PutUint16(b[10:], c.sigmaThresh)
	binary.BigEndian.PutUint16(b[12:], c.minCountRateRtnLimit)
	b[14] = c.validPhaseLow
	b[15] = c.validPhaseHigh
	binary.BigEndian.PutUint32(b[18:], c.intermeasurementPeriod)
}

type dynamicConfig struct {
	gph0               uint8
	threshHigh         uint16
	threshLow          uint16
	seedConfig         uint8
	woiSD0             uint8
	woiSD1             uint8
	initialPhaseSD0    uint8
	initialPhaseSD1    uint8
	gph1               uint8
	quantifier         uint8
	userROICentreSpad  uint8
	userROIRequestedXY uint8
	sequenceConfig     uint8
	groupedParamHold   uint8
}

func (c *dynamicConfig) encode(b []byte) {
	b[0] = c.gph0 & 0x3
	binary.BigEndian.PutUint16(b[1:], c.threshHigh)
	binary.BigEndian.PutUint16(b[3:], c.threshLow)
	b[6] = c.seedConfig & 0x7
	b[7] = c.woiSD0
	b[8] = c.woiSD1
	b[9] = c.initialPhaseSD0 & 0x7F
	b[10] = c.initialPhaseSD1 & 0x7F
	b[11] = c.gph1 & 0x3
	b[13] = c.quantifier & 0xF
	b[14] = c.userROICentreSpad
	b[15] = c.userROIRequestedXY
	b[16] = c.sequenceConfig
	b[17] = c.groupedParamHold & 0x3
}

type systemControl struct {
	powerForce     uint8
	firmwareEnable uint8
	interruptClear uint8
	modeStart      uint8
}

func (c *systemControl) encode(b []byte) {
	b[0] = c.powerForce & 0x1
	b[2] = c.firmwareEnable & 0x1
	b[3] = c.interruptClear & 0x3
	b[4] = c.modeStart
}

// Masks applied by ST's API when it writes back the NVM managed blocks.
var (
	staticNVMMasks   = [staticNVMSize]uint8{0x7F, 0x0F, 0x7F, 0x03, 0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0x3F, 0xFF}
	customerNVMMasks = [customerNVMSize]uint8{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x0F, 0xFF, 0x3F, 0x03,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x1F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
)

// histogramConfig holds the bin sequence. ST's API always sets the low, mid
// and high ambient sequences to the same values.
type histogramConfig struct {
	even [6]uint8
	odd  [6]uint8
}

func (h *histogramConfig) init(even, odd [6]uint8) {
	h.even = even
	h.odd = odd
}

func pack(lo, hi uint8) uint8 {
	return hi<<4 + lo
}

// copyToStatic is VL53LX_copy_hist_cfg_to_static_cfg.
func (d *Device) copyHistogramConfigToStatic() {
	h := &d.histCfg
	s := &d.staticCfg
	t := &d.timingCfg
	s.sigmaEstPulseWidth = pack(h.even[0], h.even[1])
	s.sigmaEstAmbientWidth = pack(h.even[2], h.even[3])
	s.sigmaRefMM = pack(h.even[4], h.even[5])
	s.xtalkValidHeight = pack(h.odd[0], h.odd[1])
	s.spare0 = pack(h.odd[2], h.odd[3])
	s.spare1 = pack(h.odd[4], h.odd[5])
	s.rangeIgnoreThreshold = uint16(pack(h.even[0], h.even[1]))<<8 + uint16(pack(h.even[2], h.even[3]))
	s.rangeIgnoreValidHeight = pack(h.even[4], h.even[5])
	s.rangeMinClip = pack(h.odd[0], h.odd[1])
	s.consistencyTolerance = h.odd[2]
	s.spare2 = pack(h.odd[3], h.odd[4])
	s.resetStagesMSB = h.odd[5]
	s.resetStagesLSB = 0
	t.sigmaThresh = uint16(pack(h.even[0], h.even[1]))<<8 + uint16(pack(h.even[2], h.even[3]))
	t.minCountRateRtnLimit = uint16(pack(h.even[4], h.even[5]))<<8 + uint16(pack(h.odd[0], h.odd[1]))
	t.validPhaseLow = pack(h.odd[2], h.odd[3])
	t.validPhaseHigh = pack(h.odd[4], h.odd[5])
	d.dynamicCfg.threshHigh = 0xFFFF
	d.dynamicCfg.threshLow = 0xFFFF
}

// presetHistogram combines VL53LX_preset_mode_standard_ranging,
// VL53LX_preset_mode_histogram_ranging and the long, medium and short
// range histogram presets.
func (d *Device) presetHistogram(mode DistanceMode) {
	d.staticCfg = staticConfig{
		gpioHVMuxCtrl:         0x10 | 0x01,
		gpioTioHVStatus:       0x02,
		spadSelPswidth:        0x02,
		vcselPulseWidthOffset: 0x08,
	}
	d.generalCfg = generalConfig{
		interruptConfigGPIO:    0x20,
		vcselWidth:             0x02,
		phasecalTimeoutMacrop:  0xF5,
		phasecalTarget:         33,
		dssROIModeControl:      1,
		dssManualEffSpads:      0x8C00,
		dssApertureAttenuation: 0x38,
		dssMaxSpadsLimit:       0xFF,
		dssMinSpadsLimit:       0x01,
	}
	d.timingCfg = timingConfig{}
	d.dynamicCfg = dynamicConfig{
		gph0:               0x01,
		seedConfig:         2,
		gph1:               0x01,
		quantifier:         2,
		userROICentreSpad:  0xC7,
		userROIRequestedXY: 0xFF,
		groupedParamHold:   0x02,
	}
	d.sysCtrl = systemControl{
		firmwareEnable: 0x01,
		interruptClear: 0x01,
		modeStart:      schedulerModeHistogram | readoutModeDualSD | measurementModeBack2B,
	}

	d.dynamicCfg.sequenceConfig = sequenceVHV | sequencePhasecal | sequenceDSS1 | sequenceDSS2 | sequenceRange
	d.validPhaseLow = 0x08

	switch mode {
	case Short:
		d.histCfg.init([6]uint8{7, 7, 0, 1, 1, 1}, [6]uint8{0, 1, 1, 1, 2, 2})
		d.timingCfg.vcselPeriodA = 0x03
		d.timingCfg.vcselPeriodB = 0x05
		d.generalCfg.calVcselStart = 0x03
		d.dynamicCfg.woiSD0 = 0x03
		d.dynamicCfg.woiSD1 = 0x05
		d.dynamicCfg.initialPhaseSD0 = 3
		d.dynamicCfg.initialPhaseSD1 = 6
		d.validPhaseHigh = 0x28
		d.dynamicCfg.sequenceConfig |= sequenceMM1
	case Medium:
		d.histCfg.init([6]uint8{7, 0, 1, 1, 2, 2}, [6]uint8{0, 1, 2, 1, 2, 3})
		d.timingCfg.vcselPeriodA = 0x05
		d.timingCfg.vcselPeriodB = 0x07
		d.generalCfg.calVcselStart = 0x05
		d.dynamicCfg.woiSD0 = 0x05
		d.dynamicCfg.woiSD1 = 0x07
		d.dynamicCfg.initialPhaseSD0 = 5
		d.dynamicCfg.initialPhaseSD1 = 6
		d.validPhaseHigh = 0x48
	default:
		d.histCfg.init([6]uint8{7, 0, 1, 2, 3, 4}, [6]uint8{0, 1, 2, 3, 4, 5})
		d.timingCfg.vcselPeriodA = 0x09
		d.timingCfg.vcselPeriodB = 0x0B
		d.generalCfg.calVcselStart = 0x09
		d.dynamicCfg.woiSD0 = 0x09
		d.dynamicCfg.woiSD1 = 0x0B
		d.dynamicCfg.initialPhaseSD0 = 9
		d.dynamicCfg.initialPhaseSD1 = 6
		d.validPhaseHigh = 0x88
	}
	d.copyHistogramConfigToStatic()
}

// setPresetMode is VL53LX_set_preset_mode for the histogram modes.
func (d *Device) setPresetMode(mode DistanceMode, phasecalUs, mmUs, rangeUs, imp uint32) {
	d.initState(stateSWStandby)
	d.presetHistogram(mode)
	d.staticCfg.dssTargetRate = 0x1400
	d.setTimeouts(phasecalUs, mmUs, rangeUs)
	d.timingCfg.intermeasurementPeriod = imp * uint32(d.oscCalibrateVal)
	d.prevObjects = zoneObjects{rdDeviceState: stateSWStandby}
}

// presetPhasecalTimeout returns the phasecal timeout from ST's tuning
// defaults for each distance mode.
func presetPhasecalTimeout(mode DistanceMode) uint32 {
	switch mode {
	case Short:
		return 6000
	case Medium:
		return 9000
	default:
		return 15000
	}
}

const (
	presetMMTimeoutUs    = 2000
	presetRangeTimeoutUs = 2500
)

func calcPLLPeriodUs(fastOscFrequency uint16) uint32 {
	if fastOscFrequency == 0 {
		return 0
	}
	return (1 << 30) / uint32(fastOscFrequency)
}

func decodeVcselPeriod(reg uint8) uint8 {
	return (reg + 1) << 1
}

func calcMacroPeriodUs(fastOscFrequency uint16, vcselPeriod uint8) uint32 {
	pll := calcPLLPeriodUs(fastOscFrequency)
	macro := uint32(2048+256) * pll
	macro >>= 6
	macro *= uint32(decodeVcselPeriod(vcselPeriod))
	return macro >> 6
}

func calcTimeoutMclks(timeoutUs, macroPeriodUs uint32) uint32 {
	return ((timeoutUs << 12) + (macroPeriodUs >> 1)) / macroPeriodUs
}

func encodeTimeout(mclks uint32) uint16 {
	if mclks == 0 {
		return 0
	}
	ls := mclks - 1
	var ms uint16
	for ls&0xFFFFFF00 > 0 {
		ls >>= 1
		ms++
	}
	return ms<<8 + uint16(ls&0xFF)
}

func decodeTimeout(encoded uint16) uint32 {
	return uint32(encoded&0xFF)<<uint32(encoded>>8) + 1
}

func calcTimeoutUs(mclks, macroPeriodUs uint32) uint32 {
	tmp := uint64(mclks) * uint64(macroPeriodUs)
	tmp += 0x800
	return uint32(tmp >> 12)
}

// setTimeouts is VL53LX_set_timeouts_us.
func (d *Device) setTimeouts(phasecalUs, mmUs, rangeUs uint32) {
	freq := d.fastOscFrequency()
	t := &d.timingCfg
	macro := calcMacroPeriodUs(freq, t.vcselPeriodA)
	mclks := calcTimeoutMclks(phasecalUs, macro)
	if mclks > 0xFF {
		mclks = 0xFF
	}
	d.generalCfg.phasecalTimeoutMacrop = uint8(mclks)
	t.mmTimeoutA = encodeTimeout(calcTimeoutMclks(mmUs, macro))
	t.rangeTimeoutA = encodeTimeout(calcTimeoutMclks(rangeUs, macro))
	macro = calcMacroPeriodUs(freq, t.vcselPeriodB)
	t.mmTimeoutB = encodeTimeout(calcTimeoutMclks(mmUs, macro))
	t.rangeTimeoutB = encodeTimeout(calcTimeoutMclks(rangeUs, macro))
}

// timeouts is VL53LX_get_timeouts_us.
func (d *Device) timeouts() (phasecalUs, mmUs, rangeUs uint32) {
	macro := calcMacroPeriodUs(d.fastOscFrequency(), d.timingCfg.vcselPeriodA)
	phasecalUs = calcTimeoutUs(uint32(d.generalCfg.phasecalTimeoutMacrop), macro)
	mmUs = calcTimeoutUs(decodeTimeout(d.timingCfg.mmTimeoutA), macro)
	rangeUs = calcTimeoutUs(decodeTimeout(d.timingCfg.rangeTimeoutA), macro)
	return
}
