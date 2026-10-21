// Package vl53l3cx provides a driver for the ST VL53L3CX time-of-flight
// ranging sensor with multi-target detection.
//
// Datasheet: https://www.st.com/resource/en/datasheet/vl53l3cx.pdf
//
// The ranging and histogram processing is ported from ST's VL53LX API as
// published in https://github.com/stm32duino/VL53L3CX (see LICENSE in this
// directory). Crosstalk, offset and SPAD calibration routines are not ported.
package vl53l3cx

import (
	"encoding/binary"
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

// DistanceMode selects the histogram preset. Shorter modes are less
// sensitive to ambient light.
type DistanceMode uint8

const (
	Short DistanceMode = iota + 1
	Medium
	Long
)

// RangeStatus uses the same values as VL53LX_RANGESTATUS in ST's API.
type RangeStatus uint8

const (
	RangeValid                RangeStatus = 0
	SigmaFail                 RangeStatus = 1
	SignalFail                RangeStatus = 2
	OutOfBoundsFail           RangeStatus = 4
	HardwareFail              RangeStatus = 5
	RangeValidNoWrapCheckFail RangeStatus = 6
	WrapTargetFail            RangeStatus = 7
	RangeValidMergedPulse     RangeStatus = 11
	TargetPresentLackOfSignal RangeStatus = 12
	MinRangeFail              RangeStatus = 13
	RangeInvalid              RangeStatus = 14
	NoTarget                  RangeStatus = 255
)

// MaxTargets is the number of targets reported per measurement.
const MaxTargets = 4

// Target is one detected object.
type Target struct {
	Status RangeStatus

	// Distance in millimeters.
	Distance    int16
	DistanceMin int16
	DistanceMax int16

	// Signal and ambient rates in kilo counts per second.
	SignalRate  uint32
	AmbientRate uint32
}

// Measurement is the result of one ranging cycle. When NumTargets is 0,
// Targets[0].Status tells why.
type Measurement struct {
	StreamCount uint8
	NumTargets  uint8
	Targets     [MaxTargets]Target
}

// Config holds the ranging settings. Zero values select ST's defaults.
type Config struct {
	Mode DistanceMode

	// TimingBudget in microseconds, from 1701 to 550000+1700.
	TimingBudget uint32
}

var (
	errNotConnected  = errors.New("vl53l3cx: not connected")
	errTimeout       = errors.New("vl53l3cx: timeout")
	errInvalidParams = errors.New("vl53l3cx: invalid parameter")
	errNotConfigured = errors.New("vl53l3cx: not configured")
)

const (
	bootTimeout  = 500 * time.Millisecond
	rangeTimeout = 2000 * time.Millisecond

	defaultTimingBudget = 33333
	timingGuardUs       = 1700
	maxTimingBudgetUs   = 550000

	// The inter measurement period in ms is scaled by 1.064 in ST's API.
	defaultInterMeasurementMs = 1000 + 1000*64/1000
)

type llState struct {
	cfgDeviceState uint8
	cfgStreamCount uint8
	cfgGphID       uint8
	rdDeviceState  uint8
	rdStreamCount  uint8
	rdGphID        uint8
	rdTimingStatus uint8
}

type additionalOffsetCal struct {
	mmInnerEffectiveSpads uint16
	mmOuterEffectiveSpads uint16
	mmInnerPeakRate       uint16
	mmOuterPeakRate       uint16
}

// zoneObjects keeps the previous results for the phase consistency check.
type zoneObjects struct {
	rdDeviceState uint8
	activeObjects uint8
	phase         [MaxTargets]uint16
}

// zoneHistInfo keeps the previous histogram ambient data.
type zoneHistInfo struct {
	numAmbientBins      uint8
	effectiveSpads      uint16
	totalPeriodsElapsed uint32
	ambientEventsSum    int32
}

// Device is a VL53L3CX sensor.
type Device struct {
	bus     drivers.I2C
	address uint16

	staticNVM         [staticNVMSize]uint8
	customerNVM       [customerNVMSize]uint8
	modeROICentreSpad uint8
	modeROIXYSize     uint8
	goodSpads         [32]uint8
	oscCalibrateVal   uint16
	addOffCal         additionalOffsetCal

	staticCfg  staticConfig
	generalCfg generalConfig
	timingCfg  timingConfig
	dynamicCfg dynamicConfig
	sysCtrl    systemControl
	histCfg    histogramConfig

	validPhaseLow      uint8
	validPhaseHigh     uint8
	interMeasurementMs uint32
	rangeOffsetMM      int16

	state        llState
	savedROI     uint8
	savedROISize uint8

	hist        histogramData
	results     rangeResults
	prevObjects zoneObjects
	prevHist    zoneHistInfo
	merge       histogramMerge
	algo        algoData
	filtered    filteredData
	averaged    histogramData

	buf  [2 + 0x88 - regI2CSlaveDeviceAddress]uint8
	rbuf [2]uint8
}

// New creates a new VL53L3CX connection on the given I2C bus.
func New(bus drivers.I2C) *Device {
	return &Device{
		bus:     bus,
		address: Address,
	}
}

// Connected returns whether a VL53L3CX has been found.
func (d *Device) Connected() bool {
	id, err := d.read16(regIdentificationModelID)
	return err == nil && id == modelID
}

// Configure resets the sensor, reads its calibration data and applies the
// given settings. It is VL53LX_DataInit from ST's API. Call it before
// SetAddress or any other setter, since the reset clears their values.
func (d *Device) Configure(cfg Config) error {
	if !d.Connected() {
		return errNotConnected
	}
	if err := d.softwareReset(); err != nil {
		return err
	}

	d.initState(stateUnknown)
	if err := d.readPartToPartData(); err != nil {
		return err
	}

	d.interMeasurementMs = defaultInterMeasurementMs
	d.setPresetMode(Long, presetPhasecalTimeout(Long), presetMMTimeoutUs, presetRangeTimeoutUs, 1000)
	if err := d.SetTimingBudget(defaultTimingBudget); err != nil {
		return err
	}
	d.timingCfg.intermeasurementPeriod = d.interMeasurementMs * uint32(d.oscCalibrateVal)

	if cfg.Mode != 0 {
		if err := d.SetDistanceMode(cfg.Mode); err != nil {
			return err
		}
	}
	if cfg.TimingBudget != 0 {
		return d.SetTimingBudget(cfg.TimingBudget)
	}
	return nil
}

// SetAddress changes the I2C address of the sensor.
func (d *Device) SetAddress(address uint8) error {
	if err := d.write8(regI2CSlaveDeviceAddress, address&0x7F); err != nil {
		return err
	}
	d.address = uint16(address)
	d.staticNVM[0] = address
	return nil
}

// SetDistanceMode changes the histogram preset. The timing budget is kept.
// It must be called after Configure.
func (d *Device) SetDistanceMode(mode DistanceMode) error {
	if mode != Short && mode != Medium && mode != Long {
		return errInvalidParams
	}
	if d.fastOscFrequency() == 0 {
		return errNotConfigured
	}
	phasecal, mm, rng := d.timeouts()
	d.setPresetMode(mode, presetPhasecalTimeout(mode), presetMMTimeoutUs, presetRangeTimeoutUs, d.interMeasurementMs)
	d.setTimeouts(phasecal, mm, rng)
	return nil
}

// SetTimingBudget sets the measurement time in microseconds.
// It must be called after Configure.
func (d *Device) SetTimingBudget(us uint32) error {
	if us <= timingGuardUs || us-timingGuardUs > maxTimingBudgetUs {
		return errInvalidParams
	}
	if d.fastOscFrequency() == 0 {
		return errNotConfigured
	}
	phasecal, mm, _ := d.timeouts()
	d.setTimeouts(phasecal, mm, (us-timingGuardUs)/6)
	return nil
}

// TimingBudget returns the measurement time in microseconds.
func (d *Device) TimingBudget() uint32 {
	_, _, rng := d.timeouts()
	return 6*rng + timingGuardUs
}

// StartRanging starts continuous back to back ranging.
func (d *Device) StartRanging() error {
	if err := d.loadPatch(); err != nil {
		return err
	}
	return d.initAndStartRange(true)
}

// StopRanging stops ranging.
func (d *Device) StopRanging() error {
	d.sysCtrl.modeStart = d.sysCtrl.modeStart&measurementStopMask | measurementModeAbort
	err := d.writeSystemControl()
	d.sysCtrl.modeStart &= measurementStopMask
	d.initState(stateSWStandby)
	d.prevObjects = zoneObjects{rdDeviceState: stateSWStandby}
	if err != nil {
		return err
	}
	return d.unloadPatch()
}

// DataReady returns whether a new measurement is available.
func (d *Device) DataReady() (bool, error) {
	v, err := d.read8(regGPIOTioHVStatus)
	if err != nil {
		return false, err
	}
	// The interrupt is configured active low in gpio_hv_mux__ctrl.
	return v&0x01 == 0, nil
}

// Read waits for the next measurement, processes it and starts the next
// one.
func (d *Device) Read(m *Measurement) error {
	start := time.Now()
	for {
		ready, err := d.DataReady()
		if err != nil {
			return err
		}
		if ready {
			break
		}
		if time.Since(start) > rangeTimeout {
			return errTimeout
		}
		time.Sleep(time.Millisecond)
	}

	if err := d.deviceResults(); err != nil {
		return err
	}
	d.fillMeasurement(m)
	return d.initAndStartRange(false)
}

func (d *Device) softwareReset() error {
	if err := d.write8(regSoftReset, 0x00); err != nil {
		return err
	}
	time.Sleep(100 * time.Microsecond)
	if err := d.write8(regSoftReset, 0x01); err != nil {
		return err
	}
	time.Sleep(1200 * time.Microsecond)

	start := time.Now()
	for {
		status, err := d.read8(regFirmwareSystemStatus)
		if err == nil && status&0x01 == 0x01 {
			break
		}
		if time.Since(start) > bootTimeout {
			return errTimeout
		}
		time.Sleep(time.Millisecond)
	}
	d.initState(stateSWStandby)
	return nil
}

// readPartToPartData is VL53LX_read_p2p_data.
func (d *Device) readPartToPartData() error {
	if err := d.readBytes(regI2CSlaveDeviceAddress, d.staticNVM[:]); err != nil {
		return err
	}
	if err := d.readBytes(regSpadEnablesRef0, d.customerNVM[:]); err != nil {
		return err
	}
	nvmCopy := d.buf[2 : 2+nvmCopyDataSize]
	if err := d.readBytes(regIdentificationModelID, nvmCopy); err != nil {
		return err
	}
	copy(d.goodSpads[:], nvmCopy[15:47])
	d.modeROICentreSpad = nvmCopy[47]
	d.modeROIXYSize = nvmCopy[48]

	// readNVM writes registers using the first 4 bytes of d.buf.
	raw := d.buf[4 : 4+nvmAdditionalOffsetCalSize]
	if err := d.readNVM(nvmAdditionalOffsetCalIndex>>2, raw); err != nil {
		return err
	}
	d.addOffCal = additionalOffsetCal{
		mmInnerEffectiveSpads: binary.BigEndian.Uint16(raw[0:]),
		mmOuterEffectiveSpads: binary.BigEndian.Uint16(raw[2:]),
		mmInnerPeakRate:       binary.BigEndian.Uint16(raw[4:]),
		mmOuterPeakRate:       binary.BigEndian.Uint16(raw[6:]),
	}
	if d.addOffCal.mmInnerPeakRate == 0 && d.addOffCal.mmOuterPeakRate == 0 {
		d.addOffCal.mmInnerPeakRate = 0x0080
		d.addOffCal.mmOuterPeakRate = 0x0180
		d.addOffCal.mmInnerEffectiveSpads, d.addOffCal.mmOuterEffectiveSpads = calcMMEffectiveSpads(
			d.modeROICentreSpad, d.modeROIXYSize, 0xC7, 0xFF, &d.goodSpads, 0x38)
	}

	osc, err := d.read16(regResultOscCalibrateVal)
	if err != nil {
		return err
	}
	d.oscCalibrateVal = osc

	if d.fastOscFrequency() < 0x1000 {
		binary.BigEndian.PutUint16(d.staticNVM[5:], 0xBCCC)
	}
	return nil
}

func (d *Device) fastOscFrequency() uint16 {
	return binary.BigEndian.Uint16(d.staticNVM[5:])
}

func (d *Device) customerMMInnerOffset() int16 {
	return int16(binary.BigEndian.Uint16(d.customerNVM[19:]))
}

func (d *Device) customerMMOuterOffset() int16 {
	return int16(binary.BigEndian.Uint16(d.customerNVM[21:]))
}

// readNVM reads 4 byte words from the NVM. It is VL53LX_read_nvm_raw_data.
func (d *Device) readNVM(start uint8, data []uint8) error {
	if err := d.setFirmwareEnable(0); err != nil {
		return err
	}
	if err := d.setPowerForce(1); err != nil {
		return err
	}
	time.Sleep(250 * time.Microsecond)
	if err := d.write8(regNVMCtrlPDN, 0x01); err != nil {
		return err
	}
	if err := d.write8(regRangingCoreClkCtrl1, 0x05); err != nil {
		return err
	}
	time.Sleep(50 * time.Microsecond)
	if err := d.write8(regNVMCtrlMode, 0x01); err != nil {
		return err
	}
	if err := d.writeBytes(regNVMCtrlPulseWidthMSB, []uint8{0x00, 0x04}); err != nil {
		return err
	}

	for i := 0; i < len(data)/4; i++ {
		if err := d.write8(regNVMCtrlAddr, start+uint8(i)); err != nil {
			return err
		}
		if err := d.write8(regNVMCtrlReadN, 0x00); err != nil {
			return err
		}
		time.Sleep(5 * time.Microsecond)
		if err := d.write8(regNVMCtrlReadN, 0x01); err != nil {
			return err
		}
		if err := d.readBytes(regNVMCtrlDataOut, data[i*4:i*4+4]); err != nil {
			return err
		}
	}

	if err := d.write8(regNVMCtrlReadN, 0x01); err != nil {
		return err
	}
	if err := d.write8(regNVMCtrlPDN, 0x00); err != nil {
		return err
	}
	if err := d.setPowerForce(0); err != nil {
		return err
	}
	return d.setFirmwareEnable(1)
}

func (d *Device) setFirmwareEnable(v uint8) error {
	d.sysCtrl.firmwareEnable = v
	return d.write8(regFirmwareEnable, v)
}

func (d *Device) setPowerForce(v uint8) error {
	d.sysCtrl.powerForce = v
	return d.write8(regPowerManagementGo1, v)
}

// loadPatch is VL53LX_load_patch.
func (d *Device) loadPatch() error {
	if err := d.write8(regFirmwareEnable, 0x00); err != nil {
		return err
	}
	if err := d.setPowerForce(1); err != nil {
		return err
	}
	if err := d.writeBytes(regPatchOffset0, []uint8{0x29, 0xC9, 0x0E, 0x40, 0x28, 0x00}); err != nil {
		return err
	}
	if err := d.writeBytes(regPatchAddress0, []uint8{0x03, 0x6D, 0x03, 0x6F, 0x07, 0x29}); err != nil {
		return err
	}
	if err := d.writeBytes(regPatchJmpEnables, []uint8{0x00, 0x07}); err != nil {
		return err
	}
	if err := d.writeBytes(regPatchDataEnables, []uint8{0x00, 0x07}); err != nil {
		return err
	}
	if err := d.write8(regPatchCtrl, 0x01); err != nil {
		return err
	}
	return d.write8(regFirmwareEnable, 0x01)
}

// unloadPatch is VL53LX_unload_patch.
func (d *Device) unloadPatch() error {
	if err := d.write8(regFirmwareEnable, 0x00); err != nil {
		return err
	}
	if err := d.setPowerForce(0); err != nil {
		return err
	}
	if err := d.write8(regPatchCtrl, 0x00); err != nil {
		return err
	}
	return d.write8(regFirmwareEnable, 0x01)
}

func (d *Device) writeSystemControl() error {
	var b [systemControlSize]uint8
	d.sysCtrl.encode(b[:])
	return d.writeBytes(regPowerManagementGo1, b[:])
}

// initAndStartRange is VL53LX_init_and_start_range in back to back mode.
// A full write also sends the NVM managed and static config blocks.
func (d *Device) initAndStartRange(full bool) error {
	d.sysCtrl.modeStart = d.sysCtrl.modeStart&measurementStopMask | measurementModeBack2B

	d.savedROI = d.dynamicCfg.userROICentreSpad
	d.savedROISize = d.dynamicCfg.userROIRequestedXY

	index := uint16(regGPHConfigStreamCount)
	if full {
		index = regI2CSlaveDeviceAddress
	}
	size := int(regPowerManagementGo1 + systemControlSize - index)
	b := d.buf[2 : 2+size]
	for i := range b {
		b[i] = 0
	}

	if full {
		s := b[regI2CSlaveDeviceAddress-index:]
		for i := range d.staticNVM {
			s[i] = d.staticNVM[i] & staticNVMMasks[i]
		}
		c := b[regSpadEnablesRef0-index:]
		for i := range d.customerNVM {
			c[i] = d.customerNVM[i] & customerNVMMasks[i]
		}
		d.staticCfg.encode(b[regDSSConfigTargetRate-index:])
	}
	d.generalCfg.encode(b[regGPHConfigStreamCount-index:])
	d.timingCfg.encode(b[regGPHConfigStreamCount+generalConfigSize-index:])

	d.dynamicCfg.gph0 = d.state.cfgGphID | 0x01
	d.dynamicCfg.gph1 = d.state.cfgGphID | 0x01
	d.dynamicCfg.groupedParamHold = d.state.cfgGphID
	d.dynamicCfg.encode(b[regGroupedParameterHold0-index:])
	d.sysCtrl.encode(b[regPowerManagementGo1-index:])

	binary.BigEndian.PutUint16(d.buf[:], index)
	if err := d.bus.Tx(d.address, d.buf[:2+size], nil); err != nil {
		return err
	}

	d.updateRdState()
	d.updateCfgState()
	return nil
}

func (d *Device) initState(state uint8) {
	d.state = llState{
		cfgDeviceState: state,
		cfgGphID:       gphIDMask,
		rdDeviceState:  state,
		rdGphID:        gphIDMask,
	}
}

// updateRdState is VL53LX_update_ll_driver_rd_state for a single zone.
func (d *Device) updateRdState() {
	s := &d.state
	if d.sysCtrl.modeStart&measurementModeMask == 0 {
		s.rdDeviceState = stateSWStandby
		s.rdStreamCount = 0
		s.rdGphID = gphIDMask
		s.rdTimingStatus = 0
		return
	}
	if s.rdStreamCount == 0xFF {
		s.rdStreamCount = 0x80
	} else {
		s.rdStreamCount++
	}
	s.rdGphID ^= gphIDMask

	switch s.rdDeviceState {
	case stateSWStandby:
		if d.dynamicCfg.groupedParamHold&gphIDMask > 0 {
			s.rdDeviceState = stateWaitGPHSync
		} else {
			s.rdDeviceState = stateOutputData
		}
		s.rdStreamCount = 0
		s.rdTimingStatus = 0
	case stateWaitGPHSync:
		s.rdStreamCount = 0
		s.rdDeviceState = stateOutputData
	case stateOutputData:
		s.rdTimingStatus ^= 0x01
	default:
		s.rdDeviceState = stateSWStandby
		s.rdStreamCount = 0
		s.rdGphID = gphIDMask
		s.rdTimingStatus = 0
	}
}

// updateCfgState is VL53LX_update_ll_driver_cfg_state for a single zone.
func (d *Device) updateCfgState() {
	s := &d.state
	if d.sysCtrl.modeStart&measurementModeMask == 0 {
		s.cfgDeviceState = stateSWStandby
		s.cfgStreamCount = 0
		s.cfgGphID = gphIDMask
		return
	}
	if s.cfgStreamCount == 0xFF {
		s.cfgStreamCount = 0x80
	} else {
		s.cfgStreamCount++
	}
	s.cfgGphID ^= gphIDMask

	switch s.cfgDeviceState {
	case stateSWStandby:
		s.cfgStreamCount = 1
		s.cfgDeviceState = stateDSSAuto
	case stateDSSAuto:
	default:
		s.cfgDeviceState = stateSWStandby
		s.cfgStreamCount = 0
		s.cfgGphID = gphIDMask
	}
}

func (d *Device) read8(reg uint16) (uint8, error) {
	err := d.readBytes(reg, d.rbuf[:1])
	return d.rbuf[0], err
}

func (d *Device) read16(reg uint16) (uint16, error) {
	err := d.readBytes(reg, d.rbuf[:2])
	return binary.BigEndian.Uint16(d.rbuf[:]), err
}

// readBytes uses the first two bytes of d.buf, so data may be the rest of it.
func (d *Device) readBytes(reg uint16, data []uint8) error {
	binary.BigEndian.PutUint16(d.buf[:], reg)
	return d.bus.Tx(d.address, d.buf[:2], data)
}

func (d *Device) write8(reg uint16, v uint8) error {
	binary.BigEndian.PutUint16(d.buf[:], reg)
	d.buf[2] = v
	return d.bus.Tx(d.address, d.buf[:3], nil)
}

func (d *Device) writeBytes(reg uint16, data []uint8) error {
	binary.BigEndian.PutUint16(d.buf[:], reg)
	n := copy(d.buf[2:], data)
	return d.bus.Tx(d.address, d.buf[:2+n], nil)
}
