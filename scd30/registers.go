package scd30

const (
	// Address is the default and only I2C address of the SCD30.
	Address uint16 = 0x61

	commandStartContinuousMeasurement = 0x0010
	commandStopContinuousMeasurement  = 0x0104
	commandDataReady                  = 0x0202
	commandReadMeasurement            = 0x0300
	commandSetMeasurementInterval     = 0x4600
	commandSetAutoCalibration         = 0x5306
)

const (
	minimumMeasurementInterval = 2
	maximumMeasurementInterval = 1800
	minimumAmbientPressure     = 700
	maximumAmbientPressure     = 1400
)
