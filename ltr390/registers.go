package ltr390

// DefaultAddress is the I2C address of the device. Datasheet section 5.1 V.
const DefaultAddress = 0x53

// Registers. Datasheet section 6.
const (
	MainCtrl   = 0x00
	MeasRate   = 0x04
	GainRange  = 0x05
	PartID     = 0x06
	MainStatus = 0x07
	ALSData    = 0x0D
	UVSData    = 0x10
	IntCfg     = 0x19
	IntPst     = 0x1A
	ThresUp    = 0x21
	ThresLow   = 0x24
)

// Register fields. Datasheet sections 6.1 to 6.8.
const (
	ctrlEnable  = 1 << 1
	ctrlUVSMode = 1 << 3

	statusDataReady = 1 << 3

	intCfgEnable = 1 << 2
	intCfgALS    = 0x1 << 4
	intCfgUVS    = 0x3 << 4

	partNumberID = 0xB

	dataMask = 0xFFFFF
)

// Mode selects which channel is measured. The sensor measures one channel at a time.
type Mode uint8

const (
	ModeALS Mode = iota
	ModeUVS
)

// Gain is the analog gain applied to both channels.
type Gain uint8

const (
	Gain1 Gain = iota + 1
	Gain3
	Gain6
	Gain9
	Gain18
)

// Resolution is the ADC resolution. Higher resolution means a longer conversion time.
type Resolution uint8

const (
	Resolution20Bit Resolution = iota + 1 // 400ms
	Resolution19Bit                       // 200ms
	Resolution18Bit                       // 100ms
	Resolution17Bit                       // 50ms
	Resolution16Bit                       // 25ms
	Resolution13Bit                       // 12.5ms
)

// MeasurementRate is the time between measurements. The sensor uses the
// conversion time instead when it is longer.
type MeasurementRate uint8

const (
	Rate25ms MeasurementRate = iota + 1
	Rate50ms
	Rate100ms
	Rate200ms
	Rate500ms
	Rate1000ms
	Rate2000ms
)

// Config holds the initial settings. Zero values use the power-on defaults.
type Config struct {
	Gain       Gain
	Resolution Resolution
	Rate       MeasurementRate
	Mode       Mode
}

// Power-on defaults. Datasheet sections 6.2 and 6.3.
const (
	defaultGain       = Gain3
	defaultResolution = Resolution18Bit
	defaultRate       = Rate100ms
)
