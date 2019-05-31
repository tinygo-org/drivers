package vl53l1x

// The I2C address which this device listens to.
const Address = 0x29 //0x52

// Registers
const (
	CHIP_ID                                                 = 0xEACC
	SOFT_RESET                                              = 0x0000
	OSC_MEASURED_FAST_OSC_FREQUENCY                         = 0x0006
	VHV_CONFIG_TIMEOUT_MACROP_LOOP_BOUND                    = 0x0008
	VHV_CONFIG_INIT                                         = 0x000B
	ALGO_PART_TO_PART_RANGE_OFFSET_MM                       = 0x001E
	MM_CONFIG_OUTER_OFFSET_MM                               = 0x0022
	DSS_CONFIG_TARGET_TOTAL_RATE_MCPS                       = 0x0024
	PAD_I2C_HV_EXTSUP_CONFIG                                = 0x002E
	GPIO_TIO_HV_STATUS                                      = 0x0031
	SIGMA_ESTIMATOR_EFFECTIVE_PULSE_WIDTH_NS                = 0x0036
	SIGMA_ESTIMATOR_EFFECTIVE_AMBIENT_WIDTH_NS              = 0x0037
	ALGO_CROSSTALK_COMPENSATION_VALID_HEIGHT_MM             = 0x0039
	ALGO_RANGE_MIN_CLIP                                     = 0x003F
	ALGO_CONSISTENCY_CHECK_TOLERANCE                        = 0x0040
	CAL_CONFIG_VCSEL_START                                  = 0x0047
	PHASECAL_CONFIG_TIMEOUT_MACROP                          = 0x004B
	PHASECAL_CONFIG_OVERRIDE                                = 0x004D
	DSS_CONFIG_ROI_MODE_CONTROL                             = 0x004F
	SYSTEM_THRESH_RATE_HIGH                                 = 0x0050
	SYSTEM_THRESH_RATE_LOW                                  = 0x0052
	DSS_CONFIG_MANUAL_EFFECTIVE_SPADS_SELECT                = 0x0054
	DSS_CONFIG_APERTURE_ATTENUATION                         = 0x0057
	MM_CONFIG_TIMEOUT_MACROP_A                              = 0x005A
	MM_CONFIG_TIMEOUT_MACROP_B                              = 0x005C
	RANGE_CONFIG_TIMEOUT_MACROP_A                           = 0x005E
	RANGE_CONFIG_VCSEL_PERIOD_A                             = 0x0060
	RANGE_CONFIG_TIMEOUT_MACROP_B                           = 0x0061
	RANGE_CONFIG_VCSEL_PERIOD_B                             = 0x0063
	RANGE_CONFIG_SIGMA_THRESH                               = 0x0064
	RANGE_CONFIG_MIN_COUNT_RATE_RTN_LIMIT_MCPS              = 0x0066
	RANGE_CONFIG_VALID_PHASE_HIGH                           = 0x0069
	SYSTEM_INTERMEASUREMENT_PERIOD                          = 0x006C
	SYSTEM_GROUPED_PARAMETER_HOLD_0                         = 0x0071
	SYSTEM_SEED_CONFIG                                      = 0x0077
	SD_CONFIG_WOI_SD0                                       = 0x0078
	SD_CONFIG_WOI_SD1                                       = 0x0079
	SD_CONFIG_INITIAL_PHASE_SD0                             = 0x007A
	SD_CONFIG_INITIAL_PHASE_SD1                             = 0x007B
	SYSTEM_GROUPED_PARAMETER_HOLD_1                         = 0x007C
	SD_CONFIG_QUANTIFIER                                    = 0x007E
	SYSTEM_SEQUENCE_CONFIG                                  = 0x0081
	SYSTEM_GROUPED_PARAMETER_HOLD                           = 0x0082
	SYSTEM_INTERRUPT_CLEAR                                  = 0x0086
	SYSTEM_MODE_START                                       = 0x0087
	RESULT_RANGE_STATUS                                     = 0x0089
	PHASECAL_RESULT_VCSEL_START                             = 0x00D8
	RESULT_OSC_CALIBRATE_VAL                                = 0x00DE
	FIRMWARE_SYSTEM_STATUS                                  = 0x00E5
	WHO_AM_I                                                = 0x010F
	SHADOW_RESULT_FINAL_CROSSTALK_CORRECTED_RANGE_MM_SD0_HI = 0x0FBE

	TIMING_GUARD = 4528
	TARGETRATE   = 0x0A00
)

const (
	SHORT DistanceMode = iota
	MEDIUM
	LONG
)

const (
	RangeValid RangeStatus = iota
	SigmaFail
	SignalFail
	RangeValidMinRangeClipped
	OutOfBoundsFail
	HardwareFail
	RangeValidNoWrapCheckFail
	WrapTargetFail
	ProcessingFail
	XtalkSignalFail
	SynchronizationInt
	MergedPulse
	TargetPresentLackOfSignal
	MinRangeFail
	RangeInvalid

	None RangeStatus = 255
)
