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
	// Gas or liquid concentration, usually measured in ppm (parts per million).
	Concentration
	// Add Measurements above AllMeasurements.

	// AllMeasurements is the OR of all Measurement values. It ensures all measurements are done.
	AllMeasurements Measurement = (1 << 32) - 1
)

// Sensor represents an object capable of making one
// or more measurements. A sensor will then have methods
// which read the last updated measurements.
//
// Many Sensors may be collected into
// one Sensor interface to synchronize measurements.
type Sensor interface {
	// Update performs IO to update the measurements of a sensor.
	// It shall return error only when the sensor encounters an error that prevents it from
	// storing all or part of the measurements it was called to do.
	Update(which Measurement) error
}
