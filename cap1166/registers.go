package cap1166

// Register addresses for the Microchip CAP1166.
// Datasheet: https://ww1.microchip.com/downloads/en/DeviceDoc/00001621B.pdf
const (
	regMainControl     = 0x00
	regInputStatus     = 0x03
	regSensitivity     = 0x1F
	regGeneralConfig   = 0x20
	regInputEnable     = 0x21
	regSamplingConfig  = 0x24
	regInterruptEnable = 0x27
	regRepeatEnable    = 0x28
	regMultiTouchConf  = 0x2A
	regConfiguration2  = 0x44

	regLEDLinking    = 0x72
	regLEDPolarity   = 0x73
	regLEDOutputCon  = 0x74
	regLEDBehaviour1 = 0x81
	regLEDBehaviour2 = 0x82
	regLEDDirectDuty = 0x93
	regLEDDirectRamp = 0x94

	regProductID = 0xFD
)

// ProductID is the value the CAP1166 reports in regProductID.
const ProductID = 0x51

// Touch channel numbers, matching the labels silk-screened on the
// Display-o-Tron HAT.
const (
	Cancel = 0
	Up     = 1
	Down   = 2
	Left   = 3
	Button = 4
	Right  = 5
)

// NumInputs is the number of touch inputs and LED outputs on the CAP1166.
const NumInputs = 6
