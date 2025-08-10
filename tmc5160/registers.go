package tmc5160

import (
	math "github.com/orsinium-labs/tinymath"
)

// RegisterComm defines an interface for reading from and writing to hardware registers.
type RegisterComm interface {
	ReadRegister(register uint8, driverIndex uint8) (uint32, error)
	WriteRegister(register uint8, value uint32, driverIndex uint8) error
}

// ReadRegister function using the register constants
func ReadRegister(comm RegisterComm, driverIndex uint8, register uint8) (uint32, error) {
	// Read the register value using the comm interface

	value, err := comm.ReadRegister(register, driverIndex)
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

// Register and methods to pack and unpack
// Base Register struct
type Register struct {
	RegisterAddr uint8
	Bytes        uint32
}

// Common New function for creating a new register instance
func NewRegister(addr uint8) *Register {
	return &Register{
		RegisterAddr: addr,
	}
}

// Common Pack method: for subclasses to implement their packing logic
func (r *Register) Pack() uint32 {
	return r.Bytes // Default, should be overridden in register-specific structs
}

// Common Unpack method: for subclasses to implement their unpacking logic
func (r *Register) Unpack(registerValue uint32) {
	r.Bytes = registerValue // Default, should be overridden in register-specific structs
}

// Common GetAddress method
func (r *Register) GetAddress() uint8 {
	return r.RegisterAddr
}

// Common Read method (assuming the communication interface is implemented)
func (r *Register) Read(comm RegisterComm, driverIndex uint8) (uint32, error) {
	return ReadRegister(comm, driverIndex, r.RegisterAddr)
}

// Common Write method
func (r *Register) Write(comm RegisterComm, driverIndex uint8, value uint32) error {
	return WriteRegister(comm, r.RegisterAddr, driverIndex, value)
}

// GCONF Register bit fields' masks and shifts
const (
	// Recalibrate: Zero crossing recalibration during driver disable
	GCONF_Recalibrate_Mask = 1 << 0
	// Faststandstill: Timeout for step execution until standstill detection
	GCONF_Faststandstill_Mask = 1 << 1
	// Enable PWM mode for StealthChop
	GCONF_EnPwmMode_Mask = 1 << 2
	// Enable step input filtering for StealthChop optimization
	GCONF_MultistepFilt_Mask = 1 << 3
	// Motor direction
	GCONF_Shaft_Mask = 1 << 4
	// Error flags on DIAG0 pin
	GCONF_Diag0Error_Mask = 1 << 5
	// Enable DIAG0 for Over temperature warning
	GCONF_Diag0Otpw_Mask = 1 << 6
	// Enable DIAG0 for stall step detection
	GCONF_Diag0StallStep_Mask = 1 << 7
	// Enable DIAG1 for stall direction
	GCONF_Diag1StallDir_Mask = 1 << 8
	// Enable DIAG1 for index position
	GCONF_Diag1Index_Mask = 1 << 9
	// Enable DIAG1 for chopper on state
	GCONF_Diag1Onstate_Mask = 1 << 10
	// Enable DIAG1 for skipped steps
	GCONF_Diag1StepsSkipped_Mask = 1 << 11
	// Enable DIAG0 push-pull output
	GCONF_Diag0IntPushPull_Mask = 1 << 12
	// Enable DIAG1 push-pull output
	GCONF_Diag1PosCompPushPull_Mask = 1 << 13
	// Small hysteresis for step frequency comparison
	GCONF_SmallHysteresis_Mask = 1 << 14
	// Enable emergency stop
	GCONF_StopEnable_Mask = 1 << 15
	// Direct motor coil current and polarity control
	GCONF_DirectMode_Mask = 1 << 16
	// Test mode (not for normal use)
	GCONF_TestMode_Mask = 1 << 17
)

// GCONF Register structure
type GCONF_Register struct {
	Register
	// Fields corresponding to individual settings in GCONF register
	Recalibrate          bool
	Faststandstill       bool
	EnPwmMode            bool
	MultistepFilt        bool
	Shaft                bool
	Diag0Error           bool
	Diag0Otpw            bool
	Diag0StallStep       bool
	Diag1StallDir        bool
	Diag1Index           bool
	Diag1Onstate         bool
	Diag1StepsSkipped    bool
	Diag0IntPushPull     bool
	Diag1PosCompPushPull bool
	SmallHysteresis      bool
	StopEnable           bool
	DirectMode           bool
	TestMode             bool
}

// NewGCONF initializes a new GCONF register instance
func NewGCONF() *GCONF_Register {
	return &GCONF_Register{
		Register: Register{
			RegisterAddr: GCONF, // GSTAT register address
		},
	}
}

// Pack the fields into a single 32-bit register value
func (g *GCONF_Register) Pack() uint32 {
	var registerValue uint32

	// Use bitwise OR to set individual bits based on the field values
	if g.Recalibrate {
		registerValue |= GCONF_Recalibrate_Mask
	}
	if g.Faststandstill {
		registerValue |= GCONF_Faststandstill_Mask
	}
	if g.EnPwmMode {
		registerValue |= GCONF_EnPwmMode_Mask
	}
	if g.MultistepFilt {
		registerValue |= GCONF_MultistepFilt_Mask
	}
	if g.Shaft {
		registerValue |= GCONF_Shaft_Mask
	}
	if g.Diag0Error {
		registerValue |= GCONF_Diag0Error_Mask
	}
	if g.Diag0Otpw {
		registerValue |= GCONF_Diag0Otpw_Mask
	}
	if g.Diag0StallStep {
		registerValue |= GCONF_Diag0StallStep_Mask
	}
	if g.Diag1StallDir {
		registerValue |= GCONF_Diag1StallDir_Mask
	}
	if g.Diag1Index {
		registerValue |= GCONF_Diag1Index_Mask
	}
	if g.Diag1Onstate {
		registerValue |= GCONF_Diag1Onstate_Mask
	}
	if g.Diag1StepsSkipped {
		registerValue |= GCONF_Diag1StepsSkipped_Mask
	}
	if g.Diag0IntPushPull {
		registerValue |= GCONF_Diag0IntPushPull_Mask
	}
	if g.Diag1PosCompPushPull {
		registerValue |= GCONF_Diag1PosCompPushPull_Mask
	}
	if g.SmallHysteresis {
		registerValue |= GCONF_SmallHysteresis_Mask
	}
	if g.StopEnable {
		registerValue |= GCONF_StopEnable_Mask
	}
	if g.DirectMode {
		registerValue |= GCONF_DirectMode_Mask
	}
	if g.TestMode {
		registerValue |= GCONF_TestMode_Mask
	}
	return registerValue
}

// Unpack a 32-bit register value into individual fields
func (g *GCONF_Register) Unpack(registerValue uint32) {
	g.Recalibrate = (registerValue & GCONF_Recalibrate_Mask) != 0
	g.Faststandstill = (registerValue & GCONF_Faststandstill_Mask) != 0
	g.EnPwmMode = (registerValue & GCONF_EnPwmMode_Mask) != 0
	g.MultistepFilt = (registerValue & GCONF_MultistepFilt_Mask) != 0
	g.Shaft = (registerValue & GCONF_Shaft_Mask) != 0
	g.Diag0Error = (registerValue & GCONF_Diag0Error_Mask) != 0
	g.Diag0Otpw = (registerValue & GCONF_Diag0Otpw_Mask) != 0
	g.Diag0StallStep = (registerValue & GCONF_Diag0StallStep_Mask) != 0
	g.Diag1StallDir = (registerValue & GCONF_Diag1StallDir_Mask) != 0
	g.Diag1Index = (registerValue & GCONF_Diag1Index_Mask) != 0
	g.Diag1Onstate = (registerValue & GCONF_Diag1Onstate_Mask) != 0
	g.Diag1StepsSkipped = (registerValue & GCONF_Diag1StepsSkipped_Mask) != 0
	g.Diag0IntPushPull = (registerValue & GCONF_Diag0IntPushPull_Mask) != 0
	g.Diag1PosCompPushPull = (registerValue & GCONF_Diag1PosCompPushPull_Mask) != 0
	g.SmallHysteresis = (registerValue & GCONF_SmallHysteresis_Mask) != 0
	g.StopEnable = (registerValue & GCONF_StopEnable_Mask) != 0
	g.DirectMode = (registerValue & GCONF_DirectMode_Mask) != 0
	g.TestMode = (registerValue & GCONF_TestMode_Mask) != 0
}

// Example Register: GSTAT
type GSTAT_Register struct {
	Register
	Reset  bool
	DrvErr bool
	UvCp   bool
}

// NewGSTAT creates a new GSTAT register instance
func NewGSTAT() *GSTAT_Register {
	return &GSTAT_Register{
		Register: Register{
			RegisterAddr: GSTAT, // GSTAT register address
		},
	}
}

// Pack method for GSTAT: overrides the base Pack
func (g *GSTAT_Register) Pack() uint32 {
	var registerValue uint32
	if g.Reset {
		registerValue |= 1 << 0
	}
	if g.DrvErr {
		registerValue |= 1 << 1
	}
	if g.UvCp {
		registerValue |= 1 << 2
	}
	return registerValue
}

// Unpack method for GSTAT: overrides the base Unpack
func (g *GSTAT_Register) Unpack(registerValue uint32) {
	g.Reset = (registerValue & (1 << 0)) != 0
	g.DrvErr = (registerValue & (1 << 1)) != 0
	g.UvCp = (registerValue & (1 << 2)) != 0
}

// IOIN_Register struct to represent the IOIN register
type IOIN_Register struct {
	Register
	ReflStep     bool
	RefrDir      bool
	EncbDcenCfg4 bool
	EncaDcinCfg5 bool
	DrvEnn       bool
	EncNDcoCfg6  bool
	SdMode       bool
	SwcompIn     bool
	Version      uint8
}

// NewIOIN creates a new IOIN register instance
func NewIOIN() *IOIN_Register {
	return &IOIN_Register{
		Register: Register{
			RegisterAddr: IOIN,
		},
	}
}

// Pack method for IOIN: overrides the base Pack
func (i *IOIN_Register) Pack() uint32 {
	var registerValue uint32

	// Set individual bits based on the field values
	if i.ReflStep {
		registerValue |= 1 << 0
	}
	if i.RefrDir {
		registerValue |= 1 << 1
	}
	if i.EncbDcenCfg4 {
		registerValue |= 1 << 2
	}
	if i.EncaDcinCfg5 {
		registerValue |= 1 << 3
	}
	if i.DrvEnn {
		registerValue |= 1 << 4
	}
	if i.EncNDcoCfg6 {
		registerValue |= 1 << 5
	}
	if i.SdMode {
		registerValue |= 1 << 6
	}
	if i.SwcompIn {
		registerValue |= 1 << 7
	}
	// Handle the version field (8 bits, starting at bit 24)
	registerValue |= uint32(i.Version) << 24

	return registerValue
}

// Unpack method for IOIN: overrides the base Unpack
func (i *IOIN_Register) Unpack(registerValue uint32) {
	i.ReflStep = (registerValue & (1 << 0)) != 0
	i.RefrDir = (registerValue & (1 << 1)) != 0
	i.EncbDcenCfg4 = (registerValue & (1 << 2)) != 0
	i.EncaDcinCfg5 = (registerValue & (1 << 3)) != 0
	i.DrvEnn = (registerValue & (1 << 4)) != 0
	i.EncNDcoCfg6 = (registerValue & (1 << 5)) != 0
	i.SdMode = (registerValue & (1 << 6)) != 0
	i.SwcompIn = (registerValue & (1 << 7)) != 0
	// Extract the version field (8 bits, starting at bit 24)
	i.Version = uint8((registerValue >> 24) & 0xFF)
}

// SHORT_CONF_Register struct to represent the SHORT_CONF register
type SHORT_CONF_Register struct {
	Register
	S2vsLevel   uint8 // Short to VS detector sensitivity (4 bits)
	S2gLevel    uint8 // Short to GND detector sensitivity (4 bits)
	ShortFilter uint8 // Spike filtering bandwidth for short detection (2 bits)
	ShortDelay  bool  // Short detection delay (1 bit)
}

// NewSHORT_CONF creates a new SHORT_CONF register instance
func NewSHORT_CONF() *SHORT_CONF_Register {
	return &SHORT_CONF_Register{
		Register: Register{
			RegisterAddr: SHORT_CONF,
		},
	}
}

// Pack method for SHORT_CONF: overrides the base Pack
func (s *SHORT_CONF_Register) Pack() uint32 {
	var registerValue uint32

	// Pack each field using bitwise operations
	registerValue |= uint32(s.S2vsLevel&0xF) << 0    // S2vsLevel: 4 bits
	registerValue |= uint32(s.S2gLevel&0xF) << 8     // S2gLevel: 4 bits
	registerValue |= uint32(s.ShortFilter&0x3) << 16 // ShortFilter: 2 bits
	if s.ShortDelay {
		registerValue |= 1 << 18 // ShortDelay: 1 bit
	}
	return registerValue
}

// Unpack method for SHORT_CONF: overrides the base Unpack
func (s *SHORT_CONF_Register) Unpack(registerValue uint32) {
	// Unpack each field using bitwise operations
	s.S2vsLevel = uint8((registerValue >> 0) & 0xF)    // Extract 4 bits for S2vsLevel
	s.S2gLevel = uint8((registerValue >> 8) & 0xF)     // Extract 4 bits for S2gLevel
	s.ShortFilter = uint8((registerValue >> 16) & 0x3) // Extract 2 bits for ShortFilter
	s.ShortDelay = (registerValue & (1 << 18)) != 0    // Extract 1 bit for ShortDelay
}

// DRV_CONF_Register struct to represent the DRV_CONF register
type DRV_CONF_Register struct {
	Register
	BBMTime     uint8 // Break before make delay (5 bits)
	BBMClks     uint8 // Digital BBM Time in clock cycles (4 bits)
	OTSelect    uint8 // Over temperature level selection for bridge disable (2 bits)
	DrvStrength uint8 // Gate drivers current selection (2 bits)
	FiltIsense  uint8 // Filter time constant of sense amplifier (2 bits)
}

// NewDRV_CONF creates a new DRV_CONF register instance
func NewDRV_CONF() *DRV_CONF_Register {
	return &DRV_CONF_Register{
		Register: Register{
			RegisterAddr: DRV_CONF,
		},
	}
}

// Pack method for DRV_CONF: overrides the base Pack
func (d *DRV_CONF_Register) Pack() uint32 {
	var registerValue uint32

	// Pack each field using bitwise operations
	registerValue |= uint32(d.BBMTime&0x1F) << 0     // BBMTime: 5 bits
	registerValue |= uint32(d.BBMClks&0xF) << 8      // BBMClks: 4 bits
	registerValue |= uint32(d.OTSelect&0x3) << 16    // OTSelect: 2 bits
	registerValue |= uint32(d.DrvStrength&0x3) << 18 // DrvStrength: 2 bits
	registerValue |= uint32(d.FiltIsense&0x3) << 20  // FiltIsense: 2 bits

	return registerValue
}

// Unpack method for DRV_CONF: overrides the base Unpack
func (d *DRV_CONF_Register) Unpack(registerValue uint32) {
	// Unpack each field using bitwise operations
	d.BBMTime = uint8((registerValue >> 0) & 0x1F)     // Extract 5 bits for BBMTime
	d.BBMClks = uint8((registerValue >> 8) & 0xF)      // Extract 4 bits for BBMClks
	d.OTSelect = uint8((registerValue >> 16) & 0x3)    // Extract 2 bits for OTSelect
	d.DrvStrength = uint8((registerValue >> 18) & 0x3) // Extract 2 bits for DrvStrength
	d.FiltIsense = uint8((registerValue >> 20) & 0x3)  // Extract 2 bits for FiltIsense
}

// OFFSET_READ_Register struct to represent the OFFSET_READ register
type OFFSET_READ_Register struct {
	Register
	PhaseB uint8 // Phase B offset (8 bits)
	PhaseA uint8 // Phase A offset (8 bits)
}

// NewOFFSET_READ creates a new OFFSET_READ register instance
func NewOFFSET_READ() *OFFSET_READ_Register {
	return &OFFSET_READ_Register{
		Register: Register{
			RegisterAddr: OFFSET_READ,
		},
	}
}

// Pack method for OFFSET_READ: overrides the base Pack
func (o *OFFSET_READ_Register) Pack() uint32 {
	var registerValue uint32

	// Pack each field using bitwise operations
	registerValue |= uint32(o.PhaseB&0xFF) << 0 // PhaseB: 8 bits
	registerValue |= uint32(o.PhaseA&0xFF) << 8 // PhaseA: 8 bits

	return registerValue
}

// Unpack method for OFFSET_READ: overrides the base Unpack
func (o *OFFSET_READ_Register) Unpack(registerValue uint32) {
	// Unpack each field using bitwise operations
	o.PhaseB = uint8((registerValue >> 0) & 0xFF) // Extract 8 bits for PhaseB
	o.PhaseA = uint8((registerValue >> 8) & 0xFF) // Extract 8 bits for PhaseA
}

// IHOLD_IRUN_Register struct to represent the IHOLD_IRUN register
type IHOLD_IRUN_Register struct {
	Register
	Ihold      uint8 // Standstill current (5 bits)
	Irun       uint8 // Motor run current (5 bits)
	IholdDelay uint8 // Motor power down delay (4 bits)
}

// NewIHOLD_IRUN creates a new IHOLD_IRUN register instance
func NewIHOLD_IRUN() *IHOLD_IRUN_Register {
	return &IHOLD_IRUN_Register{
		Register: Register{
			RegisterAddr: IHOLD_IRUN,
		},
	}
}

// Pack method for IHOLD_IRUN: overrides the base Pack
func (i *IHOLD_IRUN_Register) Pack() uint32 {
	var registerValue uint32

	// Pack each field using bitwise operations
	registerValue |= uint32(i.Ihold&0x1F) << 0      // Ihold: 5 bits
	registerValue |= uint32(i.Irun&0x1F) << 8       // Irun: 5 bits
	registerValue |= uint32(i.IholdDelay&0xF) << 16 // IholdDelay: 4 bits

	return registerValue
}

// Unpack method for IHOLD_IRUN: overrides the base Unpack
func (i *IHOLD_IRUN_Register) Unpack(registerValue uint32) {
	// Unpack each field using bitwise operations
	i.Ihold = uint8((registerValue >> 0) & 0x1F)      // Extract 5 bits for Ihold
	i.Irun = uint8((registerValue >> 8) & 0x1F)       // Extract 5 bits for Irun
	i.IholdDelay = uint8((registerValue >> 16) & 0xF) // Extract 4 bits for IholdDelay
}

// SW_MODE_Register struct to represent the SW_MODE register
type SW_MODE_Register struct {
	Register
	StopLEnable    bool // Enable automatic motor stop during active left reference switch input
	StopREnable    bool // Enable automatic motor stop during active right reference switch input
	PolStopL       bool // Sets the active polarity of the left reference switch input
	PolStopR       bool // Sets the active polarity of the right reference switch input
	SwapLR         bool // Swap the left and right reference switch inputs
	LatchLActive   bool // Activate latching of the position to XLATCH upon an active going edge on REFL
	LatchLInactive bool // Activate latching of the position to XLATCH upon an inactive going edge on REFL
	LatchRActive   bool // Activate latching of the position to XLATCH upon an active going edge on REFR
	LatchRInactive bool // Activate latching of the position to XLATCH upon an inactive going edge on REFR
	EnLatchEncoder bool // Latch encoder position to ENC_LATCH upon reference switch event
	SgStop         bool // Enable stop by stallGuard2
	EnSoftStop     bool // Enable soft stop upon a stop event
}

// NewSW_MODE creates a new SW_MODE register instance
func NewSW_MODE() *SW_MODE_Register {
	return &SW_MODE_Register{
		Register: Register{
			RegisterAddr: SW_MODE,
		},
	}
}

// Pack method for SW_MODE: overrides the base Pack
func (s *SW_MODE_Register) Pack() uint32 {
	var registerValue uint32

	// Pack each field using bitwise operations
	if s.StopLEnable {
		registerValue |= 1 << 0
	}
	if s.StopREnable {
		registerValue |= 1 << 1
	}
	if s.PolStopL {
		registerValue |= 1 << 2
	}
	if s.PolStopR {
		registerValue |= 1 << 3
	}
	if s.SwapLR {
		registerValue |= 1 << 4
	}
	if s.LatchLActive {
		registerValue |= 1 << 5
	}
	if s.LatchLInactive {
		registerValue |= 1 << 6
	}
	if s.LatchRActive {
		registerValue |= 1 << 7
	}
	if s.LatchRInactive {
		registerValue |= 1 << 8
	}
	if s.EnLatchEncoder {
		registerValue |= 1 << 9
	}
	if s.SgStop {
		registerValue |= 1 << 10
	}
	if s.EnSoftStop {
		registerValue |= 1 << 11
	}

	return registerValue
}

// Unpack method for SW_MODE: overrides the base Unpack
func (s *SW_MODE_Register) Unpack(registerValue uint32) {
	// Unpack each field using bitwise operations
	s.StopLEnable = (registerValue & (1 << 0)) != 0
	s.StopREnable = (registerValue & (1 << 1)) != 0
	s.PolStopL = (registerValue & (1 << 2)) != 0
	s.PolStopR = (registerValue & (1 << 3)) != 0
	s.SwapLR = (registerValue & (1 << 4)) != 0
	s.LatchLActive = (registerValue & (1 << 5)) != 0
	s.LatchLInactive = (registerValue & (1 << 6)) != 0
	s.LatchRActive = (registerValue & (1 << 7)) != 0
	s.LatchRInactive = (registerValue & (1 << 8)) != 0
	s.EnLatchEncoder = (registerValue & (1 << 9)) != 0
	s.SgStop = (registerValue & (1 << 10)) != 0
	s.EnSoftStop = (registerValue & (1 << 11)) != 0
}

// RAMP_STAT_Register struct to represent the RAMP_STAT register
type RAMP_STAT_Register struct {
	Register
	StatusStopL     bool // Reference switch left status (1=active)
	StatusStopR     bool // Reference switch right status (1=active)
	StatusLatchL    bool // Latch left ready (enable position latching)
	StatusLatchR    bool // Latch right ready (enable position latching)
	EventStopL      bool // Active stop left condition due to stop switch
	EventStopR      bool // Active stop right condition due to stop switch
	EventStopSG     bool // Active StallGuard2 stop event
	EventPosReached bool // Target position reached
	VelocityReached bool // Target velocity reached
	PositionReached bool // Target position reached
	VZero           bool // Actual velocity is 0
	TZeroWaitActive bool // TZEROWAIT is active after motor stop
	SecondMove      bool // Automatic ramp required moving back in opposite direction
	StatusSG        bool // Active stallGuard2 input
}

// NewRAMP_STAT creates a new RAMP_STAT register instance
func NewRAMP_STAT() *RAMP_STAT_Register {
	return &RAMP_STAT_Register{
		Register: Register{
			RegisterAddr: RAMP_STAT,
		},
	}
}

// Pack method for RAMP_STAT: overrides the base Pack
func (r *RAMP_STAT_Register) Pack() uint32 {
	var registerValue uint32

	// Pack each field using bitwise operations
	if r.StatusStopL {
		registerValue |= 1 << 0
	}
	if r.StatusStopR {
		registerValue |= 1 << 1
	}
	if r.StatusLatchL {
		registerValue |= 1 << 2
	}
	if r.StatusLatchR {
		registerValue |= 1 << 3
	}
	if r.EventStopL {
		registerValue |= 1 << 4
	}
	if r.EventStopR {
		registerValue |= 1 << 5
	}
	if r.EventStopSG {
		registerValue |= 1 << 6
	}
	if r.EventPosReached {
		registerValue |= 1 << 7
	}
	if r.VelocityReached {
		registerValue |= 1 << 8
	}
	if r.PositionReached {
		registerValue |= 1 << 9
	}
	if r.VZero {
		registerValue |= 1 << 10
	}
	if r.TZeroWaitActive {
		registerValue |= 1 << 11
	}
	if r.SecondMove {
		registerValue |= 1 << 12
	}
	if r.StatusSG {
		registerValue |= 1 << 13
	}

	return registerValue
}

// Unpack method for RAMP_STAT: overrides the base Unpack
func (r *RAMP_STAT_Register) Unpack(registerValue uint32) {
	// Unpack each field using bitwise operations
	r.StatusStopL = (registerValue & (1 << 0)) != 0
	r.StatusStopR = (registerValue & (1 << 1)) != 0
	r.StatusLatchL = (registerValue & (1 << 2)) != 0
	r.StatusLatchR = (registerValue & (1 << 3)) != 0
	r.EventStopL = (registerValue & (1 << 4)) != 0
	r.EventStopR = (registerValue & (1 << 5)) != 0
	r.EventStopSG = (registerValue & (1 << 6)) != 0
	r.EventPosReached = (registerValue & (1 << 7)) != 0
	r.VelocityReached = (registerValue & (1 << 8)) != 0
	r.PositionReached = (registerValue & (1 << 9)) != 0
	r.VZero = (registerValue & (1 << 10)) != 0
	r.TZeroWaitActive = (registerValue & (1 << 11)) != 0
	r.SecondMove = (registerValue & (1 << 12)) != 0
	r.StatusSG = (registerValue & (1 << 13)) != 0
}

// ENCMODE_Register struct to represent the ENCMODE register
type ENCMODE_Register struct {
	Register
	PolA          bool  // Required A polarity for an N channel event
	PolB          bool  // Required B polarity for an N channel event
	PolN          bool  // Defines active polarity of N (0=low active, 1=high active)
	IgnoreAB      bool  // Ignore A and B polarity for N channel event
	ClrCont       bool  // Always latch or latch and clear X_ENC upon an N event
	ClrOnce       bool  // Latch or latch and clear X_ENC on the next N event
	Sensitivity   uint8 // N channel event sensitivity (2 bits)
	ClrEncX       bool  // Clear encoder counter X_ENC upon N-event
	LatchXAct     bool  // Also latch XACTUAL position together with X_ENC
	EncSelDecimal bool  // Encoder prescaler divisor binary mode (0) / decimal mode (1)
}

// NewENCMODE creates a new ENCMODE register instance
func NewENCMODE() *ENCMODE_Register {
	return &ENCMODE_Register{
		Register: Register{
			RegisterAddr: ENCMODE,
		},
	}
}

// Pack method for ENCMODE: overrides the base Pack
func (e *ENCMODE_Register) Pack() uint32 {
	var registerValue uint32

	// Pack each field using bitwise operations
	if e.PolA {
		registerValue |= 1 << 0
	}
	if e.PolB {
		registerValue |= 1 << 1
	}
	if e.PolN {
		registerValue |= 1 << 2
	}
	if e.IgnoreAB {
		registerValue |= 1 << 3
	}
	if e.ClrCont {
		registerValue |= 1 << 4
	}
	if e.ClrOnce {
		registerValue |= 1 << 5
	}
	registerValue |= uint32(e.Sensitivity&0x3) << 6 // Sensitivity: 2 bits
	if e.ClrEncX {
		registerValue |= 1 << 8
	}
	if e.LatchXAct {
		registerValue |= 1 << 9
	}
	if e.EncSelDecimal {
		registerValue |= 1 << 10
	}

	return registerValue
}

// Unpack method for ENCMODE: overrides the base Unpack
func (e *ENCMODE_Register) Unpack(registerValue uint32) {
	// Unpack each field using bitwise operations
	e.PolA = (registerValue & (1 << 0)) != 0
	e.PolB = (registerValue & (1 << 1)) != 0
	e.PolN = (registerValue & (1 << 2)) != 0
	e.IgnoreAB = (registerValue & (1 << 3)) != 0
	e.ClrCont = (registerValue & (1 << 4)) != 0
	e.ClrOnce = (registerValue & (1 << 5)) != 0
	e.Sensitivity = uint8((registerValue >> 6) & 0x3) // Extract 2 bits for Sensitivity
	e.ClrEncX = (registerValue & (1 << 8)) != 0
	e.LatchXAct = (registerValue & (1 << 9)) != 0
	e.EncSelDecimal = (registerValue & (1 << 10)) != 0
}

// ENC_STATUS_Register struct to represent the ENC_STATUS register
type ENC_STATUS_Register struct {
	Register
	NEvent        bool // N event detected
	DeviationWarn bool // Deviation between X_ACTUAL and X_ENC detected
}

// NewENC_STATUS creates a new ENC_STATUS register instance
func NewENC_STATUS() *ENC_STATUS_Register {
	return &ENC_STATUS_Register{
		Register: Register{
			RegisterAddr: ENC_STATUS,
		},
	}
}

// Pack method for ENC_STATUS: overrides the base Pack
func (e *ENC_STATUS_Register) Pack() uint32 {
	var registerValue uint32

	// Pack each field using bitwise operations
	if e.NEvent {
		registerValue |= 1 << 0
	}
	if e.DeviationWarn {
		registerValue |= 1 << 1
	}

	return registerValue
}

// Unpack method for ENC_STATUS: overrides the base Unpack
func (e *ENC_STATUS_Register) Unpack(registerValue uint32) {
	// Unpack each field using bitwise operations
	e.NEvent = (registerValue & (1 << 0)) != 0
	e.DeviationWarn = (registerValue & (1 << 1)) != 0
}

// CHOPCONF_Register struct to represent the CHOPCONF register
type CHOPCONF_Register struct {
	Register
	Toff       uint8 // Off time setting (4 bits)
	HstrtTfd   uint8 // Hysteresis start value or fast decay time setting (3 bits)
	HendOffset uint8 // Hysteresis low value or sine wave offset (4 bits)
	Tfd3       bool  // Fast decay time setting bit 3
	Disfdcc    bool  // Disable current comparator usage for fast decay termination
	Rndtf      bool  // Enable random modulation of chopper TOFF time
	Chm        bool  // Chopper mode (0=standard, 1=constant off time with fast decay)
	Tbl        uint8 // Comparator blank time select (2 bits)
	Vsense     bool  // Select resistor voltage sensitivity (low or high)
	Vhighfs    bool  // Enable fullstep switching when VHIGH is exceeded
	Vhighchm   bool  // Enable switching to chm=1 and fd=0 when VHIGH is exceeded
	Tpfd       uint8 // Passive fast decay time (4 bits)
	Mres       uint8 // Microstep resolution (4 bits)
	Intpol     bool  // Enable interpolation to 256 microsteps
	Dedge      bool  // Enable double edge step pulses
	Diss2g     bool  // Disable short to GND protection
	Diss2vs    bool  // Disable short to supply protection
}

// NewCHOPCONF creates a new CHOPCONF register instance
func NewCHOPCONF() *CHOPCONF_Register {
	return &CHOPCONF_Register{
		Register: Register{
			RegisterAddr: CHOPCONF,
		},
	}
}

// Pack method for CHOPCONF: overrides the base Pack
func (c *CHOPCONF_Register) Pack() uint32 {
	var registerValue uint32

	// Pack each field using bitwise operations
	registerValue |= uint32(c.Toff&0xF) << 0       // Toff: 4 bits
	registerValue |= uint32(c.HstrtTfd&0x7) << 4   // HstrtTfd: 3 bits
	registerValue |= uint32(c.HendOffset&0xF) << 7 // HendOffset: 4 bits
	if c.Tfd3 {
		registerValue |= 1 << 11 // Tfd3: 1 bit
	}
	if c.Disfdcc {
		registerValue |= 1 << 12 // Disfdcc: 1 bit
	}
	if c.Rndtf {
		registerValue |= 1 << 13 // Rndtf: 1 bit
	}
	if c.Chm {
		registerValue |= 1 << 14 // Chm: 1 bit
	}
	registerValue |= uint32(c.Tbl&0x3) << 15 // Tbl: 2 bits
	if c.Vsense {
		registerValue |= 1 << 17 // Vsense: 1 bit
	}
	if c.Vhighfs {
		registerValue |= 1 << 18 // Vhighfs: 1 bit
	}
	if c.Vhighchm {
		registerValue |= 1 << 19 // Vhighchm: 1 bit
	}
	registerValue |= uint32(c.Tpfd&0xF) << 20 // Tpfd: 4 bits
	registerValue |= uint32(c.Mres&0xF) << 24 // Mres: 4 bits
	if c.Intpol {
		registerValue |= 1 << 28 // Intpol: 1 bit
	}
	if c.Dedge {
		registerValue |= 1 << 29 // Dedge: 1 bit
	}
	if c.Diss2g {
		registerValue |= 1 << 30 // Diss2g: 1 bit
	}
	if c.Diss2vs {
		registerValue |= 1 << 31 // Diss2vs: 1 bit
	}

	return registerValue
}

// Unpack method for CHOPCONF: overrides the base Unpack
func (c *CHOPCONF_Register) Unpack(registerValue uint32) {
	// Unpack each field using bitwise operations
	c.Toff = uint8((registerValue >> 0) & 0xF)       // Extract 4 bits for Toff
	c.HstrtTfd = uint8((registerValue >> 4) & 0x7)   // Extract 3 bits for HstrtTfd
	c.HendOffset = uint8((registerValue >> 7) & 0xF) // Extract 4 bits for HendOffset
	c.Tfd3 = (registerValue & (1 << 11)) != 0        // Extract 1 bit for Tfd3
	c.Disfdcc = (registerValue & (1 << 12)) != 0     // Extract 1 bit for Disfdcc
	c.Rndtf = (registerValue & (1 << 13)) != 0       // Extract 1 bit for Rndtf
	c.Chm = (registerValue & (1 << 14)) != 0         // Extract 1 bit for Chm
	c.Tbl = uint8((registerValue >> 15) & 0x3)       // Extract 2 bits for Tbl
	c.Vsense = (registerValue & (1 << 17)) != 0      // Extract 1 bit for Vsense
	c.Vhighfs = (registerValue & (1 << 18)) != 0     // Extract 1 bit for Vhighfs
	c.Vhighchm = (registerValue & (1 << 19)) != 0    // Extract 1 bit for Vhighchm
	c.Tpfd = uint8((registerValue >> 20) & 0xF)      // Extract 4 bits for Tpfd
	c.Mres = uint8((registerValue >> 24) & 0xF)      // Extract 4 bits for Mres
	c.Intpol = (registerValue & (1 << 28)) != 0      // Extract 1 bit for Intpol
	c.Dedge = (registerValue & (1 << 29)) != 0       // Extract 1 bit for Dedge
	c.Diss2g = (registerValue & (1 << 30)) != 0      // Extract 1 bit for Diss2g
	c.Diss2vs = (registerValue & (1 << 31)) != 0     // Extract 1 bit for Diss2vs
}

// COOLCONF_Register struct to represent the COOLCONF register
type COOLCONF_Register struct {
	Register
	Semin  uint8 // Minimum stallGuard2 value for smart current control (4 bits)
	Seup   uint8 // Current increment step width (2 bits)
	Semax  uint8 // stallGuard2 hysteresis value for smart current control (4 bits)
	Sedn   uint8 // Current decrement step speed (2 bits)
	Seimin bool  // Minimum current for smart current control (1 bit)
	Sgt    uint8 // stallGuard2 threshold value (7 bits)
	Sfilt  bool  // Enable stallGuard2 filter (1 bit)
}

// NewCOOLCONF creates a new COOLCONF register instance
func NewCOOLCONF() *COOLCONF_Register {
	return &COOLCONF_Register{
		Register: Register{
			RegisterAddr: COOLCONF,
		},
	}
}

// Pack method for COOLCONF: overrides the base Pack
func (c *COOLCONF_Register) Pack() uint32 {
	var registerValue uint32

	// Pack each field using bitwise operations
	registerValue |= uint32(c.Semin&0xF) << 0 // Semin: 4 bits
	registerValue |= uint32(c.Seup&0x3) << 5  // Seup: 2 bits
	registerValue |= uint32(c.Semax&0xF) << 8 // Semax: 4 bits
	registerValue |= uint32(c.Sedn&0x3) << 13 // Sedn: 2 bits
	if c.Seimin {
		registerValue |= 1 << 15 // Seimin: 1 bit
	}
	registerValue |= uint32(c.Sgt&0x7F) << 16 // Sgt: 7 bits
	if c.Sfilt {
		registerValue |= 1 << 24 // Sfilt: 1 bit
	}

	return registerValue
}

// Unpack method for COOLCONF: overrides the base Unpack
func (c *COOLCONF_Register) Unpack(registerValue uint32) {
	// Unpack each field using bitwise operations
	c.Semin = uint8((registerValue >> 0) & 0xF) // Extract 4 bits for Semin
	c.Seup = uint8((registerValue >> 5) & 0x3)  // Extract 2 bits for Seup
	c.Semax = uint8((registerValue >> 8) & 0xF) // Extract 4 bits for Semax
	c.Sedn = uint8((registerValue >> 13) & 0x3) // Extract 2 bits for Sedn
	c.Seimin = (registerValue & (1 << 15)) != 0 // Extract 1 bit for Seimin
	c.Sgt = uint8((registerValue >> 16) & 0x7F) // Extract 7 bits for Sgt
	c.Sfilt = (registerValue & (1 << 24)) != 0  // Extract 1 bit for Sfilt
}

// DCCTRL_Register struct to represent the DCCTRL register
type DCCTRL_Register struct {
	Register
	DcTime uint16 // Upper PWM on time limit for commutation (10 bits)
	DcSg   uint8  // Max. PWM on time for step loss detection using dcStep (8 bits)
}

// NewDCCTRL creates a new DCCTRL register instance
func NewDCCTRL() *DCCTRL_Register {
	return &DCCTRL_Register{
		Register: Register{
			RegisterAddr: DCCTRL,
		},
	}
}

// Pack method for DCCTRL: overrides the base Pack
func (d *DCCTRL_Register) Pack() uint32 {
	var registerValue uint32

	// Pack each field using bitwise operations
	registerValue |= uint32(d.DcTime&0x3FF) << 0 // DcTime: 10 bits
	registerValue |= uint32(d.DcSg&0xFF) << 16   // DcSg: 8 bits

	return registerValue
}

// Unpack method for DCCTRL: overrides the base Unpack
func (d *DCCTRL_Register) Unpack(registerValue uint32) {
	// Unpack each field using bitwise operations
	d.DcTime = uint16((registerValue >> 0) & 0x3FF) // Extract 10 bits for DcTime
	d.DcSg = uint8((registerValue >> 16) & 0xFF)    // Extract 8 bits for DcSg
}

// DRV_STATUS_Register struct to represent the DRV_STATUS register
type DRV_STATUS_Register struct {
	Register
	SgResult   uint16 // stallGuard2 result or motor temperature estimation in standstill (9 bits)
	S2vsa      bool   // Short to supply indicator phase A
	S2vsb      bool   // Short to supply indicator phase B
	Stealth    bool   // stealthChop indicator
	FsActive   bool   // Full step active indicator
	CsActual   uint8  // Actual motor current / smart energy current (5 bits)
	StallGuard bool   // stallGuard2 status
	Ot         bool   // Overtemperature flag
	Otpw       bool   // Overtemperature pre-warning flag
	S2ga       bool   // Short to ground indicator phase A
	S2gb       bool   // Short to ground indicator phase B
	Ola        bool   // Open load indicator phase A
	Olb        bool   // Open load indicator phase B
	Stst       bool   // Standstill indicator
}

// NewDRV_STATUS creates a new DRV_STATUS register instance
func NewDRV_STATUS() *DRV_STATUS_Register {
	return &DRV_STATUS_Register{
		Register: Register{
			RegisterAddr: DRV_STATUS,
		},
	}
}

// Pack method for DRV_STATUS: overrides the base Pack
func (d *DRV_STATUS_Register) Pack() uint32 {
	var registerValue uint32

	// Pack each field using bitwise operations
	registerValue |= uint32(d.SgResult&0x1FF) << 0 // SgResult: 9 bits
	if d.S2vsa {
		registerValue |= 1 << 12 // S2vsa: 1 bit
	}
	if d.S2vsb {
		registerValue |= 1 << 13 // S2vsb: 1 bit
	}
	if d.Stealth {
		registerValue |= 1 << 14 // Stealth: 1 bit
	}
	if d.FsActive {
		registerValue |= 1 << 15 // FsActive: 1 bit
	}
	registerValue |= uint32(d.CsActual&0x1F) << 16 // CsActual: 5 bits
	if d.StallGuard {
		registerValue |= 1 << 24 // StallGuard: 1 bit
	}
	if d.Ot {
		registerValue |= 1 << 25 // Ot: 1 bit
	}
	if d.Otpw {
		registerValue |= 1 << 26 // Otpw: 1 bit
	}
	if d.S2ga {
		registerValue |= 1 << 27 // S2ga: 1 bit
	}
	if d.S2gb {
		registerValue |= 1 << 28 // S2gb: 1 bit
	}
	if d.Ola {
		registerValue |= 1 << 29 // Ola: 1 bit
	}
	if d.Olb {
		registerValue |= 1 << 30 // Olb: 1 bit
	}
	if d.Stst {
		registerValue |= 1 << 31 // Stst: 1 bit
	}

	return registerValue
}

// Unpack method for DRV_STATUS: overrides the base Unpack
func (d *DRV_STATUS_Register) Unpack(registerValue uint32) {
	// Unpack each field using bitwise operations
	d.SgResult = uint16((registerValue >> 0) & 0x1FF) // Extract 9 bits for SgResult
	d.S2vsa = (registerValue & (1 << 12)) != 0        // Extract 1 bit for S2vsa
	d.S2vsb = (registerValue & (1 << 13)) != 0        // Extract 1 bit for S2vsb
	d.Stealth = (registerValue & (1 << 14)) != 0      // Extract 1 bit for Stealth
	d.FsActive = (registerValue & (1 << 15)) != 0     // Extract 1 bit for FsActive
	d.CsActual = uint8((registerValue >> 16) & 0x1F)  // Extract 5 bits for CsActual
	d.StallGuard = (registerValue & (1 << 24)) != 0   // Extract 1 bit for StallGuard
	d.Ot = (registerValue & (1 << 25)) != 0           // Extract 1 bit for Ot
	d.Otpw = (registerValue & (1 << 26)) != 0         // Extract 1 bit for Otpw
	d.S2ga = (registerValue & (1 << 27)) != 0         // Extract 1 bit for S2ga
	d.S2gb = (registerValue & (1 << 28)) != 0         // Extract 1 bit for S2gb
	d.Ola = (registerValue & (1 << 29)) != 0          // Extract 1 bit for Ola
	d.Olb = (registerValue & (1 << 30)) != 0          // Extract 1 bit for Olb
	d.Stst = (registerValue & (1 << 31)) != 0         // Extract 1 bit for Stst
}

// PWMCONF_Register struct to represent the PWMCONF register
type PWMCONF_Register struct {
	Register
	PwmOfs       uint8 // User defined PWM amplitude offset (8 bits)
	PwmGrad      uint8 // User defined PWM amplitude gradient (8 bits)
	PwmFreq      uint8 // PWM frequency selection (2 bits)
	PwmAutoscale bool  // Enable PWM automatic amplitude scaling (1 bit)
	PwmAutograd  bool  // PWM automatic gradient adaptation (1 bit)
	Freewheel    uint8 // Standstill option when motor current setting is zero (2 bits)
	PwmReg       uint8 // Regulation loop gradient (4 bits)
	PwmLim       uint8 // PWM automatic scale amplitude limit when switching on (4 bits)
}

// NewPWMCONF creates a new PWMCONF register instance
func NewPWMCONF() *PWMCONF_Register {
	return &PWMCONF_Register{
		Register: Register{
			RegisterAddr: PWMCONF,
		},
	}
}

// Pack method for PWMCONF: overrides the base Pack
func (p *PWMCONF_Register) Pack() uint32 {
	var registerValue uint32

	// Pack each field using bitwise operations
	registerValue |= uint32(p.PwmOfs&0xFF) << 0  // PwmOfs: 8 bits
	registerValue |= uint32(p.PwmGrad&0xFF) << 8 // PwmGrad: 8 bits
	registerValue |= uint32(p.PwmFreq&0x3) << 16 // PwmFreq: 2 bits
	if p.PwmAutoscale {
		registerValue |= 1 << 18 // PwmAutoscale: 1 bit
	}
	if p.PwmAutograd {
		registerValue |= 1 << 19 // PwmAutograd: 1 bit
	}
	registerValue |= uint32(p.Freewheel&0x3) << 20 // Freewheel: 2 bits
	registerValue |= uint32(p.PwmReg&0xF) << 24    // PwmReg: 4 bits
	registerValue |= uint32(p.PwmLim&0xF) << 28    // PwmLim: 4 bits

	return registerValue
}

// Unpack method for PWMCONF: overrides the base Unpack
func (p *PWMCONF_Register) Unpack(registerValue uint32) {
	// Unpack each field using bitwise operations
	p.PwmOfs = uint8((registerValue >> 0) & 0xFF)     // Extract 8 bits for PwmOfs
	p.PwmGrad = uint8((registerValue >> 8) & 0xFF)    // Extract 8 bits for PwmGrad
	p.PwmFreq = uint8((registerValue >> 16) & 0x3)    // Extract 2 bits for PwmFreq
	p.PwmAutoscale = (registerValue & (1 << 18)) != 0 // Extract 1 bit for PwmAutoscale
	p.PwmAutograd = (registerValue & (1 << 19)) != 0  // Extract 1 bit for PwmAutograd
	p.Freewheel = uint8((registerValue >> 20) & 0x3)  // Extract 2 bits for Freewheel
	p.PwmReg = uint8((registerValue >> 24) & 0xF)     // Extract 4 bits for PwmReg
	p.PwmLim = uint8((registerValue >> 28) & 0xF)     // Extract 4 bits for PwmLim
}

// PWM_SCALE_Register struct to represent the PWM_SCALE register
type PWM_SCALE_Register struct {
	Register
	PwmScaleSum  uint8  // Actual PWM duty cycle (8 bits)
	PwmScaleAuto uint16 // Result of the automatic amplitude regulation based on current measurement (9 bits)
}

// NewPWM_SCALE creates a new PWM_SCALE register instance
func NewPWM_SCALE() *PWM_SCALE_Register {
	return &PWM_SCALE_Register{
		Register: Register{
			RegisterAddr: PWM_SCALE,
		},
	}
}

// Pack method for PWM_SCALE: overrides the base Pack
func (p *PWM_SCALE_Register) Pack() uint32 {
	var registerValue uint32

	// Pack each field using bitwise operations
	registerValue |= uint32(p.PwmScaleSum&0xFF) << 0    // PwmScaleSum: 8 bits
	registerValue |= uint32(p.PwmScaleAuto&0x1FF) << 16 // PwmScaleAuto: 9 bits

	return registerValue
}

// Unpack method for PWM_SCALE: overrides the base Unpack
func (p *PWM_SCALE_Register) Unpack(registerValue uint32) {
	// Unpack each field using bitwise operations
	p.PwmScaleSum = uint8((registerValue >> 0) & 0xFF)     // Extract 8 bits for PwmScaleSum
	p.PwmScaleAuto = uint16((registerValue >> 16) & 0x1FF) // Extract 9 bits for PwmScaleAuto
}

// PWM_AUTO_Register struct to represent the PWM_AUTO register
type PWM_AUTO_Register struct {
	Register
	PwmOfsAuto  uint8 // Automatically determined offset value (8 bits)
	PwmGradAuto uint8 // Automatically determined gradient value (8 bits)
}

// NewPWM_AUTO creates a new PWM_AUTO register instance
func NewPWM_AUTO() *PWM_AUTO_Register {
	return &PWM_AUTO_Register{
		Register: Register{
			RegisterAddr: PWM_AUTO,
		},
	}
}

// Pack method for PWM_AUTO: overrides the base Pack
func (p *PWM_AUTO_Register) Pack() uint32 {
	var registerValue uint32

	// Pack each field using bitwise operations
	registerValue |= uint32(p.PwmOfsAuto&0xFF) << 0   // PwmOfsAuto: 8 bits
	registerValue |= uint32(p.PwmGradAuto&0xFF) << 16 // PwmGradAuto: 8 bits

	return registerValue
}

// Unpack method for PWM_AUTO: overrides the base Unpack
func (p *PWM_AUTO_Register) Unpack(registerValue uint32) {
	// Unpack each field using bitwise operations
	p.PwmOfsAuto = uint8((registerValue >> 0) & 0xFF)   // Extract 8 bits for PwmOfsAuto
	p.PwmGradAuto = uint8((registerValue >> 16) & 0xFF) // Extract 8 bits for PwmGradAuto
}

// MSCNT_Register struct to represent the MSCNT register (10-bit value)
type MSCNT_Register struct {
	Register
	Value uint16 // Microstep counter value (10 bits)
}

// NewMSCNT creates a new MSCNT register instance
func NewMSCNT() *MSCNT_Register {
	return &MSCNT_Register{
		Register: Register{
			RegisterAddr: MSCNT,
		},
	}
}

// Pack method for MSCNT: combines the 10-bit value into a 16-bit value
func (m *MSCNT_Register) Pack() uint16 {
	return m.Value & 0x3FF // Mask the value to ensure it is within the 10-bit range (0-1023)
}

// Unpack method for MSCNT: extracts the 10-bit value from a 16-bit value
func (m *MSCNT_Register) Unpack(registerValue uint16) {
	m.Value = registerValue & 0x3FF // Mask to extract the 10-bit value (0-1023)
}

// VDCMIN_Register struct for VDCMIN register (23 bits)
type VDCMIN_Register struct {
	Register
	Value uint32 // 23-bit value
}

// NewVDCMIN creates a new VDCMIN register instance
func NewVDCMIN() *VDCMIN_Register {
	return &VDCMIN_Register{
		Register: Register{
			RegisterAddr: VDCMIN,
		},
	}
}

// Pack method for VDCMIN: packs the 23-bit value into a 32-bit value
func (v *VDCMIN_Register) Pack() uint32 {
	return v.Value & 0x7FFFFF // Mask to 23 bits
}

// Unpack method for VDCMIN: unpacks the 23-bit value from a 32-bit value
func (v *VDCMIN_Register) Unpack(registerValue uint32) {
	v.Value = registerValue & 0x7FFFFF // Mask to 23 bits
}

// XLATCH_Register struct for XLATCH register (32 bits)
type XLATCH_Register struct {
	Register
	Value uint32 // 32-bit value
}

// NewXLATCH creates a new XLATCH register instance
func NewXLATCH() *XLATCH_Register {
	return &XLATCH_Register{
		Register: Register{
			RegisterAddr: XLATCH,
		},
	}
}

// Pack method for XLATCH: directly returns the 32-bit value
func (x *XLATCH_Register) Pack() uint32 {
	return x.Value // No mask needed, since it's 32 bits
}

// Unpack method for XLATCH: unpacks the 32-bit value from a 32-bit register
func (x *XLATCH_Register) Unpack(registerValue uint32) {
	x.Value = registerValue // Direct assignment since it's 32 bits
}

// RAMPMODE_Register struct for RAMPMODE register (2 bits)
type RAMPMODE_Register struct {
	Register
	mode        RampMode // Mode is now an enum-like type
	comm        RegisterComm
	driverIndex uint8
}
type RampMode uint8

const (
	PositioningMode      RampMode = iota // 0
	VelocityPositiveMode                 // 1
	VelocityNegativeMode                 // 2
	HoldMode                             // 3
)

func NewRAMPMODE(comm RegisterComm, driverIndex uint8) *RAMPMODE_Register {
	return &RAMPMODE_Register{
		Register: Register{
			RegisterAddr: RAMPMODE,
		},
		driverIndex: driverIndex,
		comm:        comm,
		mode:        PositioningMode, // Default to Positioning Mode
	}
}

// SetMode sets the mode of the RAMPMODE register
func (r *RAMPMODE_Register) SetMode(mode RampMode) error {
	r.mode = mode
	registerValue := r.Pack()
	return r.comm.WriteRegister(r.RegisterAddr, uint32(registerValue), r.driverIndex)
}

// GetMode returns the current mode of the RAMPMODE register
func (r *RAMPMODE_Register) GetMode() (RampMode, error) {
	registerValue, err := r.comm.ReadRegister(r.RegisterAddr, r.driverIndex)
	if err != nil {
		return 0, err //Defaults to Postioning Mode
	}

	// Unpack the register value to get the mode
	r.Unpack(uint8(registerValue))
	return r.mode, nil

}

// Pack method for RAMPMODE: packs the mode value into a single byte (now using enums)
func (r *RAMPMODE_Register) Pack() uint8 {
	return uint8(r.mode) // Simply cast the mode to uint8
}

// Unpack method for RAMPMODE: unpacks the mode value from a byte
func (r *RAMPMODE_Register) Unpack(registerValue uint8) {
	r.mode = RampMode(registerValue & 0x03) // Mask to 2 bits
}

// String method to display the mode as a string (useful for logging or debugging)
func (r RampMode) String() string {
	switch r {
	case PositioningMode:
		return "Positioning Mode"
	case VelocityPositiveMode:
		return "Velocity Mode (Positive VMAX)"
	case VelocityNegativeMode:
		return "Velocity Mode (Negative VMAX)"
	case HoldMode:
		return "Hold Mode"
	default:
		return "Unknown Mode"
	}
}

// XACTUAL_Register struct for XACTUAL register (32 bits)
type XACTUAL_Register struct {
	Register
	Value uint32 // 32-bit value
}

// NewXACTUAL creates a new XACTUAL register instance
func NewXACTUAL() *XACTUAL_Register {
	return &XACTUAL_Register{
		Register: Register{
			RegisterAddr: XACTUAL,
		},
	}
}

// Pack method for XACTUAL: returns the 32-bit value
func (x *XACTUAL_Register) Pack() uint32 {
	return x.Value // 32 bits, no masking needed
}

// Unpack method for XACTUAL: unpacks the 32-bit value
func (x *XACTUAL_Register) Unpack(registerValue uint32) {
	x.Value = registerValue // Direct assignment since it's 32 bits
}

// VACTUAL_Register struct for VACTUAL register (24 bits)
type VACTUAL_Register struct {
	Register
	Value uint32 // 24-bit value (stored in a 32-bit field)
}

// NewVACTUAL creates a new VACTUAL register instance
func NewVACTUAL() *VACTUAL_Register {
	return &VACTUAL_Register{
		Register: Register{
			RegisterAddr: VACTUAL,
		},
	}
}

// Pack method for VACTUAL: packs the 24-bit value into a 32-bit value
func (v *VACTUAL_Register) Pack() uint32 {
	return v.Value & 0xFFFFFF // Mask to 24 bits
}

// Unpack method for VACTUAL: unpacks the 24-bit value from a 32-bit value
func (v *VACTUAL_Register) Unpack(registerValue uint32) {
	v.Value = registerValue & 0xFFFFFF // Mask to 24 bits
}

// VSTART_Register struct for VSTART register (18 bits)
type VSTART_Register struct {
	Register
	Value uint32 // 18-bit value
}

// NewVSTART creates a new VSTART register instance
func NewVSTART() *VSTART_Register {
	return &VSTART_Register{
		Register: Register{
			RegisterAddr: VSTART,
		},
	}
}

// Pack method for VSTART: packs the 18-bit value into a 16-bit value
func (v *VSTART_Register) Pack() uint32 {
	return v.Value & 0x3FFFF // Mask to 18 bits
}

// Unpack method for VSTART: unpacks the 18-bit value from a 16-bit value
func (v *VSTART_Register) Unpack(registerValue uint32) {
	v.Value = registerValue & 0x3FFFF // Mask to 18 bits
}

// A1_Register struct for A1 register (16 bits)
type A1_Register struct {
	Register
	Value uint16 // 16-bit value
}

// NewA1 creates a new A1 register instance
func NewA1() *A1_Register {
	return &A1_Register{
		Register: Register{
			RegisterAddr: A_1,
		},
	}
}

// Pack method for A1: returns the 16-bit value
func (a *A1_Register) Pack() uint16 {
	return a.Value // 16 bits, no masking needed
}

// Unpack method for A1: unpacks the 16-bit value
func (a *A1_Register) Unpack(registerValue uint16) {
	a.Value = registerValue // Direct assignment since it's 16 bits
}

// V1_Register struct for V1 register (20 bits)
type V1_Register struct {
	Register
	Value uint32 // 20-bit value (stored in a 32-bit field)
}

// NewV1 creates a new V1 register instance
func NewV1() *V1_Register {
	return &V1_Register{
		Register: Register{
			RegisterAddr: V_1,
		},
	}
}

// Pack method for V1: packs the 20-bit value into a 32-bit value
func (v *V1_Register) Pack() uint32 {
	return v.Value & 0xFFFFF // Mask to 20 bits
}

// Unpack method for V1: unpacks the 20-bit value from a 32-bit value
func (v *V1_Register) Unpack(registerValue uint32) {
	v.Value = registerValue & 0xFFFFF // Mask to 20 bits
}

// AMAX_Register struct for AMAX register (16 bits)
type AMAX_Register struct {
	Register
	Value uint16 // 16-bit value
}

// NewAMAX creates a new AMAX register instance
func NewAMAX() *AMAX_Register {
	return &AMAX_Register{
		Register: Register{
			RegisterAddr: AMAX,
		},
	}
}

// Pack method for AMAX: returns the 16-bit value
func (a *AMAX_Register) Pack() uint16 {
	return a.Value // 16 bits, no masking needed
}

// Unpack method for AMAX: unpacks the 16-bit value
func (a *AMAX_Register) Unpack(registerValue uint16) {
	a.Value = registerValue // Direct assignment since it's 16 bits
}

// VMAX_Register struct for VMAX register (23 bits)
type VMAX_Register struct {
	Register
	Value uint32 // 23-bit value (stored in a 32-bit field)
}

// NewVMAX creates a new VMAX register instance
func NewVMAX() *VMAX_Register {
	return &VMAX_Register{
		Register: Register{
			RegisterAddr: VMAX,
		},
	}
}

// Pack method for VMAX: packs the 23-bit value into a 32-bit value
func (v *VMAX_Register) Pack() uint32 {
	return v.Value & 0x7FFFFF // Mask to 23 bits
}

// Unpack method for VMAX: unpacks the 23-bit value from a 32-bit value
func (v *VMAX_Register) Unpack(registerValue uint32) {
	v.Value = registerValue & 0x7FFFFF // Mask to 23 bits
}

// D1_Register struct for D1 register (16 bits)
type D1_Register struct {
	Register
	Value uint16 // 16-bit value
}

// NewD1 creates a new D1 register instance
func NewD1() *D1_Register {
	return &D1_Register{
		Register: Register{
			RegisterAddr: D_1,
		},
	}
}

// Pack method for D1: returns the 16-bit value
func (d *D1_Register) Pack() uint16 {
	return d.Value // 16 bits, no masking needed
}

// Unpack method for D1: unpacks the 16-bit value
func (d *D1_Register) Unpack(registerValue uint16) {
	d.Value = registerValue // Direct assignment since it's 16 bits
}

// VSTOP_Register struct for VSTOP register (18 bits)
type VSTOP_Register struct {
	Register
	Value uint32 // 18-bit value (stored in a 32-bit field)
}

// NewVSTOP creates a new VSTOP register instance
func NewVSTOP() *VSTOP_Register {
	return &VSTOP_Register{
		Register: Register{
			RegisterAddr: VSTOP,
		},
	}
}

// Pack method for VSTOP: packs the 18-bit value into a 32-bit value
func (v *VSTOP_Register) Pack() uint32 {
	return v.Value & 0x3FFFF // Mask to 18 bits
}

// Unpack method for VSTOP: unpacks the 18-bit value from a 32-bit value
func (v *VSTOP_Register) Unpack(registerValue uint32) {
	v.Value = registerValue & 0x3FFFF // Mask to 18 bits
}

// TZEROWAIT_Register struct for TZEROWAIT register (16 bits)
type TZEROWAIT_Register struct {
	Register
	Value uint16 // 16-bit value
}

// NewTZEROWAIT creates a new TZEROWAIT register instance
func NewTZEROWAIT() *TZEROWAIT_Register {
	return &TZEROWAIT_Register{
		Register: Register{
			RegisterAddr: TZEROWAIT,
		},
	}
}

// Pack method for TZEROWAIT: returns the 16-bit value
func (t *TZEROWAIT_Register) Pack() uint16 {
	return t.Value // 16 bits, no masking needed
}

// Unpack method for TZEROWAIT: unpacks the 16-bit value
func (t *TZEROWAIT_Register) Unpack(registerValue uint16) {
	t.Value = registerValue // Direct assignment since it's 16 bits
}

// XTARGET_Register struct for XTARGET register (32 bits)
type XTARGET_Register struct {
	Register
	Value uint32 // 32-bit value
}

// NewXTARGET creates a new XTARGET register instance
func NewXTARGET() *XTARGET_Register {
	return &XTARGET_Register{
		Register: Register{
			RegisterAddr: XTARGET,
		},
	}
}

// Pack method for XTARGET: returns the 32-bit value
func (x *XTARGET_Register) Pack() uint32 {
	return x.Value // 32 bits, no masking needed
}

// Unpack method for XTARGET: unpacks the 32-bit value
func (x *XTARGET_Register) Unpack(registerValue uint32) {
	x.Value = registerValue // Direct assignment since it's 32 bits
}

// X_COMPARE_Register struct for X_COMPARE register (32 bits)
type X_COMPARE_Register struct {
	Register
	Value uint32 // 32-bit value for position comparison
}

// NewX_COMPARE creates a new X_COMPARE register instance
func NewX_COMPARE() *X_COMPARE_Register {
	return &X_COMPARE_Register{
		Register: Register{
			RegisterAddr: X_COMPARE,
		},
	}
}

// Pack method for X_COMPARE: returns the 32-bit value
func (x *X_COMPARE_Register) Pack() uint32 {
	return x.Value // 32 bits, no masking needed
}

// Unpack method for X_COMPARE: unpacks the 32-bit value
func (x *X_COMPARE_Register) Unpack(registerValue uint32) {
	x.Value = registerValue // Direct assignment since it's 32 bits
}

// GLOBAL_SCALER_Register struct for GLOBAL SCALER register (8 bits)
type GLOBAL_SCALER_Register struct {
	Register
	Value uint8 // 8-bit value for global motor current scaling
}

// NewGLOBAL_SCALER creates a new GLOBAL_SCALER register instance
func NewGLOBAL_SCALER() *GLOBAL_SCALER_Register {
	return &GLOBAL_SCALER_Register{
		Register: Register{
			RegisterAddr: GLOBAL_SCALER,
		},
	}
}

// Pack method for GLOBAL_SCALER: returns the 8-bit value
func (g *GLOBAL_SCALER_Register) Pack() uint8 {
	return g.Value // 8 bits, no masking needed
}

// Unpack method for GLOBAL_SCALER: unpacks the 8-bit value
func (g *GLOBAL_SCALER_Register) Unpack(registerValue uint8) {
	g.Value = registerValue // Direct assignment since it's 8 bits
}

// TPOWERDOWN_Register struct for TPOWERDOWN register (8 bits)
type TPOWERDOWN_Register struct {
	Register
	Value uint8 // 8-bit value for time delay after standstill
}

// NewTPOWERDOWN creates a new TPOWERDOWN register instance
func NewTPOWERDOWN() *TPOWERDOWN_Register {
	return &TPOWERDOWN_Register{
		Register: Register{
			RegisterAddr: TPOWERDOWN,
		},
	}
}

// Pack method for TPOWERDOWN: returns the 8-bit value
func (t *TPOWERDOWN_Register) Pack() uint8 {
	return t.Value // 8 bits, no masking needed
}

// Unpack method for TPOWERDOWN: unpacks the 8-bit value
func (t *TPOWERDOWN_Register) Unpack(registerValue uint8) {
	t.Value = registerValue // Direct assignment since it's 8 bits
}

// PWMTHRS_Register struct for PWMTHRS register (20 bits)
type PWMTHRS_Register struct {
	Register
	Value uint32 // 20-bit value (stored in a 32-bit field)
}

// NewPWMTHRS creates a new PWMTHRS register instance
func NewPWMTHRS() *PWMTHRS_Register {
	return &PWMTHRS_Register{
		Register: Register{
			RegisterAddr: TPWMTHRS,
		},
	}
}

// Pack method for PWMTHRS: packs the 20-bit value into a 32-bit value
func (p *PWMTHRS_Register) Pack() uint32 {
	return p.Value & 0xFFFFF // Mask to 20 bits
}

// Unpack method for PWMTHRS: unpacks the 20-bit value from a 32-bit value
func (p *PWMTHRS_Register) Unpack(registerValue uint32) {
	p.Value = registerValue & 0xFFFFF // Mask to 20 bits
}

// TCOOLTHRS_Register struct for TCOOLTHRS register (20 bits)
type TCOOLTHRS_Register struct {
	Register
	Value uint32 // 20-bit value (stored in a 32-bit field)
}

// NewTCOOLTHRS creates a new TCOOLTHRS register instance
func NewTCOOLTHRS() *TCOOLTHRS_Register {
	return &TCOOLTHRS_Register{
		Register: Register{
			RegisterAddr: TCOOLTHRS,
		},
	}
}

// Pack method for TCOOLTHRS: packs the 20-bit value into a 32-bit value
func (t *TCOOLTHRS_Register) Pack() uint32 {
	return t.Value & 0xFFFFF // Mask to 20 bits
}

// Unpack method for TCOOLTHRS: unpacks the 20-bit value from a 32-bit value
func (t *TCOOLTHRS_Register) Unpack(registerValue uint32) {
	t.Value = registerValue & 0xFFFFF // Mask to 20 bits
}

// THIGH_Register struct for THIGH register (16 bits)
type THIGH_Register struct {
	Register
	Value uint16 // 16-bit value
}

// NewTHIGH creates a new THIGH register instance
func NewTHIGH() *THIGH_Register {
	return &THIGH_Register{
		Register: Register{
			RegisterAddr: THIGH,
		},
	}
}

// Pack method for THIGH: returns the 16-bit value
func (t *THIGH_Register) Pack() uint16 {
	return t.Value // 16 bits, no masking needed
}

// Unpack method for THIGH: unpacks the 16-bit value
func (t *THIGH_Register) Unpack(registerValue uint16) {
	t.Value = registerValue // Direct assignment since it's 16 bits
}

// DMAX_Register struct for DMAX register (16 bits)
type DMAX_Register struct {
	Register
	Value uint16 // 16-bit value for deceleration between VMAX and VSTOP
}

// NewDMAX creates a new DMAX register instance
func NewDMAX() *DMAX_Register {
	return &DMAX_Register{
		Register: Register{
			RegisterAddr: DMAX,
		},
	}
}

// Pack method for DMAX: returns the 16-bit value
func (d *DMAX_Register) Pack() uint16 {
	return d.Value // 16 bits, no masking needed
}

// Unpack method for DMAX: unpacks the 16-bit value
func (d *DMAX_Register) Unpack(registerValue uint16) {
	d.Value = registerValue // Direct assignment since it's 16 bits
}

// TSTEP_Register struct for TSTEP register (20 bits)
type TSTEP_Register struct {
	Register
	Value uint32 // 20-bit value (stored in a 32-bit field)
}

// NewTSTEP creates a new TSTEP register instance
func NewTSTEP() *TSTEP_Register {
	return &TSTEP_Register{
		Register: Register{
			RegisterAddr: TSTEP,
		},
	}
}

// Pack method for TSTEP: packs the 20-bit value into a 32-bit value
func (t *TSTEP_Register) Pack() uint32 {
	return t.Value & 0xFFFFF // Mask to 20 bits
}

// Unpack method for TSTEP: unpacks the 20-bit value from a 32-bit value
func (t *TSTEP_Register) Unpack(registerValue uint32) {
	t.Value = registerValue & 0xFFFFF // Mask to 20 bits
}

// X_ENC_Register struct for X_ENC register (32 bits)
type X_ENC_Register struct {
	Register
	Value int32 // 32-bit signed value for actual encoder position
}

// NewX_ENC creates a new X_ENC register instance
func NewX_ENC() *X_ENC_Register {
	return &X_ENC_Register{
		Register: Register{
			RegisterAddr: X_ENC,
		},
	}
}

// Pack method for X_ENC: returns the 32-bit signed value
func (x *X_ENC_Register) Pack() int32 {
	return x.Value // 32 bits, no masking needed for signed integer
}

// Unpack method for X_ENC: unpacks the 32-bit signed value
func (x *X_ENC_Register) Unpack(registerValue int32) {
	x.Value = registerValue // Direct assignment since it's 32 bits signed integer
}

// ENC_CONST_Register struct for ENC_CONST register (32 bits)
type ENC_CONST_Register struct {
	Register
	Value int32 // 32-bit signed accumulation constant
}

// NewENC_CONST creates a new ENC_CONST register instance
func NewENC_CONST() *ENC_CONST_Register {
	return &ENC_CONST_Register{
		Register: Register{
			RegisterAddr: ENC_CONST,
		},
	}
}

// Pack method for ENC_CONST: returns the 32-bit signed accumulation constant
func (e *ENC_CONST_Register) Pack() int32 {
	return e.Value // 32 bits, no masking needed for signed integer
}

// Unpack method for ENC_CONST: unpacks the 32-bit signed accumulation constant
func (e *ENC_CONST_Register) Unpack(registerValue int32) {
	e.Value = registerValue // Direct assignment since it's 32 bits signed integer
}

// ENC_LATCH_Register struct for ENC_LATCH register (32 bits)
type ENC_LATCH_Register struct {
	Register
	Value int32 // 32-bit signed value for encoder position latched on N event
}

// NewENC_LATCH creates a new ENC_LATCH register instance
func NewENC_LATCH() *ENC_LATCH_Register {
	return &ENC_LATCH_Register{
		Register: Register{
			RegisterAddr: ENC_LATCH,
		},
	}
}

// Pack method for ENC_LATCH: returns the 32-bit signed value
func (e *ENC_LATCH_Register) Pack() int32 {
	return e.Value // 32 bits, no masking needed for signed integer
}

// Unpack method for ENC_LATCH: unpacks the 32-bit signed value
func (e *ENC_LATCH_Register) Unpack(registerValue int32) {
	e.Value = registerValue // Direct assignment since it's 32 bits signed integer
}

// ENC_DEVIATION_Register struct for ENC_DEVIATION register (20 bits)
type ENC_DEVIATION_Register struct {
	Register
	Value uint32 // 20-bit unsigned value for maximum deviation
}

// NewENC_DEVIATION creates a new ENC_DEVIATION register instance
func NewENC_DEVIATION() *ENC_DEVIATION_Register {
	return &ENC_DEVIATION_Register{
		Register: Register{
			RegisterAddr: ENC_DEVIATION,
		},
	}
}

// Pack method for ENC_DEVIATION: packs the 20-bit value into a 32-bit value
func (e *ENC_DEVIATION_Register) Pack() uint32 {
	return e.Value & 0xFFFFF // Mask to 20 bits
}

// Unpack method for ENC_DEVIATION: unpacks the 20-bit value from a 32-bit value
func (e *ENC_DEVIATION_Register) Unpack(registerValue uint32) {
	e.Value = registerValue & 0xFFFFF // Mask to 20 bits
}

// MSCURACT_Register struct for MSCURACT register (18 bits)
type MSCURACT_Register struct {
	Register
	CUR_B int16 // 9-bit signed value for motor phase B (sine wave)
	CUR_A int16 // 9-bit signed value for motor phase A (cosine wave)
}

// NewMSCURACT creates a new MSCURACT register instance
func NewMSCURACT() *MSCURACT_Register {
	return &MSCURACT_Register{
		Register: Register{
			RegisterAddr: MSCURACT,
		},
	}
}

// Pack method for MSCURACT: packs the 9-bit signed values for CUR_B and CUR_A into a 32-bit value
func (m *MSCURACT_Register) Pack() uint32 {
	return uint32(m.CUR_A<<16 | m.CUR_B) // Combine CUR_A and CUR_B into a 32-bit value
}

// Unpack method for MSCURACT: unpacks the 32-bit value into CUR_B and CUR_A
func (m *MSCURACT_Register) Unpack(registerValue uint32) {
	m.CUR_B = int16(registerValue & 0x1FF)         // Mask to get the lower 9 bits for CUR_B
	m.CUR_A = int16((registerValue >> 16) & 0x1FF) // Mask to get the next 9 bits for CUR_A
}

// LOST_STEPS_Register struct for LOST_STEPS register (20 bits)
type LOST_STEPS_Register struct {
	Register
	Value uint32 // 20-bit unsigned value for lost steps count
}

// NewLOST_STEPS creates a new LOST_STEPS register instance
func NewLOST_STEPS() *LOST_STEPS_Register {
	return &LOST_STEPS_Register{
		Register: Register{
			RegisterAddr: LOST_STEPS,
		},
	}
}

// Pack method for LOST_STEPS: returns the 20-bit value
func (l *LOST_STEPS_Register) Pack() uint32 {
	return l.Value & 0xFFFFF // Mask to 20 bits
}

// Unpack method for LOST_STEPS: unpacks the 20-bit value from a 32-bit value
func (l *LOST_STEPS_Register) Unpack(registerValue uint32) {
	l.Value = registerValue & 0xFFFFF // Mask to 20 bits
}

// MSLUTSEL_Register struct for MSLUTSEL register (32 bits)
type MSLUTSEL_Register struct {
	Register
	X3 uint8 // 3-bit value for LUT segment 3 start
	X2 uint8 // 3-bit value for LUT segment 2 start
	X1 uint8 // 3-bit value for LUT segment 1 start
	W3 uint8 // 2-bit value for LUT width control W3
	W2 uint8 // 2-bit value for LUT width control W2
	W1 uint8 // 2-bit value for LUT width control W1
	W0 uint8 // 2-bit value for LUT width control W0
}

// NewMSLUTSEL creates a new MSLUTSEL register instance
func NewMSLUTSEL() *MSLUTSEL_Register {
	return &MSLUTSEL_Register{
		Register: Register{
			RegisterAddr: MSLUTSEL,
		},
	}
}

// Pack method for MSLUTSEL: combines all the fields into a 32-bit value
func (m *MSLUTSEL_Register) Pack() uint32 {
	return uint32(m.X3<<27 | m.X2<<24 | m.X1<<21 | m.W3<<18 | m.W2<<16 | m.W1<<14 | m.W0<<12) // Combine fields into a 32-bit value
}

// Unpack method for MSLUTSEL: unpacks the 32-bit value into individual fields
func (m *MSLUTSEL_Register) Unpack(registerValue uint32) {
	m.X3 = uint8((registerValue >> 27) & 0x07) // Extract the 3 bits for X3
	m.X2 = uint8((registerValue >> 24) & 0x07) // Extract the 3 bits for X2
	m.X1 = uint8((registerValue >> 21) & 0x07) // Extract the 3 bits for X1
	m.W3 = uint8((registerValue >> 18) & 0x03) // Extract the 2 bits for W3
	m.W2 = uint8((registerValue >> 16) & 0x03) // Extract the 2 bits for W2
	m.W1 = uint8((registerValue >> 14) & 0x03) // Extract the 2 bits for W1
	m.W0 = uint8((registerValue >> 12) & 0x03) // Extract the 2 bits for W0
}

// MSLUT_Register struct for MSLUT register (32 bits)
type MSLUT_Register struct {
	Register
	Value uint32 // 32-bit value for microstep table entry
}

// NewMSLUT creates a new MSLUT register instance
func NewMSLUT() *MSLUT_Register {
	return &MSLUT_Register{
		Register: Register{
			RegisterAddr: MSLUT0,
		},
	}
}

// Pack method for MSLUT: returns the 32-bit value for the microstep entry
func (m *MSLUT_Register) Pack() uint32 {
	return m.Value // 32 bits, no masking needed
}

// Unpack method for MSLUT: unpacks the 32-bit value into the microstep entry
func (m *MSLUT_Register) Unpack(registerValue uint32) {
	m.Value = registerValue // Direct assignment since it's 32 bits
}

// MSLUTSTART_Register struct for MSLUTSTART register (16 bits)
type MSLUTSTART_Register struct {
	Register
	START_SIN   int8 // 8-bit signed value for the absolute current at microstep entry 0
	START_SIN90 int8 // 8-bit signed value for the absolute current at microstep entry 256
}

// NewMSLUTSTART creates a new MSLUTSTART register instance
func NewMSLUTSTART() *MSLUTSTART_Register {
	return &MSLUTSTART_Register{
		Register: Register{
			RegisterAddr: MSLUTSTART,
		},
	}
}

// Pack method for MSLUTSTART: combines START_SIN and START_SIN90 into a 16-bit value
func (m *MSLUTSTART_Register) Pack() uint16 {
	return uint16(m.START_SIN) | (uint16(m.START_SIN90) << 8) // Combine the 8-bit values into a 16-bit value
}

// Unpack method for MSLUTSTART: unpacks the 16-bit value into START_SIN and START_SIN90
func (m *MSLUTSTART_Register) Unpack(registerValue uint16) {
	m.START_SIN = int8(registerValue & 0xFF)          // Extract the lower 8 bits for START_SIN
	m.START_SIN90 = int8((registerValue >> 8) & 0xFF) // Extract the upper 8 bits for START_SIN90
}

// Function to calculate the sine wave values for the microstep table
func calculateSineWaveTable() []int {
	// Create a slice to store the sine wave table
	table := make([]int, 256)

	// Loop through each table index (i)
	for i := 0; i < 256; i++ {
		// Calculate the sine value and scale it by 248
		sineValue := 248 * math.Sin(2*math.Pi*float32(i)/1024)

		// Round the result and subtract 1
		roundedValue := int(math.Round(sineValue)) - 1

		// Store the value in the table
		table[i] = roundedValue
	}

	return table
}
