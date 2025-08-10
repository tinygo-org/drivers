package ina219

type Config struct {
	// BusVoltageRange sets the bus voltage range.
	BusVoltageRange BusVoltageRange

	// PGA sets the programmable gain amplifier.
	PGA PGA

	// BusADC sets the bus ADC resolution.
	BusADC BusADC

	// ShuntADC sets the shunt ADC resolution.
	ShuntADC ShuntADC

	// Mode sets the operating mode.
	Mode Mode

	// Calibration sets the calibration value for the expected
	// voltage and current values.
	Calibration Calibration

	// 1000 / uA per bit
	CurrentDivider float32

	// 1mW per bit
	PowerMultiplier float32
}

// RegisterValue returns the register value of the configuration.
func (c *Config) RegisterValue() uint16 {
	return c.BusVoltageRange.RegisterValue() |
		c.PGA.RegisterValue() |
		c.BusADC.RegisterValue() |
		c.ShuntADC.RegisterValue() |
		c.Mode.RegisterValue()
}

// Generate a new configuration from a register value.
func NewConfig(config int16, calibration int16) Config {
	return Config{
		BusVoltageRange: BusVoltageRange(config >> 13 & 0x1),
		PGA:             PGA(config >> 11 & 0x3),
		BusADC:          BusADC(config >> 7 & 0xF),
		ShuntADC:        ShuntADC(config >> 3 & 0xF),
		Mode:            Mode(config & 0x7),
		Calibration:     Calibration(calibration),
	}
}

// Configurations from
// https://github.com/adafruit/Adafruit_INA219/blob/master/Adafruit_INA219.cpp
var (
	// Config32V2A is a configuration for a 32V 2A range.
	Config32V2A = Config{
		BusVoltageRange: Range32V,
		PGA:             PGA8,
		BusADC:          ADC12,
		ShuntADC:        SADC12,
		Mode:            ModeContShuntBus,
		Calibration:     Calibration16V400mA,
		CurrentDivider:  10.0,
		PowerMultiplier: 2.0,
	}

	// Config32V1A is a configuration for a 32V 1A range.
	Config32V1A = Config{
		BusVoltageRange: Range32V,
		PGA:             PGA8,
		BusADC:          ADC12,
		ShuntADC:        SADC12,
		Mode:            ModeContShuntBus,
		Calibration:     Calibration32V1A,
		CurrentDivider:  25.0,
		PowerMultiplier: 0.8,
	}

	// Config16V400mA is a configuration for a 16V 400mA range.
	Config16V400mA = Config{
		BusVoltageRange: Range16V,
		PGA:             PGA1,
		BusADC:          ADC12,
		ShuntADC:        SADC12,
		Mode:            ModeContShuntBus,
		Calibration:     Calibration16V400mA,
		CurrentDivider:  20.0,
		PowerMultiplier: 1.0,
	}
)

// BusVoltageRange is the bus voltage range.
type BusVoltageRange int8

const (
	Range16V BusVoltageRange = 0 // 0-16V
	Range32V BusVoltageRange = 1 // 0-32V
)

func (r BusVoltageRange) RegisterValue() uint16 {
	return uint16(r) << 13
}

// PGA is the programmable gain amplifier.
type PGA int8

const (
	PGA1 PGA = 0 // 40mV
	PGA2 PGA = 1 // 80mV
	PGA4 PGA = 2 // 160mV
	PGA8 PGA = 3 // 320mV
)

func (p PGA) RegisterValue() uint16 {
	return uint16(p) << 11
}

// BusADC is the bus ADC resolution.
type BusADC int8

const (
	ADC9  BusADC = 0 // 9-bit
	ADC10 BusADC = 1 // 10-bit
	ADC11 BusADC = 2 // 11-bit
	ADC12 BusADC = 3 // 12-bit
)

func (b BusADC) RegisterValue() uint16 {
	return uint16(b) << 7
}

// ShuntADC is the shunt ADC resolution.
type ShuntADC int8

const (
	SADC9  ShuntADC = 0 // 9-bit
	SADC10 ShuntADC = 1 // 10-bit
	SADC11 ShuntADC = 2 // 11-bit
	SADC12 ShuntADC = 3 // 12-bit
)

func (s ShuntADC) RegisterValue() uint16 {
	return uint16(s) << 3
}

// Mode is the operating mode.
type Mode int8

const (
	ModePowerDown    Mode = 0 // power-down
	ModeTrigShunt    Mode = 1 // triggered shunt voltage
	ModeTrigBus      Mode = 2 // triggered bus voltage
	ModeTrigShuntBus Mode = 3 // triggered shunt and bus voltage
	ModeADCOff       Mode = 4 // ADC off
	ModeContShunt    Mode = 5 // continuous shunt voltage
	ModeContBus      Mode = 6 // continuous bus voltage
	ModeContShuntBus Mode = 7 // continuous shunt and bus voltage
)

// ModeTriggered is a mask for triggered modes.
const ModeTriggeredMask Mode = 0x4

// ModeTriggered returns true if the mode is a triggered mode.
func ModeTriggered(m Mode) bool {
	return m != ModePowerDown && m&ModeTriggeredMask == 0
}

func (m Mode) RegisterValue() uint16 {
	return uint16(m)
}

// Calibration is the calibration register for the INA219. Values from:
// https://github.com/adafruit/Adafruit_INA219/blob/master/Adafruit_INA219.cpp
type Calibration uint16

const (
	Calibration32V2A    Calibration = 4096
	Calibration32V1A    Calibration = 10240
	Calibration16V400mA Calibration = 8192
)

func (c Calibration) RegisterValue() uint16 {
	return uint16(c)
}
