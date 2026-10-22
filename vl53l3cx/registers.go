package vl53l3cx

// Address is the default I2C address.
const Address = 0x29

const modelID = 0xEAAA

// Registers
const (
	regSoftReset              = 0x0000
	regI2CSlaveDeviceAddress  = 0x0001
	regSpadEnablesRef0        = 0x000D
	regDSSConfigTargetRate    = 0x0024
	regPadI2CHVExtsupConfig   = 0x002E
	regGPIOTioHVStatus        = 0x0031
	regGPHConfigStreamCount   = 0x0044
	regGroupedParameterHold0  = 0x0071
	regPowerManagementGo1     = 0x0083
	regFirmwareEnable         = 0x0085
	regResultInterruptStatus  = 0x0088
	regPhasecalReferencePhase = 0x00D6
	regPhasecalVcselStart     = 0x00D8
	regHistogramBin23_0MSB    = 0x00D9
	regHistogramBin23_0LSB    = 0x00DA
	regHistogramBin23_0       = 0x00D5
	regHistogramBin0_2        = 0x008E
	regResultOscCalibrateVal  = 0x00DE
	regFirmwareSystemStatus   = 0x00E5
	regIdentificationModelID  = 0x010F
	regPatchCtrl              = 0x0470
	regPatchJmpEnables        = 0x0472
	regPatchDataEnables       = 0x0474
	regPatchOffset0           = 0x0476
	regPatchAddress0          = 0x0496
	regRangingCoreClkCtrl1    = 0x0683
	regNVMCtrlMode            = 0x0780
	regNVMCtrlPDN             = 0x0781
	regNVMCtrlReadN           = 0x0783
	regNVMCtrlPulseWidthMSB   = 0x0784
	regNVMCtrlDataOut         = 0x0790
	regNVMCtrlAddr            = 0x0794
)

// Register block sizes
const (
	staticNVMSize     = 11
	customerNVMSize   = 23
	staticConfigSize  = 32
	generalConfigSize = 22
	timingConfigSize  = 23
	dynamicConfigSize = 18
	systemControlSize = 5
	nvmCopyDataSize   = 49

	histogramBinDataSize = regHistogramBin23_0LSB - regResultInterruptStatus + 1
)

// NVM locations of the additional offset calibration data
const (
	nvmAdditionalOffsetCalIndex = 0x0194
	nvmAdditionalOffsetCalSize  = 8
)

// system__mode_start values
const (
	schedulerModeHistogram = 0x02
	readoutModeDualSD      = 0x01 << 2
	measurementModeStop    = 0x00
	measurementModeBack2B  = 0x20
	measurementModeAbort   = 0x80
	measurementModeMask    = 0xF0
	measurementStopMask    = 0x0F
)

// system__sequence_config bits
const (
	sequenceVHV      = 0x01
	sequencePhasecal = 0x02
	sequenceDSS1     = 0x08
	sequenceDSS2     = 0x10
	sequenceMM1      = 0x20
	sequenceRange    = 0x80
)

// Internal device states from ST's VL53LX API
const (
	stateSWStandby   = 3
	stateDSSAuto     = 4
	stateWaitGPHSync = 6
	stateOutputData  = 8
	stateUnknown     = 98
)

const gphIDMask = 0x02

// Device errors from ST's VL53LX API
const (
	deviceErrorNoUpdate                 = 0
	deviceErrorVcselContinuityFail      = 1
	deviceErrorVcselWatchdogFail        = 2
	deviceErrorNoVHVValueFound          = 3
	deviceErrorRangePhaseCheck          = 5
	deviceErrorSigmaThresholdCheck      = 6
	deviceErrorPhaseConsistency         = 7
	deviceErrorRangeComplete            = 9
	deviceErrorUserROIClip              = 13
	deviceErrorMultClipFail             = 17
	deviceErrorRangeCompleteNoWrapCheck = 19
	deviceErrorEventConsistency         = 20
	deviceErrorRangeCompleteMergedPulse = 22
	deviceErrorPrevRangeNoTargets       = 23
)
