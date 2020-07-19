package p1am

//go:generate go run ./internal/cmd/gen_defines

type ModuleProps struct {
	ModuleID                                 uint32
	DI, DO, AI, AO, Status, Config, DataSize byte
	Name                                     string
}

var modules = []ModuleProps{
	//{0x000000ID,di,do,ai,ao,st,cf,ds}
	{0x00000000, 0, 0, 0, 0, 0, 0, 0, "Empty"}, //Empty first entry for defaultgs

	{0x04A00081, 1, 0, 0, 0, 0, 0, 1, "P1-08ND3"}, //P1-08ND3

	{0x04A00085, 1, 0, 0, 0, 0, 0, 1, "P1-08NA"}, //P1-08NA

	{0x04A00087, 1, 0, 0, 0, 0, 0, 1, "P1-08SIM"}, //P1-08SIM

	{0x04A00088, 1, 0, 0, 0, 0, 0, 1, "P1-08NE3"}, //P1-08NE3

	{0x05200082, 2, 0, 0, 0, 0, 0, 1, "P1-16ND3"}, //P1-16ND3

	{0x05200089, 2, 0, 0, 0, 0, 0, 1, "P1-16NE3"}, //P1-16NE3

	{0x1403F481, 0, 0, 0, 32, 4, 4, 0xA0, "P1-04PWM"}, //P1-04PWM

	{0x1404008D, 0, 1, 0, 0, 0, 0, 1, "P1-08TA"}, //P1-08TA

	{0x1404008F, 0, 1, 0, 0, 0, 0, 1, "P1-08TRS"}, //P1-08TRS

	{0x14040091, 0, 2, 0, 0, 0, 0, 1, "P1-16TR"}, //P1-16TR

	{0x14050081, 0, 1, 0, 0, 0, 0, 1, "P1-08TD1"}, //P1-08TD1

	{0x14050082, 0, 1, 0, 0, 0, 0, 1, "P1-08TD2"}, //P1-08TD2

	{0x14080085, 0, 2, 0, 0, 0, 0, 1, "P1-15TD1"}, //P1-15TD1

	{0x14080086, 0, 2, 0, 0, 0, 0, 1, "P1-15TD2"}, //P1-15TD2

	{0x24A50081, 1, 1, 0, 0, 0, 0, 1, "P1-16CDR"}, //P1-16CDR

	{0x24A50082, 1, 1, 0, 0, 0, 0, 1, "P1-15CDD1"}, //P1-15CDD1

	{0x24A50083, 1, 1, 0, 0, 0, 0, 1, "P1-15CDD2"}, //P1-15CDD2

	{0x34605581, 0, 0, 16, 0, 12, 18, 16, "P1-04AD"}, //P1-04AD

	{0x34605588, 0, 0, 16, 0, 12, 8, 16, "P1-04RTD"}, //P1-04RTD

	{0x3460558F, 0, 0, 16, 0, 12, 2, 12, "P1-04ADL-1"}, //P1-04ADL-1

	{0x34605590, 0, 0, 16, 0, 12, 2, 12, "P1-04ADL-2"}, //P1-04ADL-2

	{0x34608C81, 0, 0, 16, 0, 12, 20, 32, "P1-04THM"}, //P1-04THM

	{0x34608C8E, 0, 0, 16, 0, 12, 8, 32, "P1-04NTC"}, //P1-04NTC

	{0x34A0558A, 0, 0, 32, 0, 12, 2, 12, "P1-08ADL-1"}, //P1-08ADL-1

	{0x34A0558B, 0, 0, 32, 0, 12, 2, 12, "P1-08ADL-2"}, //P1-08ADL-2

	{0x34A5A481, 2, 0, 36, 36, 4, 12, 0xC0, "P1-02HSC"}, //P1-02HSC

	{0x44035583, 0, 0, 0, 16, 4, 0, 12, "P1-04DAL-1"}, //P1-04DAL-1

	{0x44035584, 0, 0, 0, 16, 4, 0, 12, "P1-04DAL-2"}, //P1-04DAL-2

	{0x44055588, 0, 0, 0, 32, 4, 0, 12, "P1-08DAL-1"}, //P1-08DAL-1

	{0x44055589, 0, 0, 0, 32, 4, 0, 12, "P1-08DAL-2"}, //P1-08DAL-2

	{0x5461A783, 0, 0, 16, 8, 12, 2, 12, "P1-4ADL2DAL-1"}, //P1-4ADL2DAL-1

	{0x5461A784, 0, 0, 16, 8, 12, 2, 12, "P1-4ADL2DAL-2"}, //P1-4ADL2DAL-2

	{0xFFFFFFFF, 0, 0, 0, 0, 0, 0, 0, "BAD SLOT"}, //empty in case no modules are defined.

	{0x00000000, 0, 0, 0, 0, 0, 0, 0, "BAD SLOT"}, //empty in case no modules are defined.
}

var defaultConfig = map[uint32][]byte{
	0x34605590:// P1_04ADL_2_DEFAULT_CONFIG
	{0x40, 0x03},
	0x34608C8E: // P1_04NTC_DEFAULT_CONFIG
	{0x40, 0x03, 0x60, 0x05,
		0x20, 0x00, 0x80, 0x02},
	0x34608C81: // P1_04THM_DEFAULT_CONFIG
	{0x40, 0x03, 0x60, 0x05,
		0x21, 0x00, 0x22, 0x00,
		0x23, 0x00, 0x24, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00},
	0x34605588: // P1_04RTD_DEFAULT_CONFIG
	{0x40, 0x03, 0x60, 0x05,
		0x20, 0x01, 0x80, 0x00},
	0x34605581: // P1_04AD_DEFAULT_CONFIG
	{0x40, 0x03, 0x00, 0x00,
		0x20, 0x03, 0x00, 0x00,
		0x21, 0x03, 0x00, 0x00,
		0x22, 0x03, 0x00, 0x00,
		0x23, 0x03},
	0x3460558F:// P1_04ADL_1_DEFAULT_CONFIG
	{0x40, 0x03},
	0x34A0558A:// P1_08ADL_1_DEFAULT_CONFIG
	{0x40, 0x07},
	0x34A0558B:// P1_08ADL_2_DEFAULT_CONFIG
	{0x40, 0x07},
	0x5461A783:// P1_04ADL2DAL_1_DEFAULT_CONFIG
	{0x40, 0x03},
	0x5461A784:// P1_04ADL2DAL_2_DEFAULT_CONFIG
	{0x40, 0x03},
	0x1403F481:// P1_04PWM_DEFAULT_CONFIG
	{0x02, 0x02, 0x02, 0x02},
	0x34A5A481: // P1_02HSC_DEFAULT_CONFIG
	{0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x01},
}

const NUMBER_OF_MODULES = 15 //Current max 15 Modules
const SWITCH_BUILTIN = 31
const baseEnable = 33
const MOD_HDR = 0x02
const VERSION_HDR = 0x03
const ACTIVE_HDR = 0x04
const DROPOUT_HDR = 0x05
const CFG_HDR = 0x10
const READ_CFG_HDR = 0x11
const PETWD_HDR = 0x30
const STARTWD_HDR = 0x31
const STOPWD_HDR = 0x32
const CONFIGWD_HDR = 0x33
const READ_STATUS_HDR = 0x40
const READ_DISCRETE_HDR = 0x50
const READ_ANALOG_HDR = 0x51
const READ_BLOCK_HDR = 0x52
const WRITE_DISCRETE_HDR = 0x60
const WRITE_ANALOG_HDR = 0x61
const WRITE_BLOCK_HDR = 0x62
const FW_UPDATE_HDR = 0xAA
const DUMMY = 0xFF
const EMPTY_SLOT_ID = 0xFFFFFFFE
const MAX_TIMEOUT = 0xFFFFFFFF
const DISCRETE_IN_BLOCK = 0
const ANALOG_IN_BLOCK = 1
const DISCRETE_OUT_BLOCK = 2
const ANALOG_OUT_BLOCK = 3
const STATUS_IN_BLOCK = 4
const MISSING24V_STATUS = 3
const BURNOUT_STATUS = 5
const UNDER_RANGE_STATUS = 7
const OVER_RANGE_STATUS = 11
const TOGGLE = 0x01
const HOLD = 0x00
