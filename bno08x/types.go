package bno08x

// SensorID identifies a specific sensor type.
type SensorID uint8

// Sensor IDs as defined in the SH-2 specification.
const (
	SensorRawAccelerometer           SensorID = 0x14
	SensorAccelerometer              SensorID = 0x01
	SensorLinearAcceleration         SensorID = 0x04
	SensorGravity                    SensorID = 0x06
	SensorRawGyroscope               SensorID = 0x15
	SensorGyroscope                  SensorID = 0x02
	SensorGyroscopeUncalibrated      SensorID = 0x07
	SensorRawMagnetometer            SensorID = 0x16
	SensorMagneticField              SensorID = 0x03
	SensorMagneticFieldUncalibrated  SensorID = 0x0F
	SensorRotationVector             SensorID = 0x05
	SensorGameRotationVector         SensorID = 0x08
	SensorGeomagneticRotationVector  SensorID = 0x09
	SensorPressure                   SensorID = 0x0A
	SensorAmbientLight               SensorID = 0x0B
	SensorHumidity                   SensorID = 0x0C
	SensorProximity                  SensorID = 0x0D
	SensorTemperature                SensorID = 0x0E
	SensorReserved                   SensorID = 0x17
	SensorTapDetector                SensorID = 0x10
	SensorStepDetector               SensorID = 0x18
	SensorStepCounter                SensorID = 0x11
	SensorSignificantMotion          SensorID = 0x12
	SensorStabilityClassifier        SensorID = 0x13
	SensorShakeDetector              SensorID = 0x19
	SensorFlipDetector               SensorID = 0x1A
	SensorPickupDetector             SensorID = 0x1B
	SensorStabilityDetector          SensorID = 0x1C
	SensorPersonalActivityClassifier SensorID = 0x1E
	SensorSleepDetector              SensorID = 0x1F
	SensorTiltDetector               SensorID = 0x20
	SensorPocketDetector             SensorID = 0x21
	SensorCircleDetector             SensorID = 0x22
	SensorHeartRateMonitor           SensorID = 0x23
	SensorARVRStabilizedRV           SensorID = 0x28
	SensorARVRStabilizedGRV          SensorID = 0x29
	SensorGyroIntegratedRV           SensorID = 0x2A
	SensorIZROMotionRequest          SensorID = 0x2B
	SensorMaxID                      SensorID = 0x2B
)

// ProductID contains firmware information from the sensor.
type ProductID struct {
	ResetCause   uint8
	VersionMajor uint8
	VersionMinor uint8
	PartNumber   uint32
	BuildNumber  uint32
	VersionPatch uint16
	Reserved0    uint8
	Reserved1    uint8
}

// ProductIDs holds all product ID entries returned by the sensor.
type ProductIDs struct {
	Entries    [5]ProductID
	NumEntries uint8
}

// Vector3 represents a 3D vector.
type Vector3 struct {
	X float32
	Y float32
	Z float32
}

// Quaternion represents a quaternion in (real, i, j, k) format.
// Note: This maps to (w, x, y, z) convention where w=real, x=i, y=j, z=k.
type Quaternion struct {
	Real float32
	I    float32
	J    float32
	K    float32
}

// RawVector3 contains raw ADC counts with timestamp.
type RawVector3 struct {
	X         int16
	Y         int16
	Z         int16
	Timestamp uint32
}

// RawGyroscope contains raw gyro readings with temperature and timestamp.
type RawGyroscope struct {
	X           int16
	Y           int16
	Z           int16
	Temperature int16
	Timestamp   uint32
}

// GyroscopeUncalibrated contains uncalibrated gyroscope data with bias.
type GyroscopeUncalibrated struct {
	X     float32
	Y     float32
	Z     float32
	BiasX float32
	BiasY float32
	BiasZ float32
}

// MagneticFieldUncalibrated contains uncalibrated magnetometer data with bias.
type MagneticFieldUncalibrated struct {
	X     float32
	Y     float32
	Z     float32
	BiasX float32
	BiasY float32
	BiasZ float32
}

// TapDetector contains tap/double-tap detection flags.
type TapDetector struct {
	Flags uint8
}

// StepDetector contains step detection with latency.
type StepDetector struct {
	Latency uint32
}

// StepCounter contains step count with latency.
type StepCounter struct {
	Count   uint16
	Latency uint32
}

// SignificantMotion indicates significant motion was detected.
type SignificantMotion struct {
	Motion uint16
}

// ActivityClassification contains activity classification data.
type ActivityClassification struct {
	Page            uint8
	MostLikelyState uint8
	Classification  [10]uint8
	EndOfPage       uint8
}

// ShakeDetector contains shake detection data.
type ShakeDetector struct {
	Shake uint16
}

// StabilityClassifier contains stability classification.
type StabilityClassifier struct {
	Classification uint8
}

// PersonalActivityClassifier contains personal activity data.
type PersonalActivityClassifier struct {
	Page            uint8
	MostLikelyState uint8
	Confidence      [10]uint8
	EndOfPage       uint8
}

// SensorValue contains decoded sensor data for all sensor types.
type SensorValue struct {
	id        SensorID
	status    uint8
	sequence  uint8
	delay     uint8
	timestamp uint64

	// Orientation data (quaternions)
	quaternion         Quaternion
	quaternionAccuracy float32

	// Linear measurements
	accelerometer      Vector3
	linearAcceleration Vector3
	gravity            Vector3
	gyroscope          Vector3
	gyroscopeUncal     GyroscopeUncalibrated
	magneticField      Vector3
	magneticFieldUncal MagneticFieldUncalibrated

	// Raw sensor data
	rawAccelerometer RawVector3
	rawGyroscope     RawGyroscope
	rawMagnetometer  RawVector3

	// Environmental sensors
	pressure     float32 // hPa
	ambientLight float32 // lux
	humidity     float32 // %
	proximity    float32 // cm
	temperature  float32 // °C

	// Activity detection
	tapDetector                TapDetector
	stepCounter                StepCounter
	stepDetector               StepDetector
	significantMotion          SignificantMotion
	shakeDetector              ShakeDetector
	flipDetector               uint16
	stabilityClassifier        StabilityClassifier
	stabilityDetector          uint8
	activityClassifier         ActivityClassification
	personalActivityClassifier PersonalActivityClassifier
	sleepDetector              uint8
	tiltDetector               uint8
	pocketDetector             uint8
	circleDetector             uint8
	heartRateMonitor           uint16
}

// SensorConfig holds configuration settings for a sensor.
type SensorConfig struct {
	ChangeSensitivityEnabled  bool
	ChangeSensitivityRelative bool
	WakeupEnabled             bool
	AlwaysOnEnabled           bool
	ChangeSensitivity         uint16
	ReportInterval            uint32 // microseconds
	BatchInterval             uint32 // microseconds
	SensorSpecific            uint32
}

// Error represents a driver error.
type Error string

func (e Error) Error() string { return string(e) }

// Error constants.
var (
	errBufferTooSmall = Error("bno08x: buffer too small")
	errNoEvent        = Error("bno08x: no sensor event available")
	errTimeout        = Error("bno08x: operation timed out")
	errFrameTooLarge  = Error("bno08x: frame exceeds maximum size")
	errNoBus          = Error("bno08x: I2C bus not configured")
	errInvalidParam   = Error("bno08x: invalid parameter")
	errHubError       = Error("bno08x: sensor hub error")
	errIO             = Error("bno08x: I/O error")
)

// Metadata accessor methods (always available for any sensor type)

// ID returns the sensor ID.
func (sv SensorValue) ID() SensorID {
	return sv.id
}

// Status returns the sensor status flags.
func (sv SensorValue) Status() uint8 {
	return sv.status
}

// Sequence returns the sequence number.
func (sv SensorValue) Sequence() uint8 {
	return sv.sequence
}

// Delay returns the sensor delay value.
func (sv SensorValue) Delay() uint8 {
	return sv.delay
}

// Timestamp returns the sensor timestamp.
func (sv SensorValue) Timestamp() uint64 {
	return sv.timestamp
}

// Orientation data accessor methods

// Quaternion returns the quaternion value for rotation vector sensors.
// Panics if called on a sensor type that doesn't provide quaternion data.
func (sv SensorValue) Quaternion() Quaternion {
	switch sv.id {
	case SensorRotationVector, SensorGameRotationVector, SensorGeomagneticRotationVector,
		SensorARVRStabilizedRV, SensorARVRStabilizedGRV, SensorGyroIntegratedRV:
		return sv.quaternion
	default:
		panic("bno08x: Quaternion() called on non-rotation sensor type")
	}
}

// QuaternionAccuracy returns the quaternion accuracy estimate.
// Panics if called on a sensor type that doesn't provide quaternion accuracy.
func (sv SensorValue) QuaternionAccuracy() float32 {
	switch sv.id {
	case SensorRotationVector, SensorGeomagneticRotationVector, SensorARVRStabilizedRV:
		return sv.quaternionAccuracy
	default:
		panic("bno08x: QuaternionAccuracy() called on sensor type without accuracy data")
	}
}

// Linear measurement accessor methods

// Accelerometer returns the accelerometer vector.
// Panics if called on a sensor type other than SensorAccelerometer.
func (sv SensorValue) Accelerometer() Vector3 {
	if sv.id != SensorAccelerometer {
		panic("bno08x: Accelerometer() called on non-accelerometer sensor type")
	}
	return sv.accelerometer
}

// LinearAcceleration returns the linear acceleration vector.
// Panics if called on a sensor type other than SensorLinearAcceleration.
func (sv SensorValue) LinearAcceleration() Vector3 {
	if sv.id != SensorLinearAcceleration {
		panic("bno08x: LinearAcceleration() called on wrong sensor type")
	}
	return sv.linearAcceleration
}

// Gravity returns the gravity vector.
// Panics if called on a sensor type other than SensorGravity.
func (sv SensorValue) Gravity() Vector3 {
	if sv.id != SensorGravity {
		panic("bno08x: Gravity() called on non-gravity sensor type")
	}
	return sv.gravity
}

// Gyroscope returns the gyroscope vector.
// Panics if called on a sensor type other than SensorGyroscope.
func (sv SensorValue) Gyroscope() Vector3 {
	if sv.id != SensorGyroscope {
		panic("bno08x: Gyroscope() called on non-gyroscope sensor type")
	}
	return sv.gyroscope
}

// GyroscopeUncal returns the uncalibrated gyroscope data.
// Panics if called on a sensor type other than SensorGyroscopeUncalibrated.
func (sv SensorValue) GyroscopeUncal() GyroscopeUncalibrated {
	if sv.id != SensorGyroscopeUncalibrated {
		panic("bno08x: GyroscopeUncal() called on wrong sensor type")
	}
	return sv.gyroscopeUncal
}

// MagneticField returns the magnetic field vector.
// Panics if called on a sensor type other than SensorMagneticField.
func (sv SensorValue) MagneticField() Vector3 {
	if sv.id != SensorMagneticField {
		panic("bno08x: MagneticField() called on wrong sensor type")
	}
	return sv.magneticField
}

// MagneticFieldUncal returns the uncalibrated magnetic field data.
// Panics if called on a sensor type other than SensorMagneticFieldUncalibrated.
func (sv SensorValue) MagneticFieldUncal() MagneticFieldUncalibrated {
	if sv.id != SensorMagneticFieldUncalibrated {
		panic("bno08x: MagneticFieldUncal() called on wrong sensor type")
	}
	return sv.magneticFieldUncal
}

// Raw sensor data accessor methods

// RawAccelerometer returns the raw accelerometer data.
// Panics if called on a sensor type other than SensorRawAccelerometer.
func (sv SensorValue) RawAccelerometer() RawVector3 {
	if sv.id != SensorRawAccelerometer {
		panic("bno08x: RawAccelerometer() called on wrong sensor type")
	}
	return sv.rawAccelerometer
}

// RawGyroscope returns the raw gyroscope data.
// Panics if called on a sensor type other than SensorRawGyroscope.
func (sv SensorValue) RawGyroscope() RawGyroscope {
	if sv.id != SensorRawGyroscope {
		panic("bno08x: RawGyroscope() called on wrong sensor type")
	}
	return sv.rawGyroscope
}

// RawMagnetometer returns the raw magnetometer data.
// Panics if called on a sensor type other than SensorRawMagnetometer.
func (sv SensorValue) RawMagnetometer() RawVector3 {
	if sv.id != SensorRawMagnetometer {
		panic("bno08x: RawMagnetometer() called on wrong sensor type")
	}
	return sv.rawMagnetometer
}

// Environmental sensor accessor methods

// Pressure returns the pressure reading in hPa.
// Panics if called on a sensor type other than SensorPressure.
func (sv SensorValue) Pressure() float32 {
	if sv.id != SensorPressure {
		panic("bno08x: Pressure() called on non-pressure sensor type")
	}
	return sv.pressure
}

// AmbientLight returns the ambient light reading in lux.
// Panics if called on a sensor type other than SensorAmbientLight.
func (sv SensorValue) AmbientLight() float32 {
	if sv.id != SensorAmbientLight {
		panic("bno08x: AmbientLight() called on wrong sensor type")
	}
	return sv.ambientLight
}

// Humidity returns the humidity reading in percent.
// Panics if called on a sensor type other than SensorHumidity.
func (sv SensorValue) Humidity() float32 {
	if sv.id != SensorHumidity {
		panic("bno08x: Humidity() called on non-humidity sensor type")
	}
	return sv.humidity
}

// Proximity returns the proximity reading in cm.
// Panics if called on a sensor type other than SensorProximity.
func (sv SensorValue) Proximity() float32 {
	if sv.id != SensorProximity {
		panic("bno08x: Proximity() called on non-proximity sensor type")
	}
	return sv.proximity
}

// Temperature returns the temperature reading in °C.
// Panics if called on a sensor type other than SensorTemperature.
func (sv SensorValue) Temperature() float32 {
	if sv.id != SensorTemperature {
		panic("bno08x: Temperature() called on non-temperature sensor type")
	}
	return sv.temperature
}

// Activity detection accessor methods

// TapDetector returns the tap detector data.
// Panics if called on a sensor type other than SensorTapDetector.
func (sv SensorValue) TapDetector() TapDetector {
	if sv.id != SensorTapDetector {
		panic("bno08x: TapDetector() called on wrong sensor type")
	}
	return sv.tapDetector
}

// StepCounter returns the step counter value.
// Panics if called on a sensor type other than SensorStepCounter.
func (sv SensorValue) StepCounter() StepCounter {
	if sv.id != SensorStepCounter {
		panic("bno08x: StepCounter() called on wrong sensor type")
	}
	return sv.stepCounter
}

// StepDetector returns the step detector data.
// Panics if called on a sensor type other than SensorStepDetector.
func (sv SensorValue) StepDetector() StepDetector {
	if sv.id != SensorStepDetector {
		panic("bno08x: StepDetector() called on wrong sensor type")
	}
	return sv.stepDetector
}

// SignificantMotion returns the significant motion data.
// Panics if called on a sensor type other than SensorSignificantMotion.
func (sv SensorValue) SignificantMotion() SignificantMotion {
	if sv.id != SensorSignificantMotion {
		panic("bno08x: SignificantMotion() called on wrong sensor type")
	}
	return sv.significantMotion
}

// ShakeDetector returns the shake detector data.
// Panics if called on a sensor type other than SensorShakeDetector.
func (sv SensorValue) ShakeDetector() ShakeDetector {
	if sv.id != SensorShakeDetector {
		panic("bno08x: ShakeDetector() called on wrong sensor type")
	}
	return sv.shakeDetector
}

// FlipDetector returns the flip detector data.
// Panics if called on a sensor type other than SensorFlipDetector.
func (sv SensorValue) FlipDetector() uint16 {
	if sv.id != SensorFlipDetector {
		panic("bno08x: FlipDetector() called on wrong sensor type")
	}
	return sv.flipDetector
}

// StabilityClassifier returns the stability classifier data.
// Panics if called on a sensor type other than SensorStabilityClassifier.
func (sv SensorValue) StabilityClassifier() StabilityClassifier {
	if sv.id != SensorStabilityClassifier {
		panic("bno08x: StabilityClassifier() called on wrong sensor type")
	}
	return sv.stabilityClassifier
}

// StabilityDetector returns the stability detector value.
// Panics if called on a sensor type other than SensorStabilityDetector.
func (sv SensorValue) StabilityDetector() uint8 {
	if sv.id != SensorStabilityDetector {
		panic("bno08x: StabilityDetector() called on wrong sensor type")
	}
	return sv.stabilityDetector
}

// ActivityClassifier returns the activity classification data.
// Note: This field appears unused in decode.go, keeping for API compatibility.
func (sv SensorValue) ActivityClassifier() ActivityClassification {
	return sv.activityClassifier
}

// PersonalActivityClassifier returns the personal activity classifier data.
// Panics if called on a sensor type other than SensorPersonalActivityClassifier.
func (sv SensorValue) PersonalActivityClassifier() PersonalActivityClassifier {
	if sv.id != SensorPersonalActivityClassifier {
		panic("bno08x: PersonalActivityClassifier() called on wrong sensor type")
	}
	return sv.personalActivityClassifier
}

// SleepDetector returns the sleep detector value.
// Panics if called on a sensor type other than SensorSleepDetector.
func (sv SensorValue) SleepDetector() uint8 {
	if sv.id != SensorSleepDetector {
		panic("bno08x: SleepDetector() called on wrong sensor type")
	}
	return sv.sleepDetector
}

// TiltDetector returns the tilt detector value.
// Panics if called on a sensor type other than SensorTiltDetector.
func (sv SensorValue) TiltDetector() uint8 {
	if sv.id != SensorTiltDetector {
		panic("bno08x: TiltDetector() called on wrong sensor type")
	}
	return sv.tiltDetector
}

// PocketDetector returns the pocket detector value.
// Panics if called on a sensor type other than SensorPocketDetector.
func (sv SensorValue) PocketDetector() uint8 {
	if sv.id != SensorPocketDetector {
		panic("bno08x: PocketDetector() called on wrong sensor type")
	}
	return sv.pocketDetector
}

// CircleDetector returns the circle detector value.
// Panics if called on a sensor type other than SensorCircleDetector.
func (sv SensorValue) CircleDetector() uint8 {
	if sv.id != SensorCircleDetector {
		panic("bno08x: CircleDetector() called on wrong sensor type")
	}
	return sv.circleDetector
}

// HeartRateMonitor returns the heart rate monitor value.
// Panics if called on a sensor type other than SensorHeartRateMonitor.
func (sv SensorValue) HeartRateMonitor() uint16 {
	if sv.id != SensorHeartRateMonitor {
		panic("bno08x: HeartRateMonitor() called on wrong sensor type")
	}
	return sv.heartRateMonitor
}
