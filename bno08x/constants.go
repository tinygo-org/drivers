package bno08x

// I2C and protocol constants
const (
	shtpHeaderLength = 4
	maxTransferOut   = 256
	maxTransferIn    = 384

	i2cDefaultChunk = 32
	continueMask    = 0x8000
)

// SHTP channel numbers
const (
	channelCommand      = 0
	channelExecutable   = 1
	channelControl      = 2
	channelSensorReport = 3
	channelWakeReport   = 4
	channelGyroRV       = 5
)

// SH-2 report IDs
const (
	reportProdIDReq      = 0xF9
	reportProdIDResp     = 0xF8
	reportSetFeature     = 0xFD
	reportGetFeature     = 0xFE
	reportGetFeatureResp = 0xFC
	reportCommandReq     = 0xF2
	reportCommandResp    = 0xF1
	reportFRSWriteReq    = 0xF7
	reportFRSWriteData   = 0xF6
	reportFRSReadReq     = 0xF4
	reportFRSReadResp    = 0xF3
	reportBaseTimestamp  = 0xFB
	reportTimestampReuse = 0xFA
	reportForceFlush     = 0xF0
	reportFlushCompleted = 0xEF
	reportResetReq       = 0xF1
	reportResetResp      = 0xF0
)

// SH-2 commands
const (
	cmdErrors         = 0x01
	cmdCounts         = 0x02
	cmdTare           = 0x03
	cmdInitialize     = 0x04
	cmdFRS            = 0x05
	cmdDCD            = 0x06
	cmdMECal          = 0x07
	cmdProdIDReq      = 0x07
	cmdDCDSave        = 0x09
	cmdGetOscType     = 0x0A
	cmdClearDCDReset  = 0x0B
	cmdCal            = 0x0C
	cmdBootloader     = 0x0D
	cmdInteractiveZRO = 0x0E

	// Command parameters
	initSystem      = 0x01
	initUnsolicited = 0x80

	countsClearCounts = 0x01
	countsGetCounts   = 0x00

	tareTareNow          = 0x00
	tarePersist          = 0x01
	tareSetReorientation = 0x02

	calStart  = 0x00
	calFinish = 0x01

	commandParamCount  = 9
	responseValueCount = 11
)

// Feature report flags
const (
	featChangeSensitivityRelative = 0x01
	featChangeSensitivityEnabled  = 0x02
	featWakeEnabled               = 0x04
	featAlwaysOnEnabled           = 0x08
)

// Scaling factors for sensor data
// These are derived from the Q-point encoding in the SH-2 specification
const (
	scaleQuat        = 1.0 / 16384.0   // Q14
	scaleAccel       = 1.0 / 256.0     // Q8
	scaleGyro        = 1.0 / 512.0     // Q9
	scaleMag         = 1.0 / 16.0      // Q4
	scaleAccuracy    = 1.0 / 4096.0    // Q12
	scalePressure    = 1.0 / 1048576.0 // Q20
	scaleLight       = 1.0 / 256.0     // Q8
	scaleHumidity    = 1.0 / 256.0     // Q8
	scaleProximity   = 1.0 / 16.0      // Q4
	scaleTemperature = 1.0 / 128.0     // Q7
	scaleAngle       = 1.0 / 16.0      // Q4
	scaleHeartRate   = 1.0 / 16.0      // Q4
)

// Activity classifier codes (extended beyond standard SH-2)
const (
	ActivityUnknown     = 0
	ActivityInVehicle   = 1
	ActivityOnBicycle   = 2
	ActivityOnFoot      = 3
	ActivityStill       = 4
	ActivityTilting     = 5
	ActivityWalking     = 6
	ActivityRunning     = 7
	ActivityOnStairs    = 8
	ActivityOptionCount = 9
)

// Stability classifier values
const (
	StabilityUnknown    = 0
	StabilityOnTable    = 1
	StabilityStationary = 2
	StabilityStable     = 3
	StabilityMotion     = 4
)

// Tap detector flags
const (
	TapX      = 0x01 // 1 - X axis tapped
	TapXPos   = 0x02 // 2 - X positive direction
	TapY      = 0x04 // 4 - Y axis tapped
	TapYPos   = 0x08 // 8 - Y positive direction
	TapZ      = 0x10 // 16 - Z axis tapped
	TapZPos   = 0x20 // 32 - Z positive direction
	TapDouble = 0x40 // 64 - Double tap occurred
)

// GUID values for SHTP
const (
	guidSHTP       = 0
	guidExecutable = 1
	guidSensorHub  = 2
)

// Advertisement tags
const (
	tagNull                = 0
	tagGUID                = 1
	tagMaxCargoHeaderWrite = 2
	tagMaxCargoHeaderRead  = 3
	tagMaxTransferWrite    = 4
	tagMaxTransferRead     = 5
	tagNormalChannel       = 6
	tagWakeChannel         = 7
	tagAppName             = 8
	tagChannelName         = 9
	tagAdvCount            = 10
	tagAppSpecific         = 0x80
	tagSH2Version          = 0x80
	tagSH2ReportLengths    = 0x81
)

// Timeouts
const (
	advertTimeout  = 200000 // microseconds
	commandTimeout = 300000 // microseconds
)

// Executable device commands
const (
	execDeviceCmdReset = 1
	execDeviceCmdOn    = 2
	execDeviceCmdSleep = 3
)

// Executable device responses
const (
	execDeviceRespResetComplete = 1
)
