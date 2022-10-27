// Package vl53l1x provides a driver for the VL53L1X time-of-flight
// distance sensor
//
// Datasheet:
// https://www.st.com/resource/en/datasheet/vl53l1x.pdf
// This driver was based on the library https://github.com/pololu/vl53l1x-arduino
// and ST's VL53L1X API (STSW-IMG007)
// https://www.st.com/content/st_com/en/products/embedded-software/proximity-sensors-software/stsw-img007.html
package vl53l1x // import "tinygo.org/x/drivers/vl53l1x"

import (
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

type DistanceMode uint8
type RangeStatus uint8

type rangingData struct {
	mm                 uint16
	status             RangeStatus
	signalRateMCPS     int32 //MCPS : Mega Count Per Second
	ambientRateMCPS    int32
	effectiveSPADCount uint16
}

type resultBuffer struct {
	status                     uint8
	streamCount                uint8
	effectiveSPADCount         uint16
	ambientRateMCPSSD0         uint16
	mmCrosstalkSD0             uint16
	signalRateCrosstalkMCPSSD0 uint16
}

// Device wraps an I2C connection to a VL53L1X device.
type Device struct {
	bus                drivers.I2C
	Address            uint16
	mode               DistanceMode
	timeout            uint32
	fastOscillatorFreq uint16
	oscillatorOffset   uint16
	calibrated         bool
	VHVInit            uint8
	VHVTimeout         uint8
	rangingData        rangingData
	results            resultBuffer
}

// New creates a new VL53L1X connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
		mode:    LONG,
		timeout: 500,
	}
}

// Connected returns whether a VL53L1X has been found.
// It does a "who am I" request and checks the response.
func (d *Device) Connected() bool {
	return d.readReg16Bit(WHO_AM_I) == CHIP_ID
}

// Configure sets up the device for communication
func (d *Device) Configure(use2v8Mode bool) bool {
	if !d.Connected() {
		return false
	}
	d.writeReg(SOFT_RESET, 0x00)
	time.Sleep(100 * time.Microsecond)
	d.writeReg(SOFT_RESET, 0x01)
	time.Sleep(1 * time.Millisecond)

	start := time.Now()
	for (d.readReg(FIRMWARE_SYSTEM_STATUS) & 0x01) == 0 {
		elapsed := time.Since(start)
		if d.timeout > 0 && uint32(elapsed.Seconds()*1000) > d.timeout {
			return false
		}
	}

	if use2v8Mode {
		d.writeReg(PAD_I2C_HV_EXTSUP_CONFIG, d.readReg(PAD_I2C_HV_EXTSUP_CONFIG)|0x01)
	}

	d.fastOscillatorFreq = d.readReg16Bit(OSC_MEASURED_FAST_OSC_FREQUENCY)
	d.oscillatorOffset = d.readReg16Bit(RESULT_OSC_CALIBRATE_VAL)

	// static config
	d.writeReg16Bit(DSS_CONFIG_TARGET_TOTAL_RATE_MCPS, TARGETRATE)
	d.writeReg(GPIO_TIO_HV_STATUS, 0x02)
	d.writeReg(SIGMA_ESTIMATOR_EFFECTIVE_PULSE_WIDTH_NS, 8)
	d.writeReg(SIGMA_ESTIMATOR_EFFECTIVE_AMBIENT_WIDTH_NS, 16)
	d.writeReg(ALGO_CROSSTALK_COMPENSATION_VALID_HEIGHT_MM, 0xFF)
	d.writeReg(ALGO_RANGE_MIN_CLIP, 0)
	d.writeReg(ALGO_CONSISTENCY_CHECK_TOLERANCE, 2)

	// general config
	d.writeReg16Bit(SYSTEM_THRESH_RATE_HIGH, 0x0000)
	d.writeReg16Bit(SYSTEM_THRESH_RATE_LOW, 0x0000)
	d.writeReg(DSS_CONFIG_APERTURE_ATTENUATION, 0x38)

	// timing config
	d.writeReg16Bit(RANGE_CONFIG_SIGMA_THRESH, 360)
	d.writeReg16Bit(RANGE_CONFIG_MIN_COUNT_RATE_RTN_LIMIT_MCPS, 192)

	// dynamic config
	d.writeReg(SYSTEM_GROUPED_PARAMETER_HOLD_0, 0x01)
	d.writeReg(SYSTEM_GROUPED_PARAMETER_HOLD_1, 0x01)
	d.writeReg(SD_CONFIG_QUANTIFIER, 2)

	d.writeReg(SYSTEM_GROUPED_PARAMETER_HOLD, 0x00)
	d.writeReg(SYSTEM_SEED_CONFIG, 1)

	// Low power auto mode
	d.writeReg(SYSTEM_SEQUENCE_CONFIG, 0x8B) // VHV, PHASECAL, DSS1, RANGE
	d.writeReg16Bit(DSS_CONFIG_MANUAL_EFFECTIVE_SPADS_SELECT, 200<<8)
	d.writeReg(DSS_CONFIG_ROI_MODE_CONTROL, 2) // REQUESTED_EFFFECTIVE_SPADS

	d.SetDistanceMode(d.mode)
	d.SetMeasurementTimingBudget(50000)

	d.writeReg16Bit(ALGO_PART_TO_PART_RANGE_OFFSET_MM, d.readReg16Bit(MM_CONFIG_OUTER_OFFSET_MM)*4)

	return true
}

// SetAddress sets the I2C address which this device listens to.
func (d *Device) SetAddress(address uint8) {
	d.writeReg(I2C_SLAVE_DEVICE_ADDRESS, address)
	d.Address = uint16(address)
}

// GetAddress returns the I2C address which this device listens to.
func (d *Device) GetAddress() uint8 {
	return uint8(d.Address)
}

// SetTimeout configures the timeout
func (d *Device) SetTimeout(timeout uint32) {
	d.timeout = timeout
}

// SetDistanceMode sets the mode for calculating the distance.
// Distance mode vs. max. distance
// SHORT: 136cm (dark) - 135cm (strong ambient light)
// MEDIUM: 290cm (dark) - 76cm (strong ambient light)
// LONG: 360cm (dark) - 73cm (strong ambient light)
// It returns false if an invalid mode is provided
func (d *Device) SetDistanceMode(mode DistanceMode) bool {
	budgetMicroseconds := d.GetMeasurementTimingBudget()
	switch mode {
	case SHORT:
		// timing config
		d.writeReg(RANGE_CONFIG_VCSEL_PERIOD_A, 0x07)
		d.writeReg(RANGE_CONFIG_VCSEL_PERIOD_B, 0x05)
		d.writeReg(RANGE_CONFIG_VALID_PHASE_HIGH, 0x38)

		// dynamic config
		d.writeReg(SD_CONFIG_WOI_SD0, 0x07)
		d.writeReg(SD_CONFIG_WOI_SD1, 0x05)
		d.writeReg(SD_CONFIG_INITIAL_PHASE_SD0, 6)
		d.writeReg(SD_CONFIG_INITIAL_PHASE_SD1, 6)
	case MEDIUM:
		// timing config
		d.writeReg(RANGE_CONFIG_VCSEL_PERIOD_A, 0x0B)
		d.writeReg(RANGE_CONFIG_VCSEL_PERIOD_B, 0x09)
		d.writeReg(RANGE_CONFIG_VALID_PHASE_HIGH, 0x78)

		// dynamic config
		d.writeReg(SD_CONFIG_WOI_SD0, 0x0B)
		d.writeReg(SD_CONFIG_WOI_SD1, 0x09)
		d.writeReg(SD_CONFIG_INITIAL_PHASE_SD0, 10)
		d.writeReg(SD_CONFIG_INITIAL_PHASE_SD1, 10)
	case LONG:
		// timing config
		d.writeReg(RANGE_CONFIG_VCSEL_PERIOD_A, 0x0F)
		d.writeReg(RANGE_CONFIG_VCSEL_PERIOD_B, 0x0D)
		d.writeReg(RANGE_CONFIG_VALID_PHASE_HIGH, 0xB8)

		// dynamic config
		d.writeReg(SD_CONFIG_WOI_SD0, 0x0F)
		d.writeReg(SD_CONFIG_WOI_SD1, 0x0D)
		d.writeReg(SD_CONFIG_INITIAL_PHASE_SD0, 14)
		d.writeReg(SD_CONFIG_INITIAL_PHASE_SD1, 14)
	default:
		return false
	}

	d.SetMeasurementTimingBudget(budgetMicroseconds)
	d.mode = mode
	return true
}

// GetMeasurementTimingBudget returns the timing budget in microseconds
func (d *Device) GetMeasurementTimingBudget() uint32 {
	macroPeriod := d.calculateMacroPeriod(uint32(d.readReg(RANGE_CONFIG_VCSEL_PERIOD_A)))
	rangeConfigTimeout := timeoutMclksToMicroseconds(decodeTimeout(d.readReg16Bit(RANGE_CONFIG_TIMEOUT_MACROP_A)), macroPeriod)
	return 2 * uint32(rangeConfigTimeout) * TIMING_GUARD
}

// SetMeasurementTimingBudget configures the timing budget in microseconds
// It returns false if an invalid timing budget is provided
func (d *Device) SetMeasurementTimingBudget(budgetMicroseconds uint32) bool {
	if budgetMicroseconds <= TIMING_GUARD {
		return false
	}
	budgetMicroseconds -= TIMING_GUARD
	if budgetMicroseconds > 1100000 {
		return false
	}
	rangeConfigTimeout := budgetMicroseconds / 2
	// Update Macro Period for Range A VCSEL Period
	macroPeriod := d.calculateMacroPeriod(uint32(d.readReg(RANGE_CONFIG_VCSEL_PERIOD_A)))

	// Update Phase timeout - uses Timing A
	phasecalTimeoutMclks := timeoutMicrosecondsToMclks(1000, macroPeriod)
	if phasecalTimeoutMclks > 0xFF {
		phasecalTimeoutMclks = 0xFF
	}
	d.writeReg(PHASECAL_CONFIG_TIMEOUT_MACROP, uint8(phasecalTimeoutMclks))

	// Update MM Timing A timeout
	d.writeReg16Bit(MM_CONFIG_TIMEOUT_MACROP_A, encodeTimeout(timeoutMicrosecondsToMclks(1, macroPeriod)))
	// Update Range Timing A timeout
	d.writeReg16Bit(RANGE_CONFIG_TIMEOUT_MACROP_A, encodeTimeout(timeoutMicrosecondsToMclks(rangeConfigTimeout, macroPeriod)))

	macroPeriod = d.calculateMacroPeriod(uint32(d.readReg(RANGE_CONFIG_VCSEL_PERIOD_B)))
	// Update MM Timing B timeout
	d.writeReg16Bit(MM_CONFIG_TIMEOUT_MACROP_B, encodeTimeout(timeoutMicrosecondsToMclks(1, macroPeriod)))
	// Update Range Timing B timeout
	d.writeReg16Bit(RANGE_CONFIG_TIMEOUT_MACROP_B, encodeTimeout(timeoutMicrosecondsToMclks(rangeConfigTimeout, macroPeriod)))

	return true
}

// Read stores in the buffer the values of the sensor and returns
// the current distance in mm
func (d *Device) Read(blocking bool) uint16 {
	if blocking {
		start := time.Now()

		for !d.dataReady() {
			elapsed := time.Since(start)
			if d.timeout > 0 && uint32(elapsed.Seconds()*1000) > d.timeout {
				d.rangingData.status = None
				d.rangingData.mm = 0
				d.rangingData.signalRateMCPS = 0
				d.rangingData.ambientRateMCPS = 0
				return d.rangingData.mm
			}
		}
	}
	d.readResults()

	if !d.calibrated {
		d.setupManualCalibration()
		d.calibrated = true
	}

	d.updateDSS()
	d.getRangingData()
	d.writeReg(SYSTEM_INTERRUPT_CLEAR, 0x01) //sys_interrupt_clear_range

	return d.rangingData.mm
}

// updateDSS updates the DSS
func (d *Device) updateDSS() {
	spadCount := d.results.effectiveSPADCount
	if spadCount != 0 {
		totalRatePerSpad := uint32(d.results.signalRateCrosstalkMCPSSD0) + uint32(d.results.ambientRateMCPSSD0)
		if totalRatePerSpad > 0xFFFF {
			totalRatePerSpad = 0xFFFF
		}
		totalRatePerSpad <<= 16
		totalRatePerSpad /= uint32(spadCount)
		if totalRatePerSpad != 0 {
			requireSpads := (uint32(TARGETRATE) << 16) / totalRatePerSpad
			if requireSpads > 0xFFFF {
				requireSpads = 0xFFFF
			}
			d.writeReg16Bit(DSS_CONFIG_MANUAL_EFFECTIVE_SPADS_SELECT, uint16(requireSpads))
			return
		}
	}
	d.writeReg16Bit(DSS_CONFIG_MANUAL_EFFECTIVE_SPADS_SELECT, 0x8000)
}

// readResults read the register and stores the data in the results buffer
func (d *Device) readResults() {
	data := make([]byte, 17)
	msb := byte((RESULT_RANGE_STATUS >> 8) & 0xFF)
	lsb := byte(RESULT_RANGE_STATUS & 0xFF)
	d.bus.Tx(d.Address, []byte{msb, lsb}, data)
	d.results.status = data[0]
	// data[1] report_status : not used
	d.results.streamCount = data[2]
	d.results.effectiveSPADCount = readUint(data[3], data[4])
	// data[5] , data[6] peak signal count rate mcps sd0 : not used
	d.results.ambientRateMCPSSD0 = readUint(data[7], data[8])
	// data[9] , data[10] sigma_sd0 : not used
	// data[11] , data[12] phase_sd0 : not used
	d.results.mmCrosstalkSD0 = readUint(data[13], data[14])
	d.results.signalRateCrosstalkMCPSSD0 = readUint(data[15], data[16])
}

// dataReady returns true when the data is ready to be read
func (d *Device) dataReady() bool {
	return (d.readReg(GPIO_TIO_HV_STATUS) & 0x01) == 0
}

// Distance returns the distance in mm
func (d *Device) Distance() int32 {
	return int32(d.rangingData.mm)
}

// Status returns the status of the sensor
func (d *Device) Status() RangeStatus {
	return d.rangingData.status
}

// SignalRate returns the peak signal rate in count per second (cps)
func (d *Device) SignalRate() int32 {
	return d.rangingData.signalRateMCPS
}

// AmbientRate returns the ambient rate in count per second (cps)
func (d *Device) AmbientRate() int32 {
	return d.rangingData.ambientRateMCPS
}

// EffectiveSPADCount returns the effective number of SPADs
func (d *Device) EffectiveSPADCount() uint16 {
	return d.rangingData.effectiveSPADCount
}

// getRangingData stores in the buffer the ranging data
func (d *Device) getRangingData() {
	d.rangingData.mm = uint16((uint32(d.results.mmCrosstalkSD0)*2011 + 0x0400) / 0x0800)
	switch d.results.status {
	case 1, // VCSELCONTINUITYTESTFAILURE
		2,  // VCSELWATCHDOGTESTFAILURE
		3,  // NOVHVVALUEFOUND
		17: // MULTCLIPFAIL
		d.rangingData.status = HardwareFail

	case 13: // USERROICLIP
		d.rangingData.status = MinRangeFail

	case 18: // GPHSTREAMCOUNT0READY
		d.rangingData.status = SynchronizationInt

	case 5: // RANGEPHASECHECK
		d.rangingData.status = OutOfBoundsFail

	case 4: // MSRCNOTARGET
		d.rangingData.status = SignalFail

	case 6: // SIGMATHRESHOLDCHECK
		d.rangingData.status = SignalFail

	case 7: // PHASECONSISTENCY
		d.rangingData.status = WrapTargetFail

	case 12: // RANGEIGNORETHRESHOLD
		d.rangingData.status = XtalkSignalFail

	case 8: // MINCLIP
		d.rangingData.status = RangeValidMinRangeClipped

	case 9: // RANGECOMPLETE
		if d.results.streamCount == 0 {
			d.rangingData.status = RangeValidNoWrapCheckFail
		} else {
			d.rangingData.status = RangeValid
		}

	default:
		d.rangingData.status = None
	}

	d.rangingData.signalRateMCPS = 1000000 * int32(d.results.signalRateCrosstalkMCPSSD0) / (1 << 7)
	d.rangingData.ambientRateMCPS = 1000000 * int32(d.results.ambientRateMCPSSD0) / (1 << 7)
	d.rangingData.effectiveSPADCount = d.results.effectiveSPADCount
}

// setupManualCalibration configures the manual calibration
func (d *Device) setupManualCalibration() {
	// save original VHV configs
	d.VHVInit = d.readReg(VHV_CONFIG_INIT)
	d.VHVTimeout = d.readReg(VHV_CONFIG_TIMEOUT_MACROP_LOOP_BOUND)

	// disable VHV init
	d.writeReg(VHV_CONFIG_INIT, d.VHVInit&0x7F)

	// set loop bound to tuning param
	d.writeReg(VHV_CONFIG_TIMEOUT_MACROP_LOOP_BOUND, (d.VHVTimeout&0x03)+(3<<2))

	// override phasecal
	d.writeReg(PHASECAL_CONFIG_OVERRIDE, 0x01)
	d.writeReg(CAL_CONFIG_VCSEL_START, d.readReg(PHASECAL_RESULT_VCSEL_START))
}

// StartContinuous starts the continuous sensing mode
func (d *Device) StartContinuous(periodMs uint32) {
	d.writeReg32Bit(SYSTEM_INTERMEASUREMENT_PERIOD, periodMs*uint32(d.oscillatorOffset))
	d.writeReg(SYSTEM_INTERRUPT_CLEAR, 0x01) // sys_interrupt_clear_range
	d.writeReg(SYSTEM_MODE_START, 0x40)      // mode_range_timed
}

// StopContinuous stops the continuous sensing mode
func (d *Device) StopContinuous() {
	d.writeReg(SYSTEM_MODE_START, 0x80) // mode_range_abort

	d.calibrated = false

	// restore vhv configs
	if d.VHVInit != 0 {
		d.writeReg(VHV_CONFIG_INIT, d.VHVInit)
	}
	if d.VHVTimeout != 0 {
		d.writeReg(VHV_CONFIG_TIMEOUT_MACROP_LOOP_BOUND, d.VHVTimeout)
	}

	// remove phasecal override
	d.writeReg(PHASECAL_CONFIG_OVERRIDE, 0x00)
}

// SetROI sets the 'region of interest' for x and y coordinates. Valid ranges are from 4/4 to 16/16.
func (d *Device) SetROI(x, y uint8) error {
	if !validROIRange(x, y) {
		return errors.New("ROI value out of range")
	}

	if x > 10 || y > 10 {
		d.writeReg(ROI_CONFIG_USER_ROI_CENTRE_SPAD, 199)
	}

	d.writeReg(ROI_CONFIG_USER_ROI_REQUESTED_GLOBAL_XY_SIZE, (y-1)<<4|(x-1))
	return nil
}

// GetROI returns the currently configured 'region of interest' for x and y coordinates.
func (d *Device) GetROI() (x, y uint8, err error) {
	reg := d.readReg(ROI_CONFIG_USER_ROI_REQUESTED_GLOBAL_XY_SIZE)

	x = (reg & 0x0f) + 1
	y = ((reg & 0xf0) >> 4) + 1

	if !validROIRange(x, y) {
		err = errors.New("ROI value out of range")
	}

	return
}

func validROIRange(x, y uint8) bool {
	return x >= 4 && x <= 16 && y >= 4 && y <= 16
}

// writeReg sends a single byte to the specified register address
func (d *Device) writeReg(reg uint16, value uint8) {
	msb := byte((reg >> 8) & 0xFF)
	lsb := byte(reg & 0xFF)
	d.bus.Tx(d.Address, []byte{msb, lsb, value}, nil)
}

// writeReg16Bit sends two bytes to the specified register address
func (d *Device) writeReg16Bit(reg uint16, value uint16) {
	data := make([]byte, 4)
	data[0] = byte((reg >> 8) & 0xFF)
	data[1] = byte(reg & 0xFF)
	data[2] = byte((value >> 8) & 0xFF)
	data[3] = byte(value & 0xFF)
	d.bus.Tx(d.Address, data, nil)
}

// writeReg32Bit sends four bytes to the specified register address
func (d *Device) writeReg32Bit(reg uint16, value uint32) {
	data := make([]byte, 6)
	data[0] = byte((reg >> 8) & 0xFF)
	data[1] = byte(reg & 0xFF)
	data[2] = byte((value >> 24) & 0xFF)
	data[3] = byte((value >> 16) & 0xFF)
	data[4] = byte((value >> 8) & 0xFF)
	data[5] = byte(value & 0xFF)
	d.bus.Tx(d.Address, data, nil)
}

// readReg reads a single byte from the specified address
func (d *Device) readReg(reg uint16) uint8 {
	data := []byte{0}
	msb := byte((reg >> 8) & 0xFF)
	lsb := byte(reg & 0xFF)
	d.bus.Tx(d.Address, []byte{msb, lsb}, data)
	return data[0]
}

// readReg16Bit reads two bytes from the specified address
// and returns it as a uint16
func (d *Device) readReg16Bit(reg uint16) uint16 {
	data := []byte{0, 0}
	msb := byte((reg >> 8) & 0xFF)
	lsb := byte(reg & 0xFF)
	d.bus.Tx(d.Address, []byte{msb, lsb}, data)
	return readUint(data[0], data[1])
}

// readReg32Bit reads four bytes from the specified address
// and returns it as a uint32
func (d *Device) readReg32Bit(reg uint16) uint32 {
	data := make([]byte, 4)
	msb := byte((reg >> 8) & 0xFF)
	lsb := byte(reg & 0xFF)
	d.bus.Tx(d.Address, []byte{msb, lsb}, data)
	return readUint32(data)
}

// readUint converts two bytes to uint16
func readUint(msb byte, lsb byte) uint16 {
	return (uint16(msb) << 8) | uint16(lsb)
}

// readUint converts four bytes to uint32
func readUint32(data []byte) uint32 {
	if len(data) != 4 {
		return 0
	}
	var value uint32
	value = uint32(data[0]) << 24
	value |= uint32(data[1]) << 16
	value |= uint32(data[2]) << 8
	value |= uint32(data[3])
	return value
}

// encodeTimeout encodes the timeout in the correct format: (LSByte * 2^MSByte) + 1
func encodeTimeout(timeoutMclks uint32) uint16 {
	if timeoutMclks == 0 {
		return 0
	}
	msb := 0
	lsb := timeoutMclks - 1
	for (lsb & 0xFFFFFF00) > 0 {
		lsb >>= 1
		msb++
	}
	return uint16(msb<<8) | uint16(lsb&0xFF)
}

// decodeTimeout decodes the timeout from the format: (LSByte * 2^MSByte) + 1
func decodeTimeout(regVal uint16) uint32 {
	return (uint32(regVal&0xFF) << (regVal >> 8)) + 1
}

// timeoutMclksToMicroseconds transform from mclks to microseconds
func timeoutMclksToMicroseconds(timeoutMclks uint32, macroPeriodMicroseconds uint32) uint32 {
	return uint32((uint64(timeoutMclks)*uint64(macroPeriodMicroseconds) + 0x800) >> 12)
}

// timeoutMicrosecondsToMclks transform from microseconds to mclks
func timeoutMicrosecondsToMclks(timeoutMicroseconds uint32, macroPeriodMicroseconds uint32) uint32 {
	return ((timeoutMicroseconds << 12) + (macroPeriodMicroseconds >> 1)) / macroPeriodMicroseconds
}

// calculateMacroPerios calculates the macro period in microsendos from the vcsel period
func (d *Device) calculateMacroPeriod(vcselPeriod uint32) uint32 {
	pplPeriodMicroseconds := (uint32(1) << 30) / uint32(d.fastOscillatorFreq)
	vcselPeriodPclks := (vcselPeriod + 1) << 1
	macroPeriodMicroseconds := 2304 * pplPeriodMicroseconds
	macroPeriodMicroseconds >>= 6
	macroPeriodMicroseconds *= vcselPeriodPclks
	macroPeriodMicroseconds >>= 6
	return macroPeriodMicroseconds
}
