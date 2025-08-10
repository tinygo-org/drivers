package tmc2209

import "log"

// TMC2209 Register addresses
const (
	GCONF           = 0x00
	GSTAT           = 0x01
	IFCNT           = 0x02
	IOIN            = 0x06
	IHOLD_IRUN      = 0x10
	TPOWERDOWN      = 0x11
	TSTEP           = 0x12
	TPWMTHRS        = 0x13
	TCOOLTHRS       = 0x14
	VACTUAL         = 0x22
	SGTHRS          = 0x40
	SG_RESULT       = 0x41
	COOLCONF        = 0x42
	MSCNT           = 0x6A
	MSCURACT        = 0x6B
	CHOPCONF        = 0x6C
	DRV_STATUS      = 0x6F
	PWMCONF         = 0x70
	PWM_SCALE       = 0x71
	PWM_AUTO        = 0x72
	expectedVersion = 0x03
)

type RegisterComm interface {
	ReadRegister(register uint8, driverIndex uint8) (uint32, error)
	WriteRegister(register uint8, value uint32, driverIndex uint8) error
}

// Register is an interface that all register structs will implement for generic access
type Register interface {
	Pack() uint32
	Unpack(value uint32)
	GetAddress() uint8
}

// NewRegister is a generic function to initialize any register with a given address
func NewRegister(registerAddr uint8) Register {
	switch registerAddr {
	case IOIN:
		return NewIoin()
	case PWMCONF:
		return NewPWMConf()
	case GCONF:
		return NewGconf()
	case GSTAT:
		return NewGstat()
	case IFCNT:
		return NewIfcnt()
	case IHOLD_IRUN:
		return NewIholdIrun()
	case TPOWERDOWN:
		return NewTpowerdown()
	case TSTEP:
		return NewTstep()
	case TPWMTHRS:
		return NewTpwmthrs()
	case TCOOLTHRS:
		return NewTcoolthrs()
	case VACTUAL:
		return NewVactual()
	case SGTHRS:
		return NewSgthrs()
	case SG_RESULT:
		return NewSgResult()
	case COOLCONF:
		return NewCoolConf()
	case MSCNT:
		return NewMscnt()
	case MSCURACT:
		return NewMscuract()
	case CHOPCONF:
		return NewChopconf()
	case DRV_STATUS:
		return NewDrvStatus()
	case PWM_SCALE:
		return NewPwmScale()
	case PWM_AUTO:
		return NewPwmAuto()
	default:
		return nil
	}
}

// ReadRegister function using the register constants
func ReadRegister(comm RegisterComm, driverIndex uint8, register uint8) (uint32, error) {
	// Read the register value using the comm interface

	value, err := comm.ReadRegister(register, driverIndex)
	log.Printf("Request read ", register, driverIndex, value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// WriteRegister function using the register constants
func WriteRegister(comm RegisterComm, register uint8, driverIndex uint8, value uint32) error {
	// Write the value to the register using the comm interface
	return comm.WriteRegister(register, value, driverIndex)
}

// Ioin represents the fields of the IOIN register (0x06) in the TMC2209
//
// The IOIN register provides access to various inputs and control signals
// that the driver uses to determine its current state. This register contains
// bits related to motor driver status, step input, direction, diagnostics,
// and other control signals.
//
// The structure represents each field in the IOIN register and provides
// methods to pack and unpack these fields into the `Bytes` field (a 32-bit
// packed representation of all fields) for easier access and manipulation.
//
// Fields:
// - Enn:       Enables the driver. 1 = Driver enabled, 0 = Driver disabled.
// - Reserved0: Reserved bit, must be set to 0.
// - Ms1:       Microstep setting, first bit of the microstep configuration.
// - Ms2:       Microstep setting, second bit of the microstep configuration.
// - Diag:      Diagnostics flag. Used for fault detection and error reporting.
// - Reserved1: Reserved bit, must be set to 0.
// - PdnSerial: Power-down (sleep) state for the UART interface. 1 = Power down, 0 = Active.
// - Step:      Step input signal. 1 = Step signal active, 0 = No step signal.
// - SpreadEn:  SpreadCycle enable. 1 = Enable SpreadCycle, 0 = Disable SpreadCycle.
// - Dir:       Direction input. 1 = Reverse direction, 0 = Forward direction.
// - Reserved2: Reserved bits, must be set to 0.
// - Version:   Driver version information. The version of the IOIN register in the TMC2209 chip.
//
// `Bytes`: A 32-bit value that stores the packed representation of the IOIN register.
// The `Bytes` field is used to manipulate the register's value as a single 32-bit value,
// allowing you to read and write it more efficiently.
type Ioin struct {
	Enn          uint32 // 1-bit field: Driver enable status (1 = enabled, 0 = disabled)
	Reserved0    uint32 // 1-bit field: Reserved, should always be 0
	Ms1          uint32 // 1-bit field: Microstep setting (first bit)
	Ms2          uint32 // 1-bit field: Microstep setting (second bit)
	Diag         uint32 // 1-bit field: Diagnostics flag (error reporting)
	Reserved1    uint32 // 1-bit field: Reserved, should always be 0
	PdnSerial    uint32 // 1-bit field: Power-down state for the UART interface (1 = power down, 0 = active)
	Step         uint32 // 1-bit field: Step signal input (1 = active, 0 = inactive)
	SpreadEn     uint32 // 1-bit field: SpreadCycle enable (1 = enabled, 0 = disabled)
	Dir          uint32 // 1-bit field: Direction input (1 = reverse, 0 = forward)
	Reserved2    uint32 // 14-bit field: Reserved bits, should always be 0
	Version      uint32 // 8-bit field: Version information for the driver
	Bytes        uint32 // 32-bit field: Packed representation of the IOIN register (all fields packed into a single 32-bit value)
	RegisterAddr uint8  // The address of the register, in this case, IOIN (0x06)
}

// Pack the individual fields into the Bytes field (a single 32-bit value).
// This method combines all the individual fields (like Enn, Ms1, etc.)
// into a packed 32-bit value that can be written to the register.
func (ioin *Ioin) Pack() uint32 {
	ioin.Bytes = (ioin.Enn & 0x01) | // Enn field (1 bit)
		((ioin.Reserved0 & 0x01) << 1) | // Reserved0 field (1 bit)
		((ioin.Ms1 & 0x01) << 2) | // Ms1 field (1 bit)
		((ioin.Ms2 & 0x01) << 3) | // Ms2 field (1 bit)
		((ioin.Diag & 0x01) << 4) | // Diag field (1 bit)
		((ioin.Reserved1 & 0x01) << 5) | // Reserved1 field (1 bit)
		((ioin.PdnSerial & 0x01) << 6) | // PdnSerial field (1 bit)
		((ioin.Step & 0x01) << 7) | // Step field (1 bit)
		((ioin.SpreadEn & 0x01) << 8) | // SpreadEn field (1 bit)
		((ioin.Dir & 0x01) << 9) | // Dir field (1 bit)
		((ioin.Reserved2 & 0x3FFF) << 10) | // Reserved2 field (14 bits)
		((ioin.Version & 0xFF) << 24) // Version field (8 bits)
	return ioin.Bytes
}

// Unpack the Bytes field into the individual fields.
// This method takes the packed 32-bit value from the Bytes field and extracts
// the individual register fields into their corresponding variables.
func (ioin *Ioin) Unpack(uint32) {
	ioin.Enn = ioin.Bytes & 0x01
	ioin.Reserved0 = (ioin.Bytes >> 1) & 0x01
	ioin.Ms1 = (ioin.Bytes >> 2) & 0x01
	ioin.Ms2 = (ioin.Bytes >> 3) & 0x01
	ioin.Diag = (ioin.Bytes >> 4) & 0x01
	ioin.Reserved1 = (ioin.Bytes >> 5) & 0x01
	ioin.PdnSerial = (ioin.Bytes >> 6) & 0x01
	ioin.Step = (ioin.Bytes >> 7) & 0x01
	ioin.SpreadEn = (ioin.Bytes >> 8) & 0x01
	ioin.Dir = (ioin.Bytes >> 9) & 0x01
	ioin.Reserved2 = (ioin.Bytes >> 10) & 0x3FFF
	ioin.Version = (ioin.Bytes >> 24) & 0xFF
}
func (ioin *Ioin) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, ioin.RegisterAddr)
}
func (ioin *Ioin) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, ioin.RegisterAddr, driverIndex, value)
}
func (ioin *Ioin) GetAddress() uint8 {
	return ioin.RegisterAddr
}
func NewIoin() *Ioin {
	return &Ioin{
		RegisterAddr: IOIN,
	}
}

// PWMConf represents the fields in the TMC2209 PWMCONF register.
//
// This register controls the Pulse Width Modulation (PWM) configuration for the stepper motor.
// It defines parameters related to the motor's driving mechanism, including PWM offset,
// gradient, frequency, and scaling. These parameters allow fine control over the motor's
// behavior, particularly in the context of energy saving modes like StealthChop.
//
// The fields are as follows:
//   - **PwmOfs** (8 bits): The offset for PWM, affecting the motor's drive current.
//   - **PwmGrad** (8 bits): The gradient applied to the PWM waveform. It determines the
//     rise and fall of the PWM signal, influencing motor smoothness and efficiency.
//   - **PwmFreq** (2 bits): PWM frequency control, affecting the switching rate of the motor.
//     This value influences the motor's operating frequency and noise characteristics.
//   - **PwmAutoscale** (1 bit): This flag enables automatic scaling of the PWM amplitude
//     based on the motor’s load. When enabled, the motor adjusts its drive based on current
//     requirements, improving energy efficiency and reducing heat generation.
//   - **PwmAutograd** (1 bit): Similar to `PwmAutoscale`, but focuses on the gradient
//     of the PWM signal, adjusting for the motor's behavior dynamically to maintain efficient
//     operation under varying load conditions.
//   - **Freewheel** (2 bits): Determines the state of the motor's freewheel functionality,
//     which allows the motor to coast when not actively driven, thereby reducing power consumption.
//   - **PwmReg** (4 bits): Register for fine-tuning the PWM signal’s behavior. It is typically used
//     for customizing the PWM waveform for optimal performance in different conditions.
//   - **PwmLim** (4 bits): PWM limit for the motor’s current control. It provides an upper
//     bound for the current drive to prevent overheating or excessive current draw.
//
// The `PWMCONF` register allows for precise control over the motor's electrical characteristics,
// contributing to better performance and energy efficiency, particularly in StealthChop mode.
type PWMConf struct {
	PwmOfs       uint32 // 8 bits
	PwmGrad      uint32 // 8 bits
	PwmFreq      uint32 // 2 bits
	PwmAutoscale uint32 // 1 bit
	PwmAutograd  uint32 // 1 bit
	Freewheel    uint32 // 2 bits
	PwmReg       uint32 // 4 bits
	PwmLim       uint32 // 4 bits
	Bytes        uint32 // 32-bit packed representation of all fields
	RegisterAddr uint8
}

func (pwm *PWMConf) GetAddress() uint8 {
	return pwm.RegisterAddr
}

// NewPWMConf Initialize PWMConf with register address
func NewPWMConf() *PWMConf {
	return &PWMConf{
		RegisterAddr: PWMCONF,
	}
}

// Pack Method to pack the fields into the Bytes field
func (pwm *PWMConf) Pack() uint32 {
	// Pack the individual fields into the Bytes field
	pwm.Bytes = (pwm.PwmOfs & 0xFF) |
		((pwm.PwmGrad & 0xFF) << 8) |
		((pwm.PwmFreq & 0x03) << 16) |
		((pwm.PwmAutoscale & 0x01) << 18) |
		((pwm.PwmAutograd & 0x01) << 19) |
		((pwm.Freewheel & 0x03) << 20) |
		((pwm.PwmReg & 0x0F) << 24) |
		((pwm.PwmLim & 0x0F) << 28) // PWM_LIM is 4 bits and goes in the last 4 bits
	return pwm.Bytes
}

// Unpack Method to unpack the Bytes field into individual fields
func (pwm *PWMConf) Unpack(uint32) {
	// Unpack the Bytes field into individual fields
	pwm.PwmOfs = pwm.Bytes & 0xFF
	pwm.PwmGrad = (pwm.Bytes >> 8) & 0xFF
	pwm.PwmFreq = (pwm.Bytes >> 16) & 0x03
	pwm.PwmAutoscale = (pwm.Bytes >> 18) & 0x01
	pwm.PwmAutograd = (pwm.Bytes >> 19) & 0x01
	pwm.Freewheel = (pwm.Bytes >> 20) & 0x03
	pwm.PwmReg = (pwm.Bytes >> 24) & 0x0F
	pwm.PwmLim = (pwm.Bytes >> 28) & 0x0F
}
func (pwm *PWMConf) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, pwm.RegisterAddr)
}
func (pwm *PWMConf) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, driverIndex, pwm.RegisterAddr, value)
}

// Chopconf represents the fields in the TMC2209 CHOPCONF register.
//
// The CHOPCONF register configures the chopping control for the stepper motor.
// It defines parameters related to the current waveform, step behavior, and
// microstepping resolution. These parameters directly affect the motor's efficiency,
// torque output, and smoothness, especially when operating in StealthChop mode.
//
// The fields are as follows:
//   - **Toff** (4 bits): The Off-time parameter for the chopping waveform.
//     It sets the duration of the 'off' period of the waveform, which controls how long
//     the motor driver remains idle between each pulse. Adjusting this value affects motor efficiency
//     and noise characteristics. Lower values allow faster switching and more responsive motors,
//     while higher values can improve efficiency at the cost of motor speed.
//   - **Hstrt** (3 bits): The Hysteresis start value. This parameter controls the start
//     of the hysteresis in the chopping process. It determines when the current will be limited
//     based on the hysteresis function. The higher the value, the more gradual the transition
//     between phases.
//   - **Hend** (4 bits): The Hysteresis end value. This value defines the end of the hysteresis
//     region, where the current is stabilized. Like `Hstrt`, higher values result in a smoother
//     transition in current regulation and lower torque ripple.
//   - **Tbl** (2 bits): The Table selection for the internal chopping table, affecting
//     the switching behavior and optimization of the motor’s current profile.
//     This setting is typically tuned for different load conditions or step modes.
//   - **Vsense** (1 bit): This field selects the sense resistor mode. If enabled, the driver
//     uses the internal sense resistors for current sensing, which enables more precise current
//     regulation and monitoring, enhancing the motor’s performance and efficiency.
//   - **Mres** (4 bits): The microstepping resolution. This value controls how many steps the motor
//     takes per full rotation. Higher values allow for finer stepping, leading to smoother motion
//     but at the cost of reduced torque at higher microstep resolutions.
//   - **Intpol** (1 bit): This field enables interpolation in the motor driver, which improves
//     smoothness by interpolating intermediate microsteps between full steps. It helps reduce
//     torque ripple and improve overall motor performance.
//   - **Dedge** (1 bit): This parameter enables the detection of a specific edge (rising or falling)
//     in the step pulse, which is used to control the timing of motor transitions.
//     It is generally used for fine-tuning performance or reducing noise during operation.
//   - **Diss2g** (1 bit): This field, when enabled, disables the second MOSFET during idle phases
//     in the motor operation, further improving efficiency and reducing heat generation.
//   - **Diss2vs** (1 bit): Similar to `Diss2g`, this parameter disables the second MOSFET during
//     specific phases of operation to reduce losses, increase efficiency, and improve motor control
//     in certain scenarios, particularly during low-speed or low-power operation.
//
// The `CHOPCONF` register plays a crucial role in fine-tuning motor operation, ensuring
// smooth motion, efficient power consumption, and minimizing torque ripple in the system.
// These settings are particularly useful when transitioning between operating modes like StealthChop.
type Chopconf struct {
	Toff         uint32
	Hstrt        uint32
	Hend         uint32
	Tbl          uint32
	Vsense       uint32
	Mres         uint32
	Intpol       uint32
	Dedge        uint32
	Diss2g       uint32
	Diss2vs      uint32
	Bytes        uint32 // The packed 32-bit value
	RegisterAddr uint8
}

func (chopconf *Chopconf) GetAddress() uint8 {
	return chopconf.RegisterAddr
}

// Pack the individual fields into the Bytes field (a single 32-bit value).
func (chopconf *Chopconf) Pack() uint32 {
	chopconf.Bytes = (chopconf.Toff & 0x0F) |
		((chopconf.Hstrt & 0x07) << 4) |
		((chopconf.Hend & 0x0F) << 7) |
		((chopconf.Tbl & 0x03) << 15) |
		((chopconf.Vsense & 0x01) << 17) |
		((chopconf.Mres & 0x0F) << 24) |
		((chopconf.Intpol & 0x01) << 28) |
		((chopconf.Dedge & 0x01) << 29) |
		((chopconf.Diss2g & 0x01) << 30) |
		((chopconf.Diss2vs & 0x01) << 31)
	return chopconf.Bytes
}

// Unpack the Bytes field into the individual fields.
func (chopconf *Chopconf) Unpack(uint32) {
	chopconf.Toff = chopconf.Bytes & 0x0F
	chopconf.Hstrt = (chopconf.Bytes >> 4) & 0x07
	chopconf.Hend = (chopconf.Bytes >> 7) & 0x0F
	chopconf.Tbl = (chopconf.Bytes >> 15) & 0x03
	chopconf.Vsense = (chopconf.Bytes >> 17) & 0x01
	chopconf.Mres = (chopconf.Bytes >> 24) & 0x0F
	chopconf.Intpol = (chopconf.Bytes >> 28) & 0x01
	chopconf.Dedge = (chopconf.Bytes >> 29) & 0x01
	chopconf.Diss2g = (chopconf.Bytes >> 30) & 0x01
	chopconf.Diss2vs = (chopconf.Bytes >> 31) & 0x01
}
func NewChopconf() *Chopconf {
	return &Chopconf{
		RegisterAddr: CHOPCONF,
	}
}
func (chopconf *Chopconf) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, chopconf.RegisterAddr)
}
func (chopconf *Chopconf) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, chopconf.RegisterAddr, driverIndex, value)
}

// Gstat represents the fields in the TMC2209 GSTAT register.
type Gstat struct {
	Reset        uint32
	DrvErr       uint32
	UvCp         uint32
	Reserved     uint32
	Bytes        uint32 // The packed 32-bit value
	RegisterAddr uint8
}

func (gstat *Gstat) GetAddress() uint8 {
	return gstat.RegisterAddr
}

// Pack the individual fields into the Bytes field (a single 32-bit value).
func (gstat *Gstat) Pack() uint32 {
	gstat.Bytes = (gstat.Reset & 0x01) |
		((gstat.DrvErr & 0x01) << 1) |
		((gstat.UvCp & 0x01) << 2) |
		((gstat.Reserved & 0x1FFFFF) << 3) // 21 bits reserved
	return gstat.Bytes
}

// Unpack the Bytes field into the individual fields.
func (gstat *Gstat) Unpack(uint32) {
	gstat.Reset = gstat.Bytes & 0x01
	gstat.DrvErr = (gstat.Bytes >> 1) & 0x01
	gstat.UvCp = (gstat.Bytes >> 2) & 0x01
	gstat.Reserved = (gstat.Bytes >> 3) & 0x1FFFFF
}
func NewGstat() *Gstat {
	return &Gstat{
		RegisterAddr: GSTAT,
	}
}
func (gstat *Gstat) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, gstat.RegisterAddr)
}
func (gstat *Gstat) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, gstat.RegisterAddr, driverIndex, value)
}

// Gconf represents the fields in the TMC2209 GCONF register.
//
// The GCONF register configures global settings for the motor driver.
// It controls various aspects of the operation, including stepper driver modes,
// fault detection, and other motor control behaviors that influence motor performance,
// efficiency, and system diagnostics. The register also enables or disables certain features
// based on specific application requirements.
//
// The fields are as follows:
//   - **IScaleAnalog** (1 bit): This field controls the scaling of the analog current sense.
//     If enabled, it adjusts the scaling of the motor current detection circuitry, improving
//     the precision of current regulation. This is important for accurate motor control and
//     better torque handling, especially at low speeds.
//   - **InternalRsense** (1 bit): When enabled, this field indicates that the internal sense
//     resistors are used for current sensing. This allows for improved accuracy in current
//     measurement and feedback, resulting in better overall motor performance and energy efficiency.
//   - **EnSpreadcycle** (1 bit): This field enables SpreadCycle operation mode. When enabled,
//     it switches the motor driver to operate in SpreadCycle mode, which is the traditional
//     chopper operation mode. This mode typically results in higher efficiency and more precise
//     current control, but it may produce more noise compared to StealthChop mode.
//   - **Shaft** (1 bit): This field enables the detection of a shaft (step) in the system, which
//     can be used to detect whether the motor is moving or has completed a cycle. It is used
//     to improve system control and diagnostics in motor-driven applications.
//   - **IndexOtpw** (1 bit): This field, when set, allows detection of the over-temperature
//     warning flag (otpw), indicating that the motor or the driver has exceeded safe operating
//     temperatures. If enabled, the system can take actions like reducing speed or stopping the motor
//     to avoid damage due to overheating.
//   - **IndexStep** (1 bit): This field enables indexing of the motor steps. When enabled, it
//     allows for step detection, which is helpful in stepper motors where exact step counts are critical.
//     This setting ensures accurate positioning of the motor in precise step operations.
//   - **PdnDisable** (1 bit): This field enables or disables the power-down functionality for
//     the driver. When enabled, the driver will power down the motor and enter a low-power state.
//     This is typically used to reduce energy consumption when the motor is not active or when the
//     system is in idle mode. This helps save power during periods of non-operation.
//   - **MstepRegSelect** (1 bit): This field allows the selection of the multistep regulation
//     mode. When enabled, the driver uses the multistep algorithm to control current regulation,
//     which can help reduce torque ripple and improve motor smoothness, particularly in applications
//     requiring high precision or smooth motion.
//   - **MultistepFilt** (1 bit): When enabled, this field activates the multistep filtering.
//     This feature is used to reduce the effect of noise and fluctuations in the current waveform.
//     It smooths out the power supply and motor performance, particularly at higher speeds or during
//     high-load conditions.
//   - **Reserved** (21 bits): These reserved bits are unused and should always be set to zero.
//     They ensure backward compatibility and alignment with future versions of the register.
//     They have no effect on the system's operation.
type Gconf struct {
	IScaleAnalog   uint32
	InternalRsense uint32
	EnSpreadcycle  uint32
	Shaft          uint32
	IndexOtpw      uint32
	IndexStep      uint32
	PdnDisable     uint32
	MstepRegSelect uint32
	MultistepFilt  uint32
	Reserved       uint32
	Bytes          uint32 // The packed 32-bit value
	RegisterAddr   uint8
}

func (gconf *Gconf) GetAddress() uint8 {
	return gconf.RegisterAddr
}

// Pack the individual fields into the Bytes field (a single 32-bit value).
func (gconf *Gconf) Pack() uint32 {
	gconf.Bytes = (gconf.IScaleAnalog & 0x01) |
		((gconf.InternalRsense & 0x01) << 1) |
		((gconf.EnSpreadcycle & 0x01) << 2) |
		((gconf.Shaft & 0x01) << 3) |
		((gconf.IndexOtpw & 0x01) << 4) |
		((gconf.IndexStep & 0x01) << 5) |
		((gconf.PdnDisable & 0x01) << 6) |
		((gconf.MstepRegSelect & 0x01) << 7) |
		((gconf.MultistepFilt & 0x01) << 8) |
		((gconf.Reserved & 0x1FFFFF) << 9) // 21 bits reserved
	return gconf.Bytes
}

// Unpack the Bytes field into the individual fields.
func (gconf *Gconf) Unpack(uint32) {
	gconf.IScaleAnalog = gconf.Bytes & 0x01
	gconf.InternalRsense = (gconf.Bytes >> 1) & 0x01
	gconf.EnSpreadcycle = (gconf.Bytes >> 2) & 0x01
	gconf.Shaft = (gconf.Bytes >> 3) & 0x01
	gconf.IndexOtpw = (gconf.Bytes >> 4) & 0x01
	gconf.IndexStep = (gconf.Bytes >> 5) & 0x01
	gconf.PdnDisable = (gconf.Bytes >> 6) & 0x01
	gconf.MstepRegSelect = (gconf.Bytes >> 7) & 0x01
	gconf.MultistepFilt = (gconf.Bytes >> 8) & 0x01
	gconf.Reserved = (gconf.Bytes >> 9) & 0x1FFFFF
}
func NewGconf() *Gconf {
	return &Gconf{
		RegisterAddr: GCONF,
	}
}
func (gconf *Gconf) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, gconf.RegisterAddr, driverIndex)
}
func (gconf *Gconf) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, gconf.RegisterAddr, driverIndex, value)
}

// Ifcnt represents the fields in the TMC2209 IFCNT register.
//
// The IFCNT register is used to monitor the input frequency of the step signal. It holds the
// count of the number of steps that have been processed by the driver and provides
// valuable information about the frequency of step pulses being received. This can be useful
// for diagnostic purposes, to measure the actual step rate, or to check the system’s input
// signal behavior. It helps in monitoring the stepper signal's integrity and rate of pulses
// generated by the control system.
//
// The field is as follows:
//   - **Ifcnt** (8 bits): This 8-bit field holds the count of the received step pulses. The count
//     is incremented on each step pulse the driver receives and represents the number of steps
//     processed by the motor driver. The counter resets at the start of each pulse cycle. This
//     field can be used to verify the input signal's frequency and detect potential issues
//     such as missed or delayed pulses. Monitoring this register can help optimize the stepper
//     signal for better accuracy and performance.
type Ifcnt struct {
	Ifcnt        uint32 // 8-bit interface counter
	Reserved     uint32 // Reserved bits, here represented as uint32 for simplicity
	Bytes        uint32 // The packed 32-bit value
	RegisterAddr uint8
}

func (ifcnt *Ifcnt) GetAddress() uint8 {
	return ifcnt.RegisterAddr
}

// Pack the individual fields into the Bytes field (a single 32-bit value).
func (ifcnt *Ifcnt) Pack() uint32 {
	ifcnt.Bytes = (ifcnt.Ifcnt & 0xFF) | // 8 bits for the interface counter
		((ifcnt.Reserved & 0xFFFFFF) << 8) // Remaining bits for reserved (24 bits)
	return ifcnt.Bytes
}

// Unpack the Bytes field into the individual fields.
func (ifcnt *Ifcnt) Unpack(uint32) {
	ifcnt.Ifcnt = ifcnt.Bytes & 0xFF
	ifcnt.Reserved = (ifcnt.Bytes >> 8) & 0xFFFFFF
}

// Initialize IFCNT with register address
func NewIfcnt() *Ifcnt {
	return &Ifcnt{
		RegisterAddr: IFCNT, // IFCNT register address
	}
}
func (ifcnt *Ifcnt) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, ifcnt.RegisterAddr)
}
func (ifcnt *Ifcnt) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, ifcnt.RegisterAddr, driverIndex, value)
}

// IholdIrun  represents the fields in the TMC2209 IHOLD_IRUN register.
//
// The IHOLD_IRUN register is used to configure the current settings for the motor during
// different phases of operation. It specifically defines two current settings: the hold current
// and the run current.
//
//   - **Ihold** (5 bits): This field defines the current value for holding the motor in place when
//     it is idle. The value can be set between 0 and 31, with 0 representing the lowest possible current,
//     and 31 representing the maximum hold current. The hold current is typically lower than the run current
//     to minimize energy consumption when the motor is not actively moving.
//
//   - **Iruns** (5 bits): This field sets the current for the motor when it is running (i.e., moving).
//     The value can range from 0 to 31, and higher values represent stronger currents for increased torque
//     when the motor is under load. The run current is used when the motor is actively engaged in motion.
//
//   - **Iholddelay** (4 bits): This field defines the delay in microsteps before transitioning between
//     the hold current and the run current. This delay allows for smoother transitions when starting or stopping
//     the motor, preventing sudden jumps in current that could cause issues with the stepper system.
//
// This register is critical for efficient motor control, allowing dynamic management of motor currents to optimize
// energy consumption and torque output. Tuning the `Ihold` and `Iruns` settings can significantly impact motor
// performance and efficiency, especially in applications where energy efficiency is important.
type IholdIrun struct {
	Ihold        uint32 // 5 bits for hold current
	Irun         uint32 // 5 bits for run current
	Iholddelay   uint32 // 4 bits for hold delay
	Bytes        uint32 // The packed 32-bit value
	RegisterAddr uint8  // Register address
}

func (iholdIrun *IholdIrun) GetAddress() uint8 {
	return iholdIrun.RegisterAddr
}

// Pack the individual fields into the Bytes field (a single 32-bit value).
func (iholdIrun *IholdIrun) Pack() uint32 {
	iholdIrun.Bytes = (iholdIrun.Ihold & 0x1F) | // 5 bits for IHOLD
		((iholdIrun.Irun & 0x1F) << 5) | // 5 bits for IRUN
		((iholdIrun.Iholddelay & 0x0F) << 10) // 4 bits for IHOLDD_DELAY
	return iholdIrun.Bytes
}

// Unpack the Bytes field into the individual fields.
func (iholdIrun *IholdIrun) Unpack(uint32) {
	iholdIrun.Ihold = iholdIrun.Bytes & 0x1F
	iholdIrun.Irun = (iholdIrun.Bytes >> 5) & 0x1F
	iholdIrun.Iholddelay = (iholdIrun.Bytes >> 10) & 0x0F
}

// Initialize IHOLD_IRUN with register address
func NewIholdIrun() *IholdIrun {
	return &IholdIrun{
		RegisterAddr: IHOLD_IRUN, // IHOLD_IRUN register address
	}
}
func (iholdIrun *IholdIrun) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, iholdIrun.RegisterAddr)
}
func (iholdIrun *IholdIrun) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, iholdIrun.RegisterAddr, driverIndex, value)
}

// Tpwmthrs represents the fields in the TMC2209 TPWMTHRS register.
//
// The TPWMTHRS register is used to set the threshold for transitioning between
// traditional stepper driving (full stepping) and the use of PWM (Pulse Width Modulation)
// for controlling the stepper motor. This register allows for better current control
// and smoother operation, especially at higher speeds.
//
//   - **Tpwmthrs** (20 bits): This field defines the threshold value for switching
//     from traditional stepper driving to PWM control. The motor will operate with
//     standard stepping until the velocity exceeds the value set in this register,
//     at which point it will switch to PWM mode. This allows for improved efficiency
//     and smoother operation at higher speeds.
//
// The value for the `Tpwmthrs` field is typically set based on the specific motor and
// the desired speed range for the application. A higher threshold means the motor
// will operate in traditional stepping mode at lower speeds, while a lower threshold
// will enable PWM control at lower speeds for better motor control and efficiency.
type Tpwmthrs struct {
	Threshold    uint32 // 32-bit threshold value
	Bytes        uint32 // The packed 32-bit value
	RegisterAddr uint8  // Register address
}

func (tpwmthrs *Tpwmthrs) GetAddress() uint8 {
	return tpwmthrs.RegisterAddr
}

// Pack the individual fields into the Bytes field (a single 32-bit value).
func (tpwmthrs *Tpwmthrs) Pack() uint32 {
	tpwmthrs.Bytes = tpwmthrs.Threshold & 0xFFFFFFFF // 32-bit threshold value
	return tpwmthrs.Bytes
}

// Unpack the Bytes field into the individual fields.
func (tpwmthrs *Tpwmthrs) Unpack(uint32) {
	tpwmthrs.Threshold = tpwmthrs.Bytes & 0xFFFFFFFF
}

// NewTpwmthrs Initialize TPWMTHRS with register address
func NewTpwmthrs() *Tpwmthrs {
	return &Tpwmthrs{
		RegisterAddr: TPWMTHRS, // TPWMTHRS register address
	}
}
func (tpwmthrs *Tpwmthrs) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, tpwmthrs.RegisterAddr)
}
func (tpwmthrs *Tpwmthrs) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, tpwmthrs.RegisterAddr, driverIndex, value)
}

// Vactual represents the fields in the TMC2209 VACTUAL register.
//
// The VACTUAL register provides the actual velocity of the motor in terms of the
// stepper's step clock. It is used to monitor the real-time motor speed and is
// an important part of controlling the motor's behavior during operation.
//
//   - **Vactual** (20 bits): This field holds the current actual velocity of the motor.
//     It represents the stepper's velocity in terms of the step clock, providing feedback
//     on how fast the motor is turning. The value in this register is updated regularly
//     based on the microstepping and the actual motor speed.
//
// The `VACTUAL` register is particularly useful for closed-loop control systems,
// where the actual motor speed needs to be compared against the desired speed
// (set in other registers) to make adjustments. This can help in fine-tuning the
// motor's behavior for smoother operation and more accurate performance.
type Vactual struct {
	Velocity     uint32 // 32-bit velocity value
	Bytes        uint32 // The packed 32-bit value
	RegisterAddr uint8  // Register address
}

func (vactual *Vactual) GetAddress() uint8 {
	return vactual.RegisterAddr
}

// Pack the individual fields into the Bytes field (a single 32-bit value).
func (vactual *Vactual) Pack() uint32 {
	vactual.Bytes = vactual.Velocity & 0xFFFFFFFF // 32-bit velocity value
	return vactual.Bytes
}

// Unpack the Bytes field into the individual fields.
func (vactual *Vactual) Unpack(uint32) {
	vactual.Velocity = vactual.Bytes & 0xFFFFFFFF
}

// Initialize VACTUAL with register address
func NewVactual() *Vactual {
	return &Vactual{
		RegisterAddr: VACTUAL, // VACTUAL register address
	}
}
func (vactual *Vactual) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, vactual.RegisterAddr)
}
func (vactual *Vactual) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, vactual.RegisterAddr, driverIndex, value)
}

// Tcoolthrs represents the fields in the TMC2209 TCOOLTHRS register.
//
// The TCOOLTHRS register sets the lower threshold velocity for switching on
// smart energy CoolStep and StallGuard output to the DIAG pin. This helps
// manage the motor's energy usage by enabling CoolStep at higher velocities
// and StallGuard at lower velocities.
//
//   - **TCOOLTHRS** (20 bits): The threshold velocity (in units of steps per
//     microstep) used to switch between CoolStep and StallGuard modes. The motor
//     operates in CoolStep mode if the velocity exceeds this threshold and uses
//     StallGuard if the velocity falls below it.
//
// This register is important for optimizing motor performance, especially
// in applications that require energy efficiency. It ensures that the motor
// does not use unnecessary energy at low speeds while still being able to detect stalls
// and adjust torque at higher speeds. CoolStep allows the driver to dynamically adjust
// current levels based on the motor's actual load, reducing power consumption.
type Tcoolthrs struct {
	Velocity     uint32 // 20 bits for velocity
	Bytes        uint32 // Packed 32-bit value
	RegisterAddr uint8
}

func (tcoolthrs *Tcoolthrs) GetAddress() uint8 {
	return tcoolthrs.RegisterAddr
}

// Initialize TCOOLTHRS with register address
func NewTcoolthrs() *Tcoolthrs {
	return &Tcoolthrs{
		RegisterAddr: TCOOLTHRS,
	}
}

// Pack the individual fields into the Bytes field (a single 32-bit value).
func (tcoolthrs *Tcoolthrs) Pack() uint32 {
	tcoolthrs.Bytes = tcoolthrs.Velocity & 0xFFFFF // Keep only the lower 20 bits
	return tcoolthrs.Bytes
}

// Unpack the Bytes field into the individual fields.
func (tcoolthrs *Tcoolthrs) Unpack(uint32) {
	tcoolthrs.Velocity = tcoolthrs.Bytes & 0xFFFFF
}
func (tcoolthrs *Tcoolthrs) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, tcoolthrs.RegisterAddr)
}
func (tcoolthrs *Tcoolthrs) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, tcoolthrs.RegisterAddr, driverIndex, value)
}

// Sgthrs represents the fields in the TMC2209 SGTHRS register.
//
// The SGTHRS register sets the detection threshold for stall detection in the motor.
// It compares the StallGuard result (SG_RESULT) to twice the value set in this register.
// If the SG_RESULT value falls below the threshold (SG_RESULT < SGTHRS * 2), a stall
// is detected and can trigger a response such as a warning or motor shutdown.
//
// - **SGTHRS** (8 bits): This value sets the detection threshold for the StallGuard feature.
// The result of the StallGuard measurement (SG_RESULT) is compared to this threshold value
// multiplied by 2. When the SG_RESULT value is less than this threshold, a stall is detected.
//
// The `SGTHRS` register is part of the StallGuard feature, which allows the driver to
// detect and respond to motor stalls. StallGuard provides real-time feedback on the motor's
// status by measuring the motor’s back EMF to detect any irregularities that might indicate
// a stall condition. This can help in preventing mechanical damage by responding to stalls early.
type Sgthrs struct {
	Threshold    uint32 // 8 bits for threshold value
	Bytes        uint32 // Packed 32-bit value
	RegisterAddr uint8
}

func (sgthrs *Sgthrs) GetAddress() uint8 {
	return sgthrs.RegisterAddr
}

// Initialize SGTHRS with register address
func NewSgthrs() *Sgthrs {
	return &Sgthrs{
		RegisterAddr: SGTHRS,
	}
}

// Pack the individual fields into the Bytes field (a single 32-bit value).
func (sgthrs *Sgthrs) Pack() uint32 {
	sgthrs.Bytes = sgthrs.Threshold & 0xFF // Keep only the lower 8 bits
	return sgthrs.Bytes
}

// Unpack the Bytes field into the individual fields.
func (sgthrs *Sgthrs) Unpack(uint32) {
	sgthrs.Threshold = sgthrs.Bytes & 0xFF
}
func (sgthrs *Sgthrs) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, sgthrs.RegisterAddr)
}
func (sgthrs *Sgthrs) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, sgthrs.RegisterAddr, driverIndex, value)
}

// SgResult represents the fields in the TMC2209 SG_RESULT register.
//
// The SG_RESULT register stores the result of the StallGuard measurement, which is used
// to detect motor stalls. The StallGuard feature measures the motor's back electromotive force
// (back EMF) to assess the motor's load and detect irregularities that could indicate a stall.
//
// - **SG_RESULT** (10 bits): The register holds the result of the StallGuard measurement,
// which provides information about the motor's load and stall condition. A higher value in
// SG_RESULT generally means the motor is operating normally, while a lower value indicates
// the motor might be experiencing a stall or a mechanical blockage.
//
// The value in the SG_RESULT register is used in conjunction with the `SGTHRS` register
// (StallGuard Threshold) to determine if a stall condition has occurred. If the SG_RESULT
// value is below twice the value set in the `SGTHRS` register, a stall is considered to have occurred.
//
// The SG_RESULT register is particularly useful for motor stall detection, enabling the driver
// to protect the motor from damage caused by excessive load or mechanical binding by halting the motor's operation.
type SgResult struct {
	Result       uint32 // 10 bits for the result
	Bytes        uint32 // Packed 32-bit value
	RegisterAddr uint8
}

func (sgResult *SgResult) GetAddress() uint8 {
	return sgResult.RegisterAddr
}

// NewSgResult Initialize SG_RESULT with register address
func NewSgResult() *SgResult {
	return &SgResult{
		RegisterAddr: SG_RESULT,
	}
}

// Pack the individual fields into the Bytes field (a single 32-bit value).
func (sgResult *SgResult) Pack() uint32 {
	sgResult.Bytes = sgResult.Result & 0x3FF // Keep only the lower 10 bits
	return sgResult.Bytes
}

// Unpack the Bytes field into the individual fields.
func (sgResult *SgResult) Unpack(uint32) {
	sgResult.Result = sgResult.Bytes & 0x3FF
}
func (sgResult *SgResult) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, sgResult.RegisterAddr)
}
func (sgResult *SgResult) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, sgResult.RegisterAddr, driverIndex, value)
}

// CoolConf represents the fields in the TMC2209 COOLCONF register.
//
// The COOLCONF register controls various settings related to the CoolStep feature,
// which automatically adjusts the motor's current based on the load to optimize power consumption
// and minimize heat generation.
//
// - **SEMIN** (5 bits): The minimum current value for the CoolStep algorithm to be enabled. If the
// motor current falls below this value, CoolStep will reduce the current. The SEMIN field helps to
// control the minimum threshold for current scaling.
//
// - **SEUP** (2 bits): The step-up value for the current when CoolStep detects an increase in load.
// It defines the amount by which the motor current is increased when the load increases and the motor
// is at risk of stalling. It helps to balance current efficiency and motor performance.
//
// - **SEMAX** (5 bits): The maximum current value for CoolStep. This value sets the upper threshold
// for the current when the motor is under heavy load. It ensures that the motor can handle the load
// by increasing the current when necessary, while still maintaining efficiency.
//
// - **SEDN** (2 bits): The step-down value for the current when CoolStep detects a decrease in load.
// It helps to lower the current consumption when the motor is no longer under heavy load, optimizing
// energy efficiency.
//
// - **SEIMIN** (1 bit): Enables or disables the current scaling for low loads. If enabled, the motor
// current will be reduced under low load conditions, improving energy efficiency. If disabled, the
// motor current remains constant at the preset level, regardless of the load.
//
// The COOLCONF register allows for fine-tuning of the motor current scaling behavior based on the load,
// helping to optimize motor efficiency and reduce power consumption and heat generation.
type CoolConf struct {
	Semin          uint32 // 1 bit
	Sedn           uint32 // 2 bits (sedn0, sedn1)
	Semax          uint32 // 4 bits (semax0 to semax3)
	Seup           uint32 // 3 bits (seup0, seup1, seup2)
	Semin2         uint32 // 6 bits (semin0 to semin5)
	CoolStepEnable uint32 // 1 bit
	Reserved       uint32 // Reserved 10 bits
	Bytes          uint32 // The packed 32-bit value
	RegisterAddr   uint8  // The register address (COOLCONF)
}

func (coolConf *CoolConf) GetAddress() uint8 {
	//TODO implement me
	panic("implement me")
}

// Initialize COOLCONF with register address
func NewCoolConf() *CoolConf {
	return &CoolConf{
		RegisterAddr: COOLCONF,
	}
}

// Pack the individual fields into the Bytes field (a single 32-bit value).
func (coolConf *CoolConf) Pack() uint32 {
	coolConf.Bytes = (coolConf.Semin & 0x01) |
		((coolConf.Sedn & 0x03) << 1) |
		((coolConf.Semax & 0x0F) << 3) |
		((coolConf.Seup & 0x07) << 7) |
		((coolConf.Semin2 & 0x3F) << 10) |
		((coolConf.CoolStepEnable & 0x01) << 16) |
		((coolConf.Reserved & 0x3FF) << 17) // Reserve 10 bits for reserved fields
	return coolConf.Bytes
}

// Unpack the Bytes field into the individual fields.
func (coolConf *CoolConf) Unpack(uint32) {
	coolConf.Semin = coolConf.Bytes & 0x01
	coolConf.Sedn = (coolConf.Bytes >> 1) & 0x03
	coolConf.Semax = (coolConf.Bytes >> 3) & 0x0F
	coolConf.Seup = (coolConf.Bytes >> 7) & 0x07
	coolConf.Semin2 = (coolConf.Bytes >> 10) & 0x3F
	coolConf.CoolStepEnable = (coolConf.Bytes >> 16) & 0x01
	coolConf.Reserved = (coolConf.Bytes >> 17) & 0x3FF
}
func (coolConf *CoolConf) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, coolConf.RegisterAddr)
}
func (coolConf *CoolConf) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, coolConf.RegisterAddr, driverIndex, value)
}

// DrvStatus represents the fields in the TMC2209 DRV_STATUS register.
//
// The DRV_STATUS register provides information about the driver’s status,
// including over-temperature warnings, short-circuit detection, and current status.
//
// - **OTPW** (1 bit): Overtemperature warning. If this bit is set to 1, it indicates that the
// temperature of the driver has exceeded the safe operating threshold. This can be used to
// trigger protective actions, such as reducing the current or stopping the motor to prevent damage.
//
// - **OT** (1 bit): Overtemperature fault. If this bit is set to 1, it indicates that the driver
// has overheated and is currently in a fault condition due to excessive temperature. It typically
// results in disabling the motor or reducing motor power until the temperature drops to a safe level.
//
// - **S2GA** (1 bit): Short-to-Ground on A phase (phase A of the motor). This bit is set to 1
// if a short-circuit is detected between the motor's phase A and ground. This protection feature
// helps to safeguard the driver and motor from short circuits that could lead to damage.
//
// - **S2GB** (1 bit): Short-to-Ground on B phase (phase B of the motor). This bit is set to 1
// if a short-circuit is detected between the motor's phase B and ground. Similar to the previous
// bit, this provides protection against short-circuits for the B phase.
//
// - **S2VSA** (1 bit): Short-to-VCC on A phase (phase A of the motor). If this bit is set to 1,
// it indicates a short-circuit between phase A of the motor and the power supply voltage (VCC).
// This condition could be hazardous and requires corrective actions.
//
// - **S2VSB** (1 bit): Short-to-VCC on B phase (phase B of the motor). Similar to S2VSA, this
// bit is set when a short-circuit is detected between phase B of the motor and VCC.
//
// - **OLA** (1 bit): Overcurrent fault on A phase. This bit is set to 1 if the current through
// phase A exceeds the configured threshold, indicating an overcurrent condition. It may indicate
// a problem with the motor or wiring that requires attention.
//
// - **OLB** (1 bit): Overcurrent fault on B phase. Similar to OLA, this bit indicates an overcurrent
// fault in phase B, signaling an issue that needs to be addressed to prevent damage to the system.
//
// - **T120** (1 bit): Timeout fault for phase A (motor A). If this bit is set, it indicates that
// the motor did not receive any signal or is stuck for too long, possibly due to a failure in the
// motor or wiring. It indicates that the system has detected an abnormal motor condition.
//
// - **T143** (1 bit): Timeout fault for phase B (motor B). Similar to T120, this bit indicates
// an issue with phase B, such as the motor not receiving a signal within the expected timeframe.
//
// - **T150** (1 bit): Timeout fault for both phases A and B. This bit is set if both phases are
// experiencing a timeout condition, which could be due to an issue with the motor or control signals.
//
// - **T157** (1 bit): Timeout fault due to thermal shutdown. If this bit is set, it indicates that
// the driver has shut down due to an overtemperature condition, preventing further damage to the system.
//
// - **CS_ACTUAL** (5 bits): This field provides the actual current setting value for the motor.
// It is used to monitor the current being supplied to the motor in real-time, helping to detect any
// discrepancies or performance issues during operation.
//
// - **STEALTH** (1 bit): StealthChop status. When set to 1, this bit indicates that the StealthChop
// mode is active, which is a feature used to reduce motor noise and improve efficiency at low speeds.
// If set to 0, StealthChop is not active, and the driver may be using a more standard operating mode.
//
// - **STST** (1 bit): Step status. If set to 1, this bit indicates that the motor is currently
// moving or stepping. It can be used to detect if the motor is active or idle at any given time.
type DrvStatus struct {
	Stst         uint32 // Standstill indicator
	Stealth      uint32 // StealthChop indicator
	CsActual     uint32 // Actual motor current / smart energy current
	T157         uint32 // 157°C comparator
	T150         uint32 // 150°C comparator
	T143         uint32 // 143°C comparator
	T120         uint32 // 120°C comparator
	Olb          uint32 // Open load indicator phase B
	Ola          uint32 // Open load indicator phase A
	S2vsb        uint32 // Low-side short indicator phase B
	S2vsa        uint32 // Low-side short indicator phase A
	S2gb         uint32 // Short to ground indicator phase B
	S2ga         uint32 // Short to ground indicator phase A
	Ot           uint32 // Overtemperature flag
	Otpw         uint32 // Overtemperature pre-warning flag
	Reserved     uint32 // Reserved bits
	Bytes        uint32 // The packed 32-bit value
	RegisterAddr uint8  // Register address
}

func (drvStatus *DrvStatus) GetAddress() uint8 {
	return drvStatus.RegisterAddr
}

// Initialize DRV_STATUS with register address
func NewDrvStatus() *DrvStatus {
	return &DrvStatus{
		RegisterAddr: DRV_STATUS,
	}
}

// Pack the individual fields into the Bytes field (a single 32-bit value).
func (drvStatus *DrvStatus) Pack() uint32 {
	drvStatus.Bytes = (drvStatus.Stst & 0x01) |
		((drvStatus.Stealth & 0x01) << 1) |
		((drvStatus.CsActual & 0xFFFF) << 2) | // Actual current in bits 16-31
		((drvStatus.T157 & 0x01) << 18) |
		((drvStatus.T150 & 0x01) << 19) |
		((drvStatus.T143 & 0x01) << 20) |
		((drvStatus.T120 & 0x01) << 21) |
		((drvStatus.Olb & 0x01) << 22) |
		((drvStatus.Ola & 0x01) << 23) |
		((drvStatus.S2vsb & 0x01) << 24) |
		((drvStatus.S2vsa & 0x01) << 25) |
		((drvStatus.S2gb & 0x01) << 26) |
		((drvStatus.S2ga & 0x01) << 27) |
		((drvStatus.Ot & 0x01) << 28) |
		((drvStatus.Otpw & 0x01) << 29) |
		((drvStatus.Reserved & 0x7FF) << 30) // Reserved bits
	return drvStatus.Bytes
}

// Unpack the Bytes field into the individual fields.
func (drvStatus *DrvStatus) Unpack(uint32) {
	drvStatus.Stst = drvStatus.Bytes & 0x01
	drvStatus.Stealth = (drvStatus.Bytes >> 1) & 0x01
	drvStatus.CsActual = (drvStatus.Bytes >> 2) & 0xFFFF
	drvStatus.T157 = (drvStatus.Bytes >> 18) & 0x01
	drvStatus.T150 = (drvStatus.Bytes >> 19) & 0x01
	drvStatus.T143 = (drvStatus.Bytes >> 20) & 0x01
	drvStatus.T120 = (drvStatus.Bytes >> 21) & 0x01
	drvStatus.Olb = (drvStatus.Bytes >> 22) & 0x01
	drvStatus.Ola = (drvStatus.Bytes >> 23) & 0x01
	drvStatus.S2vsb = (drvStatus.Bytes >> 24) & 0x01
	drvStatus.S2vsa = (drvStatus.Bytes >> 25) & 0x01
	drvStatus.S2gb = (drvStatus.Bytes >> 26) & 0x01
	drvStatus.S2ga = (drvStatus.Bytes >> 27) & 0x01
	drvStatus.Ot = (drvStatus.Bytes >> 28) & 0x01
	drvStatus.Otpw = (drvStatus.Bytes >> 29) & 0x01
	drvStatus.Reserved = (drvStatus.Bytes >> 30) & 0x7FF
}
func (drvStatus *DrvStatus) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, drvStatus.RegisterAddr)
}
func (drvStatus *DrvStatus) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, drvStatus.RegisterAddr, driverIndex, value)
}

// PwmScale represents the fields in the TMC2209 PWM_SCALE register.
//
// The PWM_SCALE register provides information related to the current PWM scaling
// values for the motor driver. It is used to determine the effective current
// supplied to the motor and the PWM (Pulse Width Modulation) duty cycle.
//
// - **PwmScaleSum** (8 bits): This field provides the total PWM scaling value used
// for the motor driver. It represents the summed scaling value of both motor phases.
// A higher value typically means more power is delivered to the motor.
//
// - **PwmScaleAuto** (9 bits): This field contains the automatically calculated
// PWM scaling value. It adjusts the power delivery dynamically based on motor
// load and thermal conditions to optimize performance and efficiency.
// The value in this field provides an indication of the current level of scaling
// that the driver is using in the automatic mode, based on the real-time motor conditions.
type PwmScale struct {
	PwmScaleSum  uint32 // 8-bit PWM duty cycle
	PwmScaleAuto int32  // 9-bit signed offset (-255 to +255)
	Bytes        uint32 // The packed 32-bit value
	RegisterAddr uint8  // Register address
}

func (pwm *PwmScale) GetAddress() uint8 {
	return pwm.RegisterAddr
}

// NewPwmScale Initialize PwmScale with register address
func NewPwmScale() *PwmScale {
	return &PwmScale{
		RegisterAddr: PWM_SCALE, // PWM_SCALE register address
	}
}

// Pack the individual fields into the Bytes field (a single 32-bit value).
func (pwm *PwmScale) Pack() uint32 {
	pwm.Bytes = (pwm.PwmScaleSum & 0xFF) |
		((uint32(pwm.PwmScaleAuto) & 0x1FF) << 8) // 9 bits for PWM_SCALE_AUTO
	return pwm.Bytes
}

// Unpack the Bytes field into the individual fields.
func (pwm *PwmScale) Unpack(uint32) {
	pwm.PwmScaleSum = pwm.Bytes & 0xFF
	pwm.PwmScaleAuto = int32((pwm.Bytes >> 8) & 0x1FF) // 9-bit signed value
}
func (pwm *PwmScale) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, pwm.RegisterAddr)
}
func (pwm *PwmScale) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, pwm.RegisterAddr, driverIndex, value)
}

// PwmAuto represents the fields in the TMC2209 PWM_AUTO register.
//
// The PWM_AUTO register is used to configure and monitor the automatic scaling and
// adjustment of PWM settings in the TMC2209 motor driver. This register helps to
// fine-tune motor control behavior based on real-time operating conditions.
//
// - **PwmOfsAuto** (8 bits): This field contains the automatically adjusted PWM
// offset value. The PWM offset is used to set the initial duty cycle for the PWM
// signal. Adjusting this field allows for fine-tuning of motor behavior in
// specific scenarios, such as optimizing torque or minimizing motor heating.
//
// - **PwmGradAuto** (8 bits): This field holds the automatically adjusted PWM
// gradient value. The gradient determines the rate at which the PWM duty cycle
// increases or decreases over time. A higher gradient value can lead to faster
// transitions in power delivery, which is useful for certain motor acceleration
// profiles or applications requiring smooth power changes.
type PwmAuto struct {
	PwmOfsAuto   int32  // 8-bit signed offset value (-255 to +255)
	PwmGradAuto  int32  // 8-bit automatically determined gradient value (-255 to +255)
	Bytes        uint32 // The packed 32-bit value
	RegisterAddr uint8  // Register address
}

func (pwm *PwmAuto) GetAddress() uint8 {
	return pwm.RegisterAddr
}

// Initialize PwmAuto with register address
func NewPwmAuto() *PwmAuto {
	return &PwmAuto{
		RegisterAddr: PWM_AUTO, // PWM_AUTO register address
	}
}

// Pack the individual fields into the Bytes field (a single 32-bit value).
func (pwm *PwmAuto) Pack() uint32 {
	pwm.Bytes = (uint32(pwm.PwmOfsAuto) & 0xFF) |
		((uint32(pwm.PwmGradAuto) & 0xFF) << 8) // 8 bits for each value
	return pwm.Bytes
}

// Unpack the Bytes field into the individual fields.
func (pwm *PwmAuto) Unpack(uint32) {
	pwm.PwmOfsAuto = int32(pwm.Bytes & 0xFF)
	pwm.PwmGradAuto = int32((pwm.Bytes >> 8) & 0xFF)
}
func (pwm *PwmAuto) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, pwm.RegisterAddr)
}
func (pwm *PwmAuto) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, pwm.RegisterAddr, driverIndex, value)
}

// Tpowerdown represents the fields in the TMC2209 TPOWERDOWN register.
//
// The TPOWERDOWN register is used to configure the time delay before the driver
// enters the power-down state after the motor has been idle. This is useful for
// energy-saving purposes and to reduce the overall power consumption when the
// motor is not being actively driven.
//
// - **Tpowerdown** (8 bits): This field specifies the time delay (in microseconds)
// before the TMC2209 enters the power-down mode. The power-down mode reduces the
// current drawn by the motor and motor driver, which helps save energy when the
// motor is not in use. The delay is adjustable to balance between responsiveness
// and power consumption.
type Tpowerdown struct {
	DelayTime    uint32 // Delay time from standstill detection to motor current power-down (8 bits)
	RegisterAddr uint8  // Register address
	Bytes        uint32 // The packed 32-bit value
}

func (tpd *Tpowerdown) GetAddress() uint8 {
	return tpd.RegisterAddr
}

// Initialize Tpowerdown with register address
func NewTpowerdown() *Tpowerdown {
	return &Tpowerdown{
		RegisterAddr: TPOWERDOWN, // TPOWERDOWN register address
	}
}

// Pack the DelayTime field into the Bytes field (a single 8-bit value).
func (tpd *Tpowerdown) Pack() uint32 {
	tpd.Bytes = tpd.DelayTime & 0xFF
	return tpd.Bytes
}

// Unpack the Bytes field into the DelayTime field.
func (tpd *Tpowerdown) Unpack(uint32) {
	tpd.DelayTime = tpd.Bytes & 0xFF
}
func (tpd *Tpowerdown) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, tpd.RegisterAddr)
}
func (tpd *Tpowerdown) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, tpd.RegisterAddr, driverIndex, value)
}

// Tstep represents the fields in the TMC2209 TSTEP register.
//
// The TSTEP register configures the inter-step duration, which is the time between
// the steps sent to the motor driver. This is important for controlling the speed
// of the motor, and is typically used to set the stepping frequency.
//
// - **Tstep** (24 bits): This field specifies the duration of time between each step
// signal, measured in microseconds. The value determines how quickly the motor steps,
// effectively controlling the motor's speed. Smaller values result in faster stepping
// (higher speed), while larger values reduce the stepping frequency (lower speed).
//
// The `TSTEP` register is critical for applications that require precise motor control,
// especially for variable-speed applications. The duration set in `TSTEP` should be
// chosen based on the desired motor speed and the capabilities of the motor driver.
type Tstep struct {
	StepTime     uint32 // Time between 1/256 microsteps (20 bits)
	RegisterAddr uint8  // Register address
	Bytes        uint32 // The packed 32-bit value
}

func (tstep *Tstep) GetAddress() uint8 {
	return tstep.RegisterAddr
}

// Initialize Tstep with register address
func NewTstep() *Tstep {
	return &Tstep{
		RegisterAddr: TSTEP, // TSTEP register address
	}
}

// Pack the StepTime field into the Bytes field (a single 20-bit value).
func (tstep *Tstep) Pack() uint32 {
	tstep.Bytes = tstep.StepTime & 0xFFFFF // 20 bits for TSTEP
	return tstep.Bytes
}

// Unpack the Bytes field into the StepTime field.
func (tstep *Tstep) Unpack(uint32) {
	tstep.StepTime = tstep.Bytes & 0xFFFFF
}
func (tstep *Tstep) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, tstep.RegisterAddr)
}
func (tstep *Tstep) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, tstep.RegisterAddr, driverIndex, value)
}

// Mscnt  represents the Microstep Counter Register (0x6A) in the TMC2209
//
// This register provides the actual microstep position within the microstep table.
// The value of MSCNT allows determination of the motor position within the electrical wave,
// which is essential for accurately controlling the motor's position during operation.
//
// Fields:
// - The register contains a 10-bit value indicating the actual position in the microstep table.
// - Range: 0 to 1023 (0x000 to 0x3FF)
type Mscnt struct {
	Position     uint32
	RegisterAddr uint8
}

func (mscnt *Mscnt) GetAddress() uint8 {
	return mscnt.RegisterAddr
}

// Pack the Position value (10-bit) into the Bytes field
func (mscnt *Mscnt) Pack() uint32 {
	mscnt.Position = mscnt.Position & 0x03FF // Limit to 10 bits (0x3FF)
	return mscnt.Position
}

// Unpack the Bytes field into the Position field.
func (mscnt *Mscnt) Unpack(uint32) {
	mscnt.Position = mscnt.Position & 0x03FF
}

// NewMscnt initializes a new Mscnt struct with the correct register address.
func NewMscnt() *Mscnt {
	return &Mscnt{
		RegisterAddr: MSCNT, // MSCNT register address
	}
}
func (mscnt *Mscnt) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, mscnt.RegisterAddr)
}
func (mscnt *Mscnt) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, mscnt.RegisterAddr, driverIndex, value)
}

// Mscuract MSCURACT represents the Microstep Current Register (0x6B) in the TMC2209
//
// This register provides the actual current values for motor phases A and B as read from
// the internal sine wave table. The values are not scaled by the current setting but provide
// a raw value for the motor's electrical waves.
//
// Fields:
// - CUR_B (bits 0-7): The actual microstep current for motor phase B (signed value in the range +/-255).
// - CUR_A (bits 24-16): The actual microstep current for motor phase A (signed value in the range +/-255).
type Mscuract struct {
	// Microstep current for phase B (sine wave)
	CurB uint32
	// Microstep current for phase A (cosine wave)
	CurA uint32
	// Register address
	RegisterAddr uint8
}

func (mscuract *Mscuract) GetAddress() uint8 {
	return mscuract.RegisterAddr
}

// Pack the individual fields into the Bytes field (a single 32-bit value).
func (mscuract *Mscuract) Pack() uint32 {
	mscuract.CurB = mscuract.CurB & 0xFF // Limit to 8 bits for CUR_B
	mscuract.CurA = mscuract.CurA & 0xFF // Limit to 8 bits for CUR_A
	return mscuract.CurA
}

// Unpack the Bytes field into the individual fields.
func (mscuract *Mscuract) Unpack(uint32) {
	mscuract.CurB = mscuract.CurB & 0xFF         // Extract CUR_B (8 bits)
	mscuract.CurA = (mscuract.CurA >> 16) & 0xFF // Extract CUR_A (8 bits)
}

// NewMscuract initializes a new Mscuract struct with the correct register address.
func NewMscuract() *Mscuract {
	return &Mscuract{
		RegisterAddr: MSCURACT,
	}
}
func (mscuract *Mscuract) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, mscuract.RegisterAddr)
}
func (mscuract *Mscuract) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, mscuract.RegisterAddr, driverIndex, value)
}
