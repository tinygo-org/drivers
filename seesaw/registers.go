package seesaw

type ModuleBaseAddress byte

// Module Base Addreses
// The module base addresses for different seesaw modules.
const (
	ModuleStatusBase  ModuleBaseAddress = 0x00
	ModuleGpioBase    ModuleBaseAddress = 0x01
	ModuleSercom0Base ModuleBaseAddress = 0x02

	ModuleTimerBase     ModuleBaseAddress = 0x08
	ModuleAdcBase       ModuleBaseAddress = 0x09
	ModuleDacBase       ModuleBaseAddress = 0x0A
	ModuleInterruptBase ModuleBaseAddress = 0x0B
	ModuleDapBase       ModuleBaseAddress = 0x0C
	ModuleEepromBase    ModuleBaseAddress = 0x0D
	ModuleNeoPixelBase  ModuleBaseAddress = 0x0E
	ModuleTouchBase     ModuleBaseAddress = 0x0F
	ModuleKeypadBase    ModuleBaseAddress = 0x10
	ModuleEncoderBase   ModuleBaseAddress = 0x11
	ModuleSpectrumBase  ModuleBaseAddress = 0x12
)

type FunctionAddress byte

// GPIO module function address registers
const (
	FunctionGpioDirsetBulk FunctionAddress = 0x02
	FunctionGpioDirclrBulk FunctionAddress = 0x03
	FunctionGpioBulk       FunctionAddress = 0x04
	FunctionGpioBulkSet    FunctionAddress = 0x05
	FunctionGpioBulkClr    FunctionAddress = 0x06
	FunctionGpioBulkToggle FunctionAddress = 0x07
	FunctionGpioIntenset   FunctionAddress = 0x08
	FunctionGpioIntenclr   FunctionAddress = 0x09
	FunctionGpioIntflag    FunctionAddress = 0x0A
	FunctionGpioPullenset  FunctionAddress = 0x0B
	FunctionGpioPullenclr  FunctionAddress = 0x0C
)

// status module function address registers
const (
	FunctionStatusHwId    FunctionAddress = 0x01
	FunctionStatusVersion FunctionAddress = 0x02
	FunctionStatusOptions FunctionAddress = 0x03
	FunctionStatusTemp    FunctionAddress = 0x04
	FunctionStatusSwrst   FunctionAddress = 0x7F
)

// timer module function address registers
const (
	FunctionTimerStatus FunctionAddress = 0x00
	FunctionTimerPwm    FunctionAddress = 0x01
	FunctionTimerFreq   FunctionAddress = 0x02
)

// ADC module function address registers
const (
	FunctionAdcStatus        FunctionAddress = 0x00
	FunctionAdcInten         FunctionAddress = 0x02
	FunctionAdcIntenclr      FunctionAddress = 0x03
	FunctionAdcWinmode       FunctionAddress = 0x04
	FunctionAdcWinthresh     FunctionAddress = 0x05
	FunctionAdcChannelOffset FunctionAddress = 0x07
)

// Sercom module function address registers
const (
	FunctionSercomStatus   FunctionAddress = 0x00
	FunctionSercomInten    FunctionAddress = 0x02
	FunctionSercomIntenclr FunctionAddress = 0x03
	FunctionSercomBaud     FunctionAddress = 0x04
	FunctionSercomData     FunctionAddress = 0x05
)

// neopixel module function address registers
const (
	FunctionNeopixelStatus    FunctionAddress = 0x00
	FunctionNeopixelPin       FunctionAddress = 0x01
	FunctionNeopixelSpeed     FunctionAddress = 0x02
	FunctionNeopixelBufLength FunctionAddress = 0x03
	FunctionNeopixelBuf       FunctionAddress = 0x04
	FunctionNeopixelShow      FunctionAddress = 0x05
)

// touch module function address registers
const (
	FunctionTouchChannelOffset FunctionAddress = 0x10
)

// keypad module function address registers
const (
	FunctionKeypadStatus   FunctionAddress = 0x00
	FunctionKeypadEvent    FunctionAddress = 0x01
	FunctionKeypadIntenset FunctionAddress = 0x02
	FunctionKeypadIntenclr FunctionAddress = 0x03
	FunctionKeypadCount    FunctionAddress = 0x04
	FunctionKeypadFifo     FunctionAddress = 0x10
)
