// IMU or inertial measurement units are used to
// estimate the position and orientation of a body
// in 3D space (such as a drone or helicopter). These
// usually consist in an accelerometer and a gyroscope.

package drivers

// IMU represents an intertial measurement unit such as
// the MPU6050 with acceleration and angular velocity measurement
// capabilities.
type IMU interface {
	// Acceleration returns sensor accelerations in micro gravities
	// with respect to the sensor's orientation.
	Acceleration() (ax, ay, az int32)
	// AngularVelocity returns sensor angular velocity in micro radians
	// with respect to the sensor's orientation.
	AngularVelocity() (gx, gy, gz int32)
}
