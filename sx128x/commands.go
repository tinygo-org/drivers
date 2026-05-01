package sx128x

const (
	// SX128X SPI commands
	cmdGetStatus = uint8(0xC0)

	// Register Access Operations
	cmdWriteRegister = uint8(0x18)
	cmdReadRegister  = uint8(0x19)

	// Data Buffer Operations
	cmdWriteBuffer = uint8(0x1A)
	cmdReadBuffer  = uint8(0x1B)

	// Radio Operation Modes
	cmdSetSleep              = uint8(0x84)
	cmdSetStandby            = uint8(0x80)
	cmdSetFS                 = uint8(0xC1)
	cmdSetTx                 = uint8(0x83)
	cmdSetRx                 = uint8(0x82)
	cmdSetRxDutyCycle        = uint8(0x94)
	cmdSetLongPreamble       = uint8(0x9B)
	cmdSetCAD                = uint8(0xC5)
	cmdSetTxContinuousWave   = uint8(0xD1)
	cmdSetContinuousPreamble = uint8(0xD2)
	cmdSetAutoTx             = uint8(0x98)
	cmdSetAutoFS             = uint8(0x9E)

	// Radio Configuration
	cmdSetPacketType        = uint8(0x8A)
	cmdGetPacketType        = uint8(0x03)
	cmdSetRFFrequency       = uint8(0x86)
	cmdSetTxParams          = uint8(0x8E)
	cmdSetCADParams         = uint8(0x88)
	cmdSetBufferBaseAddress = uint8(0x8F)
	cmdSetModulationParams  = uint8(0x8B)
	cmdSetPacketParams      = uint8(0x8C)

	// Communication Status Information
	cmdGetRxBufferStatus = uint8(0x17)
	cmdGetPacketStatus   = uint8(0x1D)
	cmdGetRSSIInst       = uint8(0x1F)

	// IRQ Handling
	cmdSetDIOIRQParams = uint8(0x8D)
	cmdGetIRQStatus    = uint8(0x15)
	cmdClearIRQStatus  = uint8(0x97)

	// Miscellaneous
	cmdSetRegulatorMode = uint8(0x96)
	cmdSetSaveContext   = uint8(0xD5)
)
