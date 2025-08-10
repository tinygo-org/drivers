package tmc5160

const maxVMAX = 8388096

// PowerStageParameters represents the power stage parameters
type PowerStageParameters struct {
	drvStrength uint8
	bbmTime     uint8
	bbmClks     uint8
}

// MotorParameters represents the motor parameters
type MotorParameters struct {
	globalScaler   uint16
	ihold          uint8
	irun           uint8
	iholddelay     uint8
	pwmGradInitial uint16
	pwmOfsInitial  uint16
	freewheeling   uint8
}

// MotorDirection defines motor direction constants
type MotorDirection uint8

const (
	Clockwise MotorDirection = iota
	CounterClockwise
)

const (
	// Common stepper motor angles
	StepAngle_1_8  = 1.8
	StepAngle_0_9  = 0.9
	StepAngle_0_72 = 0.72
	StepAngle_1_2  = 1.2
	StepAngle_0_48 = 0.48

	// Common microstepping options
	Step_1   uint8 = 1
	Step_2   uint8 = 2
	Step_4   uint8 = 4
	Step_8   uint8 = 8
	Step_16  uint8 = 16
	Step_32  uint8 = 32
	Step_64  uint8 = 64
	Step_128 uint8 = 128
)

const (
	DefaultAngle     float32 = StepAngle_1_8
	DefaultGearRatio float32 = 1.0
	DefaultVSupply   float32 = 12.0
	DefaultRCoil     float32 = 1.2
	DefaultLCoil     float32 = 0.005
	DefaultIPeak     float32 = 2.0
	DefaultRSense    float32 = 0.1
	DefaultFclk      uint8   = 12
	DefaultStep_256          = 256
)

type Stepper struct {
	Angle       float32
	GearRatio   float32
	VelocitySPS float32 //  Velocity in Steps per sec
	VSupply     float32
	RCoil       float32
	LCoil       float32
	IPeak       float32
	RSense      float32
	MSteps      uint8
	Fclk        uint8 //Clock in Mhz

}

// NewStepper function initializes a Stepper with default values used for testing
func NewDefaultStepper() Stepper {
	return Stepper{
		Angle:     StepAngle_1_8, // Default to 1.8 degrees
		GearRatio: 1.0,           // Default to no reduction (1:1)
		VSupply:   12.0,          // Default 12V supply
		RCoil:     1.2,           // Default coil resistance (1.2 ohms)
		LCoil:     0.005,         // Default coil inductance (5 mH)
		IPeak:     2.0,           // Default peak current (2A)
		RSense:    0.1,           // Default sense resistance (0.1 ohms)
		MSteps:    Step_16,       // Default 16 Microsteps
		Fclk:      DefaultFclk,
	}
}

// NewStepper initializes a Stepper with user-defined values
func NewStepper(angle float32, gearRatio, vSupply, rCoil, lCoil, iPeak, rSense float32, mSteps uint8, fclk uint8) Stepper {
	return Stepper{
		Angle:     angle,     // User-defined stepper angle (e.g., StepAngle_1_8)
		GearRatio: gearRatio, // User-defined gear ratio
		VSupply:   vSupply,   // User-defined supply voltage
		RCoil:     rCoil,     // User-defined coil resistance
		LCoil:     lCoil,     // User-defined coil inductance
		IPeak:     iPeak,     // User-defined peak current
		RSense:    rSense,    // User-defined sense resistance
		MSteps:    mSteps,    // User-defined microstepping setting
		Fclk:      fclk,      // User-defined clock frequency in MHz

	}
}
