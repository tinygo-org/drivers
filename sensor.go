package drivers

// Measurement specifies a type of measurement,
// for example: temperature, acceleration, pressure.
type Measurement uint32

// Sensor measurements
const (
	Voltage Measurement = 1 << iota
	Temperature
	Humidity
	Pressure
	Distance
	Acceleration
	AngularVelocity
	MagneticField
	Luminosity
	Time

	// All Measurements is the OR of all Measurement defined in drivers package. This
	// may change over time. Let user-defined Measurement flags be the MSB bits of Measurement.
	AllMeasurements Measurement = 1<<iota - 1
)

// Sensor represents an object capable of making one
// or more measurements. A sensor will then have methods
// which read the last updated measurements.
//
// Many Sensors may be collected into
// one Sensor interface to synchronize measurements.
type Sensor interface {
	Update(which Measurement) error
}
