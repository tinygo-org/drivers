// Package apds9960 implements a driver for APDS-9960,
// a digital proximity, ambient light, RGB and gesture sensor.
//
// Datasheet: https://cdn.sparkfun.com/assets/learn_tutorials/3/2/1/Avago-APDS-9960-datasheet.pdf
package apds9960

import (
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// Device wraps an I2C connection to a APDS-9960 device.
type Device struct {
	bus     drivers.I2C
	Address uint8
	mode    uint8
	gesture gestureData
}

// Configuration for APDS-9960 device.
type Configuration struct {
	ProximityPulseLength uint8
	ProximityPulseCount  uint8
	GesturePulseLength   uint8
	GesturePulseCount    uint8
	ProximityGain        uint8
	GestureGain          uint8
	ColorGain            uint8
	ADCIntegrationCycles uint16
	LEDBoost             uint16
	threshold            uint8
	sensitivity          uint8
}

// for gesture-related data
type gestureData struct {
	detected    uint8
	threshold   uint8
	sensitivity uint8
	gXDelta     int16
	gYDelta     int16
	gXPrevDelta int16
	gYPrevDelta int16
	received    bool
}

// for enabling various device function
type enableConfig struct {
	GEN  bool
	PIEN bool
	AIEN bool
	WEN  bool
	PEN  bool
	AEN  bool
	PON  bool
}

// New creates a new APDS-9960 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{bus: bus, Address: ADPS9960_ADDRESS, mode: MODE_NONE}
}

// Connected returns whether APDS-9960 has been found.
// It does a "who am I" request and checks the response.
func (d *Device) Connected() bool {
	data := []byte{0}
	legacy.ReadRegister(d.bus, d.Address, APDS9960_ID_REG, data)
	return data[0] == 0xAB
}

// GetMode returns current engine mode
func (d *Device) GetMode() uint8 {
	return d.mode
}

// DisableAll turns off the device and all functions
func (d *Device) DisableAll() {
	d.enable(enableConfig{})
	legacy.WriteRegister(d.bus, d.Address, APDS9960_GCONF4_REG, []byte{0x00})
	d.mode = MODE_NONE
	d.gesture.detected = GESTURE_NONE
}

// SetProximityPulse sets proximity pulse length (4, 8, 16, 32) and count (1~64)
// default: 16, 64
func (d *Device) SetProximityPulse(length, count uint8) {
	legacy.WriteRegister(d.bus, d.Address, APDS9960_PPULSE_REG, []byte{getPulseLength(length)<<6 | getPulseCount(count)})
}

// SetGesturePulse sets gesture pulse length (4, 8, 16, 32) and count (1~64)
// default: 16, 64
func (d *Device) SetGesturePulse(length, count uint8) {
	legacy.WriteRegister(d.bus, d.Address, APDS9960_GPULSE_REG, []byte{getPulseLength(length)<<6 | getPulseCount(count)})
}

// SetADCIntegrationCycles sets ALS/color ADC internal integration cycles (1~256, 1 cycle = 2.78 ms)
// default: 4 (~10 ms)
func (d *Device) SetADCIntegrationCycles(cycles uint16) {
	if cycles > 256 {
		cycles = 256
	}
	legacy.WriteRegister(d.bus, d.Address, APDS9960_ATIME_REG, []byte{uint8(256 - cycles)})
}

// SetGains sets proximity/gesture gain (1, 2, 4, 8x) and ALS/color gain (1, 4, 16, 64x)
// default: 1, 1, 4
func (d *Device) SetGains(proximityGain, gestureGain, colorGain uint8) {
	legacy.WriteRegister(d.bus, d.Address, APDS9960_CONTROL_REG, []byte{getProximityGain(proximityGain)<<2 | getALSGain(colorGain)})
	legacy.WriteRegister(d.bus, d.Address, APDS9960_GCONF2_REG, []byte{getProximityGain(gestureGain) << 5})
}

// LEDBoost sets proximity and gesture LED current level (100, 150, 200, 300 (%))
// default: 100
func (d *Device) LEDBoost(percent uint16) {
	var v uint8
	switch percent {
	case 100:
		v = 0
	case 150:
		v = 1
	case 200:
		v = 2
	case 300:
		v = 3
	}
	legacy.WriteRegister(d.bus, d.Address, APDS9960_CONFIG2_REG, []byte{0x01 | v<<4})
}

// Setthreshold sets threshold (0~255) for detecting gestures
// default: 30
func (d *Device) Setthreshold(t uint8) {
	d.gesture.threshold = t
}

// Setsensitivity sets sensivity (0~100) for detecting gestures
// default: 20
func (d *Device) Setsensitivity(s uint8) {
	if s > 100 {
		s = 100
	}
	d.gesture.sensitivity = 100 - s
}

// EnableProximity starts the proximity engine
func (d *Device) EnableProximity() {
	if d.mode != MODE_NONE {
		d.DisableAll()
	}
	d.enable(enableConfig{PON: true, PEN: true, WEN: true})
	d.mode = MODE_PROXIMITY
}

// ProximityAvailable reports if proximity data is available
func (d *Device) ProximityAvailable() bool {
	if d.mode == MODE_PROXIMITY && d.readStatus("PVALID") {
		return true
	}
	return false
}

// ReadProximity reads proximity data (0~255)
func (d *Device) ReadProximity() (proximity int32) {
	if d.mode != MODE_PROXIMITY {
		return 0
	}
	data := []byte{0}
	legacy.ReadRegister(d.bus, d.Address, APDS9960_PDATA_REG, data)
	return 255 - int32(data[0])
}

// EnableColor starts the color engine
func (d *Device) EnableColor() {
	if d.mode != MODE_NONE {
		d.DisableAll()
	}
	d.enable(enableConfig{PON: true, AEN: true, WEN: true})
	d.mode = MODE_COLOR
}

// ColorAvailable reports if color data is available
func (d *Device) ColorAvailable() bool {
	if d.mode == MODE_COLOR && d.readStatus("AVALID") {
		return true
	}
	return false
}

// ReadColor reads color data (red, green, blue, clear color/brightness)
func (d *Device) ReadColor() (r int32, g int32, b int32, clear int32) {
	if d.mode != MODE_COLOR {
		return
	}
	data := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	legacy.ReadRegister(d.bus, d.Address, APDS9960_CDATAL_REG, data[:1])
	legacy.ReadRegister(d.bus, d.Address, APDS9960_CDATAH_REG, data[1:2])
	legacy.ReadRegister(d.bus, d.Address, APDS9960_RDATAL_REG, data[2:3])
	legacy.ReadRegister(d.bus, d.Address, APDS9960_RDATAH_REG, data[3:4])
	legacy.ReadRegister(d.bus, d.Address, APDS9960_GDATAL_REG, data[4:5])
	legacy.ReadRegister(d.bus, d.Address, APDS9960_GDATAH_REG, data[5:6])
	legacy.ReadRegister(d.bus, d.Address, APDS9960_BDATAL_REG, data[6:7])
	legacy.ReadRegister(d.bus, d.Address, APDS9960_BDATAH_REG, data[7:])
	clear = int32(uint16(data[1])<<8 | uint16(data[0]))
	r = int32(uint16(data[3])<<8 | uint16(data[2]))
	g = int32(uint16(data[5])<<8 | uint16(data[4]))
	b = int32(uint16(data[7])<<8 | uint16(data[6]))
	return
}

// EnableGesture starts the gesture engine
func (d *Device) EnableGesture() {
	if d.mode != MODE_NONE {
		d.DisableAll()
	}
	d.enable(enableConfig{PON: true, PEN: true, GEN: true, WEN: true})
	d.mode = MODE_GESTURE
	d.gesture.detected = GESTURE_NONE
	d.gesture.gXDelta = 0
	d.gesture.gYDelta = 0
	d.gesture.gXPrevDelta = 0
	d.gesture.gYPrevDelta = 0
	d.gesture.received = false
}

// GestureAvailable reports if gesture data is available
func (d *Device) GestureAvailable() bool {
	if d.mode != MODE_GESTURE {
		return false
	}

	data := []byte{0, 0, 0, 0}

	// check GVALID
	legacy.ReadRegister(d.bus, d.Address, APDS9960_GSTATUS_REG, data[:1])
	if data[0]&0x01 == 0 {
		return false
	}

	// get number of data sets available in FIFO
	legacy.ReadRegister(d.bus, d.Address, APDS9960_GFLVL_REG, data[:1])
	availableDataSets := data[0]
	if availableDataSets == 0 {
		return false
	}

	// read up, down, left and right proximity data from FIFO
	var dataSets [32][4]uint8
	for i := uint8(0); i < availableDataSets; i++ {
		legacy.ReadRegister(d.bus, d.Address, APDS9960_GFIFO_U_REG, data[:1])
		legacy.ReadRegister(d.bus, d.Address, APDS9960_GFIFO_D_REG, data[1:2])
		legacy.ReadRegister(d.bus, d.Address, APDS9960_GFIFO_L_REG, data[2:3])
		legacy.ReadRegister(d.bus, d.Address, APDS9960_GFIFO_R_REG, data[3:4])
		for j := uint8(0); j < 4; j++ {
			dataSets[i][j] = data[j]
		}
	}

	// gesture detection process
	d.gesture.detected = GESTURE_NONE
	for i := uint8(0); i < availableDataSets; i++ {
		U := dataSets[i][0]
		D := dataSets[i][1]
		L := dataSets[i][2]
		R := dataSets[i][3]

		// if all readings fall below threshold, it's possible that
		// a movement's just been made
		if U < d.gesture.threshold && D < d.gesture.threshold && L < d.gesture.threshold && R < d.gesture.threshold {
			d.gesture.received = true
			// if there were movement in the previous step (including the last data sets)
			if d.gesture.gXPrevDelta != 0 && d.gesture.gYPrevDelta != 0 {
				totalX := d.gesture.gXPrevDelta - d.gesture.gXDelta
				totalY := d.gesture.gYPrevDelta - d.gesture.gYDelta
				// if previous and current movement are in opposite directions (pass through one led then next)
				// and the difference is big enough, the gesture is recorded
				switch {
				case totalX < -int16(d.gesture.sensitivity):
					d.gesture.detected = GESTURE_LEFT
				case totalX > int16(d.gesture.sensitivity):
					d.gesture.detected = GESTURE_RIGHT
				case totalY > int16(d.gesture.sensitivity):
					d.gesture.detected = GESTURE_DOWN
				case totalY < -int16(d.gesture.sensitivity):
					d.gesture.detected = GESTURE_UP
				}
				d.gesture.gXDelta = 0
				d.gesture.gYDelta = 0
				d.gesture.gXPrevDelta = 0
				d.gesture.gYPrevDelta = 0
			}
			continue
		}

		// recording current movement
		d.gesture.gXDelta = int16(R) - int16(L)
		d.gesture.gYDelta = int16(D) - int16(U)
		if d.gesture.received {
			d.gesture.received = false
			d.gesture.gXPrevDelta = d.gesture.gXDelta
			d.gesture.gYPrevDelta = d.gesture.gYDelta
		}
	}

	return d.gesture.detected != GESTURE_NONE
}

// ReadGesture reads last gesture data
func (d *Device) ReadGesture() (gesture int32) {
	return int32(d.gesture.detected)
}

// private functions

func (d *Device) configureDevice(cfg Configuration) {
	d.DisableAll() // turn off everything

	// "default" settings
	if cfg.ProximityPulseLength == 0 {
		cfg.ProximityPulseLength = 16
	}
	if cfg.ProximityPulseCount == 0 {
		cfg.ProximityPulseCount = 64
	}
	if cfg.GesturePulseLength == 0 {
		cfg.GesturePulseLength = 16
	}
	if cfg.GesturePulseCount == 0 {
		cfg.GesturePulseCount = 64
	}
	if cfg.ProximityGain == 0 {
		cfg.ProximityGain = 1
	}
	if cfg.GestureGain == 0 {
		cfg.GestureGain = 1
	}
	if cfg.ColorGain == 0 {
		cfg.ColorGain = 4
	}
	if cfg.ADCIntegrationCycles == 0 {
		cfg.ADCIntegrationCycles = 4
	}
	if cfg.threshold == 0 {
		d.gesture.threshold = 30
	}
	if cfg.sensitivity == 0 {
		d.gesture.sensitivity = 20
	}

	d.SetProximityPulse(cfg.ProximityPulseLength, cfg.ProximityPulseCount)
	d.SetGesturePulse(cfg.GesturePulseLength, cfg.GesturePulseCount)
	d.SetGains(cfg.ProximityGain, cfg.GestureGain, cfg.ColorGain)
	d.SetADCIntegrationCycles(cfg.ADCIntegrationCycles)

	if cfg.LEDBoost > 0 {
		d.LEDBoost(cfg.LEDBoost)
	}
}

func (d *Device) enable(cfg enableConfig) {
	var gen, pien, aien, wen, pen, aen, pon uint8

	if cfg.GEN {
		gen = 1
	}
	if cfg.PIEN {
		pien = 1
	}
	if cfg.AIEN {
		aien = 1
	}
	if cfg.WEN {
		wen = 1
	}
	if cfg.PEN {
		pen = 1
	}
	if cfg.AEN {
		aen = 1
	}
	if cfg.PON {
		pon = 1
	}

	data := []byte{gen<<6 | pien<<5 | aien<<4 | wen<<3 | pen<<2 | aen<<1 | pon}
	legacy.WriteRegister(d.bus, d.Address, APDS9960_ENABLE_REG, data)

	if cfg.PON {
		time.Sleep(time.Millisecond * 10)
	}
}

func (d *Device) readStatus(param string) bool {
	data := []byte{0}
	legacy.ReadRegister(d.bus, d.Address, APDS9960_STATUS_REG, data)

	switch param {
	case "CPSAT":
		return data[0]>>7&0x01 == 1
	case "PGSAT":
		return data[0]>>6&0x01 == 1
	case "PINT":
		return data[0]>>5&0x01 == 1
	case "AINT":
		return data[0]>>4&0x01 == 1
	case "PVALID":
		return data[0]>>1&0x01 == 1
	case "AVALID":
		return data[0]&0x01 == 1
	default:
		return false
	}
}

func getPulseLength(l uint8) uint8 {
	switch l {
	case 4:
		return 0
	case 8:
		return 1
	case 16:
		return 2
	case 32:
		return 3
	default:
		return 0
	}
}

func getPulseCount(c uint8) uint8 {
	if c < 1 && c > 64 {
		return 0
	}
	return c - 1
}

func getProximityGain(g uint8) uint8 {
	switch g {
	case 1:
		return 0
	case 2:
		return 1
	case 4:
		return 2
	case 8:
		return 3
	default:
		return 0
	}
}

func getALSGain(g uint8) uint8 {
	switch g {
	case 1:
		return 0
	case 4:
		return 1
	case 16:
		return 2
	case 64:
		return 3
	default:
		return 0
	}
}
