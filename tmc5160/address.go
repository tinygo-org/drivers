package tmc5160

// Driver Register addresses
const (
	GCONF         uint8 = 0x00 // Global configuration flags
	GSTAT         uint8 = 0x01 // Global status flags
	IFCNT               = 0x02 // UART transmission counter
	SLAVECONF           = 0x03 // UART slave configuration
	IOIN                = 0x04 // Read input / write output pins
	X_COMPARE           = 0x05 // Position comparison register
	OTP_PROG            = 0x06 // OTP programming register
	OTP_READ            = 0x07 // OTP read register
	FACTORY_CONF        = 0x08 // Factory configuration (clock trim)
	SHORT_CONF          = 0x09 // Short detector configuration
	DRV_CONF            = 0x0A // Driver configuration
	GLOBAL_SCALER       = 0x0B // Global scaling of motor current
	OFFSET_READ         = 0x0C // Offset calibration results

	/* Velocity dependent driver feature control registers */
	IHOLD_IRUN = 0x10 // Driver current control
	TPOWERDOWN = 0x11 // Delay before power down
	TSTEP      = 0x12 // Actual time between microsteps
	TPWMTHRS   = 0x13 // Upper velocity for stealthChop voltage PWM mode
	TCOOLTHRS  = 0x14 // Lower threshold velocity for switching on smart energy coolStep and stallGuard feature
	THIGH      = 0x15 // Velocity threshold for switching into a different chopper mode and fullstepping

	/* Ramp generator motion control registers */
	RAMPMODE = 0x20 // Driving mode (Velocity, Positioning, Hold)
	XACTUAL  = 0x21 // Actual motor position
	VACTUAL  = 0x22 // Actual  motor  velocity  from  ramp  generator
	VSTART   = 0x23 // Motor start velocity
	A_1      = 0x24 // First acceleration between VSTART and V1
	V_1      = 0x25 // First acceleration/deceleration phase target velocity
	AMAX     = 0x26 // Second acceleration between V1 and VMAX
	VMAX     = 0x27 // Target velocity in velocity mode
	DMAX     = 0x28 // Deceleration between VMAX and V1
	D_1      = 0x2A // Deceleration between V1 and VSTOP
	//Attention:  Do  not  set  0  in  positioning  mode, even if V1=0!
	VSTOP = 0x2B // Motor stop velocity
	//Attention: Set VSTOP > VSTART!
	//Attention:  Do  not  set  0  in  positioning  mode, minimum 10 recommend!
	TZEROWAIT = 0x2C // Waiting time after ramping down to zero velocity before next movement or direction inversion can start.
	XTARGET   = 0x2D // Target position for ramp mode

	/* Ramp generator driver feature control registers */
	VDCMIN    = 0x33 // Velocity threshold for enabling automatic commutation dcStep
	SW_MODE   = 0x34 // Switch mode configuration
	RAMP_STAT = 0x35 // Ramp status and switch event status
	XLATCH    = 0x36 // Ramp generator latch position upon programmable switch event

	/* Encoder registers */
	ENCMODE       = 0x38 // Encoder configuration and use of N channel
	X_ENC         = 0x39 // Actual encoder position
	ENC_CONST     = 0x3A // Accumulation constant
	ENC_STATUS    = 0x3B // Encoder status information
	ENC_LATCH     = 0x3C // Encoder position latched on N event
	ENC_DEVIATION = 0x3D // Maximum number of steps deviation between encoder counter and XACTUAL for deviation warning

	/* Motor driver registers */
	MSLUT0          uint8 = 0x60 // 32 bits
	MSLUT1          uint8 = 0x61 // 32 bits
	MSLUT2          uint8 = 0x62 // 32 bits
	MSLUT3          uint8 = 0x63 // 32 bits
	MSLUT4          uint8 = 0x64 // 32 bits
	MSLUT5          uint8 = 0x65 // 32 bits
	MSLUT6          uint8 = 0x66 // 32 bits
	MSLUT7          uint8 = 0x67 // 32 bits
	MSLUTSEL              = 0x68 // Look up table segmentation definition
	MSLUTSTART            = 0x69 // Absolute current at microstep table entries 0 and 256
	MSCNT                 = 0x6A // Actual position in the microstep table
	MSCURACT              = 0x6B // Actual microstep current
	CHOPCONF              = 0x6C // Chopper and driver configuration
	COOLCONF              = 0x6D // coolStep smart current control register and stallGuard2 configuration
	DCCTRL                = 0x6E // dcStep automatic commutation configuration register
	DRV_STATUS            = 0x6F // stallGuard2 value and driver error flags
	PWMCONF               = 0x70 // stealthChop voltage PWM mode chopper configuration
	PWM_SCALE             = 0x71 // Results of stealthChop amplitude regulator.
	PWM_AUTO              = 0x72 // Automatically determined PWM config values
	LOST_STEPS            = 0x73 // Number of input steps skipped due to dcStep. only with SD_MODE = 1
	expectedVersion       = 0x03
	DEFAULT_F_CLK         = 12000000
)
