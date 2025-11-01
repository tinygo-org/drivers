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
	ID        SensorID
	Status    uint8
	Sequence  uint8
	Delay     uint8
	Timestamp uint64

	// Orientation data (quaternions)
	Quaternion         Quaternion
	QuaternionAccuracy float32

	// Linear measurements
	Accelerometer      Vector3
	LinearAcceleration Vector3
	Gravity            Vector3
	Gyroscope          Vector3
	GyroscopeUncal     GyroscopeUncalibrated
	MagneticField      Vector3
	MagneticFieldUncal MagneticFieldUncalibrated

	// Raw sensor data
	RawAccelerometer RawVector3
	RawGyroscope     RawGyroscope
	RawMagnetometer  RawVector3

	// Environmental sensors
	Pressure     float32 // hPa
	AmbientLight float32 // lux
	Humidity     float32 // %
	Proximity    float32 // cm
	Temperature  float32 // °C

	// Activity detection
	TapDetector                TapDetector
	StepCounter                uint32
	StepDetector               StepDetector
	SignificantMotion          SignificantMotion
	ShakeDetector              ShakeDetector
	StabilityClassifier        StabilityClassifier
	StabilityDetector          uint8
	ActivityClassifier         ActivityClassification
	PersonalActivityClassifier PersonalActivityClassifier
	SleepDetector              uint8
	TiltDetector               uint8
	PocketDetector             uint8
	CircleDetector             uint8
	HeartRateMonitor           uint16
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
