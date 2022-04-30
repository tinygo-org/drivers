package scd4x

const (
	// Address is default I2C address.
	Address = 0x62

	CmdDataReady                        = 0xE4B8
	CmdFactoryReset                     = 0x3632
	CmdForcedRecal                      = 0x362F
	CmdGetAltitude                      = 0x2322
	CmdGetASCE                          = 0x2313
	CmdGetTempOffset                    = 0x2318
	CmdPersistSettings                  = 0x3615
	CmdReadMeasurement                  = 0xEC05
	CmdReinit                           = 0x3646
	CmdSelfTest                         = 0x3639
	CmdSerialNumber                     = 0x3682
	CmdSetAltitude                      = 0x2427
	CmdSetASCE                          = 0x2416
	CmdSetPressure                      = 0xE000
	CmdSetTempOffset                    = 0x241D
	CmdStartLowPowerPeriodicMeasurement = 0x21AC
	CmdStartPeriodicMeasurement         = 0x21B1
	CmdStopPeriodicMeasurement          = 0x3F86
)
