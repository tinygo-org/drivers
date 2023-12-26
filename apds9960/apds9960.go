// Package apds9960 implements a driver for APDS-9960,
// a digital proximity, ambient light, RGB and gesture sensor.
//
// Datasheet: https://cdn.sparkfun.com/assets/learn_tutorials/3/2/1/Avago-APDS-9960-datasheet.pdf
package apds9960

import (
	"time"

	"tinygo.org/x/drivers"
)

// Device wraps an I2C connection to a APDS-9960 device.
type Device struct {
	bus     drivers.I2C
	_txerr  error
	gesture gestureData
	buf     [8]byte
	Address uint8
	mode    uint8
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

// for enabling various device functions.
type encfg uint8

// data := []byte{gen<<6 | pien<<5 | aien<<4 | wen<<3 | pen<<2 | aen<<1 | pon}
const (
	enPON encfg = 1 << iota
	enAEN
	enPEN
	enWEN
	enAIEN
	enPIEN
	enGEN
)

func (e encfg) write7bits(b []byte) {
	for i := uint8(0); i < 7; i++ {
		b[i] = byte(e>>(6-i)) & 1
	}
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
	d.txNew()
	return d.txRead8(APDS9960_ID_REG) == 0xAB && d.txErr() == nil
}

// GetMode returns current engine mode
func (d *Device) GetMode() uint8 {
	return d.mode
}

// DisableAll turns off the device and all functions
func (d *Device) DisableAll() error {
	err := d.enable(0)
	if err != nil {
		return err
	}
	d.txWrite8(APDS9960_GCONF4_REG, 0)
	err = d.txErr()
	if err == nil {
		d.mode = MODE_NONE
		d.gesture.detected = GESTURE_NONE
	}
	return err
}

// SetProximityPulse sets proximity pulse length (4, 8, 16, 32) and count (1..64)
// default: 16, 64
func (d *Device) SetProximityPulse(length, count uint8) error {
	d.txNew()
	d.txWrite8(APDS9960_PPULSE_REG, getPulseLength(length)<<6|getPulseCount(count))
	return d.txErr()
}

// SetGesturePulse sets gesture pulse length (4, 8, 16, 32) and count (1..64)
// default: 16, 64
func (d *Device) SetGesturePulse(length, count uint8) error {
	d.txNew()
	d.txWrite8(APDS9960_GPULSE_REG, getPulseLength(length)<<6|getPulseCount(count))
	return d.txErr()
}

// SetADCIntegrationCycles sets ALS/color ADC internal integration cycles (1..256, 1 cycle = 2.78 ms)
// default: 4 (approx. 10 ms)
func (d *Device) SetADCIntegrationCycles(cycles uint16) error {
	if cycles > 256 {
		cycles = 256
	}
	d.txNew()
	d.txWrite8(APDS9960_ATIME_REG, uint8(256-cycles))
	return d.txErr()
}

// SetGains sets proximity/gesture gain (1, 2, 4, 8x) and ALS/color gain (1, 4, 16, 64x)
// default: 1, 1, 4
func (d *Device) SetGains(proximityGain, gestureGain, colorGain uint8) error {
	d.txNew()
	d.txWrite8(APDS9960_CONTROL_REG, getProximityGain(proximityGain)<<2|getALSGain(colorGain))
	d.txWrite8(APDS9960_GCONF2_REG, getProximityGain(gestureGain)<<5)
	return d.txErr()
}

// LEDBoost sets proximity and gesture LED current level (100, 150, 200, 300 (%))
// default: 100
func (d *Device) LEDBoost(percent uint16) error {
	var v uint8
	switch {
	case percent < 125:
		v = 0
	case percent < 175:
		v = 1
	case percent < 250:
		v = 2
	default:
		v = 3 // Maximum case.
	}
	d.txNew()
	d.txWrite8(APDS9960_CONFIG2_REG, 0x01|(v<<4))
	return d.txErr()
}

// Setthreshold sets threshold (0..255) for detecting gestures
// default: 30
func (d *Device) Setthreshold(t uint8) {
	d.gesture.threshold = t
}

// Setsensitivity sets sensivity (0..100) for detecting gestures
// default: 20
func (d *Device) Setsensitivity(s uint8) {
	if s > 100 {
		s = 100
	}
	d.gesture.sensitivity = 100 - s
}

// EnableProximity starts the proximity engine
func (d *Device) EnableProximity() error {
	if d.mode != MODE_NONE {
		err := d.DisableAll()
		if err != nil {
			return err
		}
	}
	err := d.enable(enPON | enPEN | enWEN)
	if err == nil {
		d.mode = MODE_PROXIMITY
	}
	return err
}

// Err returns the current error state of the device if encountered during I2C communication.
// After a call to Err the error is cleared.
func (d *Device) Err() error {
	err := d.txErr()
	d.txNew()
	return err
}

// ProximityAvailable reports if proximity data is available
func (d *Device) ProximityAvailable() bool {
	if d.mode != MODE_PROXIMITY {
		return false
	}
	status, err := d.ReadStatus()
	return err == nil && status.PVALID()
}

// ReadProximity reads proximity data (0..255)
func (d *Device) ReadProximity() (proximity int32) {
	if d.mode != MODE_PROXIMITY {
		return 0
	}
	d.txNew()
	val := d.txRead8(APDS9960_PDATA_REG)
	return 255 - int32(val)
}

// EnableColor starts the color engine
func (d *Device) EnableColor() (err error) {
	if d.mode != MODE_NONE {
		err = d.DisableAll()
		if err != nil {
			return err
		}
	}
	err = d.enable(enPON | enAEN | enWEN)
	if err == nil {
		d.mode = MODE_COLOR
	}
	return err
}

// ColorAvailable reports if color data is available
func (d *Device) ColorAvailable() bool {
	if d.mode != MODE_COLOR {
		return false
	}
	status, err := d.ReadStatus()
	return err == nil && status.AVALID()
}

// ReadColor reads color data (red, green, blue, clear color/brightness)
func (d *Device) ReadColor() (r int32, g int32, b int32, clear int32) {
	if d.mode != MODE_COLOR {
		return
	}
	d.txNew()
	data := d.buf[:8]
	const numLowRegs = APDS9960_GDATAH_REG - APDS9960_CDATAL_REG + 1
	for i := uint8(0); i < numLowRegs; i++ {
		data[i] = d.txRead8(i + APDS9960_CDATAL_REG)
	}
	data[numLowRegs] = d.txRead8(APDS9960_BDATAL_REG)
	data[numLowRegs+1] = d.txRead8(APDS9960_BDATAH_REG)
	if d.txErr() != nil {
		return
	}
	clear = int32(uint16(data[1])<<8 | uint16(data[0]))
	r = int32(uint16(data[3])<<8 | uint16(data[2]))
	g = int32(uint16(data[5])<<8 | uint16(data[4]))
	b = int32(uint16(data[7])<<8 | uint16(data[6]))
	return r, g, b, clear
}

// EnableGesture starts the gesture engine
func (d *Device) EnableGesture() error {
	if d.mode != MODE_NONE {
		err := d.DisableAll()
		if err != nil {
			return err
		}
	}
	err := d.enable(enPON | enPEN | enGEN | enWEN)
	if err != nil {
		return err
	}
	d.mode = MODE_GESTURE
	d.gesture.detected = GESTURE_NONE
	d.gesture.gXDelta = 0
	d.gesture.gYDelta = 0
	d.gesture.gXPrevDelta = 0
	d.gesture.gYPrevDelta = 0
	d.gesture.received = false
	return nil
}

// GestureAvailable reports if gesture data is available
func (d *Device) GestureAvailable() bool {
	if d.mode != MODE_GESTURE {
		return false
	}
	d.txNew()
	gstatus := d.txRead8(APDS9960_GSTATUS_REG)
	if gstatus&1 == 0 {
		return false
	}
	availableDataSets := d.txRead8(APDS9960_GFLVL_REG)
	if availableDataSets == 0 {
		return false
	}
	data := d.buf[:]
	// read up, down, left and right proximity data from FIFO
	var dataSets [32][4]uint8
	const numAddrs = APDS9960_GFIFO_R_REG - APDS9960_GFIFO_U_REG + 1
	for i := uint8(0); i < availableDataSets; i++ {
		for j := uint8(0); j < numAddrs; j++ {
			data[j] = d.txRead8(j + APDS9960_GFIFO_U_REG)
		}
		if d.txErr() != nil {
			return false
		}
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

func (d *Device) configureDevice(cfg Configuration) error {
	err := d.DisableAll() // turn off everything
	if err != nil {
		return err
	}
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

	err = d.SetProximityPulse(cfg.ProximityPulseLength, cfg.ProximityPulseCount)
	if err != nil {
		return err
	}
	err = d.SetGesturePulse(cfg.GesturePulseLength, cfg.GesturePulseCount)
	if err != nil {
		return err
	}
	err = d.SetGains(cfg.ProximityGain, cfg.GestureGain, cfg.ColorGain)
	if err != nil {
		return err
	}
	err = d.SetADCIntegrationCycles(cfg.ADCIntegrationCycles)
	if err == nil && cfg.LEDBoost > 0 {
		err = d.LEDBoost(cfg.LEDBoost)
	}
	return err
}

func (d *Device) enable(cfg encfg) error {
	d.txNew()
	cfg.write7bits(d.buf[:7])
	d.txWrite(APDS9960_ENABLE_REG, d.buf[:7])
	err := d.txErr()
	if err == nil && cfg&enPON != 0 {
		time.Sleep(time.Millisecond * 10)
	}
	return err
}

func (d *Device) txErr() error { return d._txerr }

func (d *Device) txNew() { d._txerr = nil }

func (d *Device) txRead8(addr uint8) uint8 {
	if d._txerr != nil {
		return 0
	}
	d.buf[0] = addr
	d._txerr = d.bus.Tx(uint16(d.Address), d.buf[:1], d.buf[1:2])
	return d.buf[1]
}

func (d *Device) txWrite8(addr uint8, val uint8) {
	if d._txerr != nil {
		return
	}
	d.buf[0] = addr
	d.buf[1] = val
	d._txerr = d.bus.Tx(uint16(d.Address), d.buf[:2], nil)
}

func (d *Device) txWrite(addr uint8, data []byte) {
	if d._txerr != nil {
		return
	} else if len(data) > len(d.buf)-1 {
		panic("txWrite: data too long")
	}
	d.buf[0] = addr
	copy(d.buf[1:], data)
	d._txerr = d.bus.Tx(uint16(d.Address), d.buf[:len(data)+1], nil)
}

type status uint8

const (
	statusAVALID status = 1 << iota
	statusPVALID
	_
	_
	statusAINT
	statusPINT
	statusPGSAT
	statusCPSAT
)

func (s status) CPSAT() bool  { return s&statusCPSAT != 0 }
func (s status) PGSAT() bool  { return s&statusPGSAT != 0 }
func (s status) PINT() bool   { return s&statusPINT != 0 }
func (s status) AINT() bool   { return s&statusAINT != 0 }
func (s status) PVALID() bool { return s&statusPVALID != 0 }
func (s status) AVALID() bool { return s&statusAVALID != 0 }

func (d *Device) ReadStatus() (status, error) {
	d.txNew()
	return status(d.txRead8(APDS9960_STATUS_REG)), d.txErr()
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
