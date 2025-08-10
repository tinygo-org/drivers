//go:build tinygo

package tmc5160

import (
	"github.com/orsinium-labs/tinymath"
	"machine"
)

type Driver struct {
	comm      RegisterComm
	address   uint8
	enablePin machine.Pin
	stepper   Stepper
}

func NewDriver(comm RegisterComm, address uint8, enablePin machine.Pin, stepper Stepper) *Driver {
	return &Driver{
		comm:      comm,
		address:   address,
		enablePin: enablePin,
		stepper:   stepper,
	}
}

// WriteRegister sends a register write command to the Driver.
func (driver *Driver) WriteRegister(reg uint8, value uint32) error {
	if driver.comm == nil {
		return CustomError("communication interface not set")
	}
	// Use the communication interface (RegisterComm) to write the register
	return driver.comm.WriteRegister(reg, value, driver.address)
}

// ReadRegister sends a register read command to the Driver and returns the read value.
func (driver *Driver) ReadRegister(reg uint8) (uint32, error) {
	if driver.comm == nil {
		return 0, CustomError("communication interface not set")
	}
	// Use the communication interface (RegisterComm) to read the register
	return driver.comm.ReadRegister(reg, driver.address)
}

// Begin initializes the Driver driver with power and motor parameters
func (driver *Driver) Begin(powerParams PowerStageParameters, motorParams MotorParameters, stepperDirection MotorDirection) bool {
	// Clear the reset and charge pump undervoltage flags
	gstat := NewGSTAT()
	gstat.Reset = true
	gstat.UvCp = true
	err := driver.WriteRegister(GSTAT, gstat.Pack())
	if err != nil {
		return false
	}

	// Configure driver settings
	drvConf := NewDRV_CONF()
	drvConf.DrvStrength = constrain(powerParams.drvStrength, 0, 3)
	drvConf.BBMTime = constrain(powerParams.bbmTime, 0, 24)
	drvConf.BBMClks = constrain(powerParams.bbmClks, 0, 15)
	err = driver.WriteRegister(DRV_CONF, drvConf.Pack())
	if err != nil {
		return false
	}

	// Set global scaler
	err = driver.WriteRegister(GLOBAL_SCALER, uint32(constrain(motorParams.globalScaler, 32, 256)))
	if err != nil {
		return false
	}

	// Set initial currents and delay
	iholdrun := NewIHOLD_IRUN()
	iholdrun.Ihold = constrain(motorParams.ihold, 0, 31)
	iholdrun.Ihold = constrain(motorParams.irun, 0, 31)
	iholdrun.IholdDelay = 7
	err = driver.WriteRegister(IHOLD_IRUN, iholdrun.Pack())
	if err != nil {
		return false
	}

	// Set PWM configuration values
	pwmconf := NewPWMCONF()
	err = driver.WriteRegister(PWMCONF, 0xC40C001E)
	if err != nil {
		return false
	} // Reset default value pwm_ofs = 196,pwm_grad = 12,pwm_freq = 0, pwm_autoscale = false, pwm_autograd = false,freewheel = 3
	pwmconf.PwmAutoscale = false // Temporarily set to false for setting OFS and GRAD values
	_fclk := int(driver.stepper.Fclk) * 1000000
	if _fclk > DEFAULT_F_CLK {
		pwmconf.PwmFreq = 0
	} else {
		pwmconf.PwmFreq = 0b01 // Recommended: 35kHz with internal 12MHz clock
	}
	pwmconf.PwmGrad = uint8(motorParams.pwmGradInitial)
	pwmconf.PwmOfs = uint8(motorParams.pwmOfsInitial)
	pwmconf.Freewheel = motorParams.freewheeling
	err = driver.WriteRegister(PWMCONF, pwmconf.Pack())
	if err != nil {
		return false
	}

	// Enable PWM auto-scaling and gradient adjustment
	pwmconf.PwmAutoscale = true
	pwmconf.PwmAutograd = true
	err = driver.WriteRegister(PWMCONF, pwmconf.Pack())
	if err != nil {
		return false
	}

	// Recommended chop configuration settings
	_chopConf := NewCHOPCONF()
	_chopConf.Toff = 5
	_chopConf.Tbl = 2
	_chopConf.HstrtTfd = 4
	_chopConf.HendOffset = 0
	err = driver.WriteRegister(CHOPCONF, _chopConf.Pack())
	if err != nil {
		return false
	}
	rampMode := NewRAMPMODE(driver.comm, driver.address)
	rampMode.SetMode(PositioningMode)
	gconf := NewGCONF()
	gconf.EnPwmMode = true // Enable stealthChop PWM mode
	gconf.Shaft = stepperDirection == Clockwise
	err = driver.WriteRegister(GCONF, gconf.Pack())
	if err != nil {
		return false
	}

	// Set default start, stop, threshold speeds
	driver.setRampSpeeds(0.0, 0.1, 0.0) // Start, stop, threshold speeds

	// Set default D1 (must not be = 0 in positioning mode even with V1=0)
	err = driver.WriteRegister(D_1, 100)
	if err != nil {
		return false
	}

	return false
}
func (driver *Driver) setRampSpeeds(startSpeed float32, stopSpeed float32, transitionSpeed float32) {
	str := driver.stepper.DesiredSpeedToTSTEP(uint32(startSpeed))
	stp := driver.stepper.DesiredSpeedToTSTEP(uint32(stopSpeed))
	ts := driver.stepper.DesiredSpeedToTSTEP(uint32(transitionSpeed))
	driver.WriteRegister(VSTART, uint32(tinymath.Min(0x3FFFF, float32(str))))
	driver.WriteRegister(VSTOP, uint32(tinymath.Min(0x3FFFF, float32(stp))))
	driver.WriteRegister(V_1, uint32(tinymath.Min(0xFFFFF, float32(ts))))
	println("Ramp set to: startSpeed:", startSpeed, "stopSpeed:", stopSpeed, "transitionSpeed:", transitionSpeed)
}

// setMaxSpeed sets the maximum speed to 0 (placeholder function)
func setMaxSpeed(speed uint32) {
	// This is a placeholder function that sets the speed
	// Implement the actual logic to set the maximum speed register value
	println("Max Speed set to:", speed)
}

// Dump_TMC reads multiple registers from the Driver and logs their values with their names.
func (driver *Driver) Dump_TMC() error {
	registers := []uint8{
		GCONF, CHOPCONF, GSTAT, DRV_STATUS, FACTORY_CONF, IOIN, LOST_STEPS, MSCNT,
		MSCURACT, OTP_READ, PWM_SCALE, PWM_AUTO, TSTEP,
	}
	registerNames := map[uint8]string{
		GCONF:        "GCONF",
		CHOPCONF:     "CHOPCONF",
		GSTAT:        "GSTAT",
		DRV_STATUS:   "DRV_STATUS",
		FACTORY_CONF: "FACTORY_CONF",
		IOIN:         "IOIN",
		LOST_STEPS:   "LOST_STEPS",
		MSCNT:        "MSCNT",
		MSCURACT:     "MSCURACT",
		OTP_READ:     "OTP_READ",
		PWM_SCALE:    "PWM_SCALE",
		PWM_AUTO:     "PWM_AUTO",
		TSTEP:        "TSTEP",
	}
	for _, reg := range registers {
		// Fetch the register name from the map
		regName, exists := registerNames[reg]
		if !exists {
			regName = "Unknown Register"
		}
		val, err := driver.ReadRegister(reg)
		if err != nil {
			println("Error reading register", regName, err)
			return err
		}
		println("Register", regName, "Value:", val)
	}

	return nil
}
