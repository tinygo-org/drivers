package bme68x

type Option func(*Device)

// WithAddress sets the I2C/SPI address of the device.
func WithAddress(addr uint16) Option {
	return func(d *Device) {
		d.address = addr
	}
}

// WithPeriodPoll sets the polling period.
func WithPeriodPoll(period uint32) Option {
	return func(d *Device) {
		d.config.PeriodPoll = period
	}
}

// WithMode sets the mode of the device.
func WithMode(mode Mode) Option {
	return func(d *Device) {
		d.config.mode = mode
	}
}

// WithIIRFilter sets the IIR filter coefficient.
func WithIIRFilter(filter FilterCoefficient) Option {
	return func(d *Device) {
		d.config.IIR = filter
	}
}

// WithHumidityOversampling sets the humidity oversampling.
func WithHumidityOversampling(os Oversampling) Option {
	return func(d *Device) {
		d.config.Humidity = os
	}
}

// WithPressureOversampling sets the pressure oversampling.
func WithPressureOversampling(os Oversampling) Option {
	return func(d *Device) {
		d.config.Pressure = os
	}
}

// WithTemperatureOversampling sets the temperature oversampling.
func WithTemperatureOversampling(os Oversampling) Option {
	return func(d *Device) {
		d.config.Temperature = os
	}
}

// With HeatrTemperature sets the target heater temperature.
func WithHeatrTemperature(temp uint16) Option {
	return func(d *Device) {
		d.config.HeatrTemp = temp
	}
}

// WithHeatrDuration sets the target heater duration.
func WithHeatrDuration(duration uint16) Option {
	return func(d *Device) {
		d.config.HeatrDur = duration
	}
}

// WithAmbientTemperature sets the ambient temperature.
// The temperature in deg C is used for defining the heater temperature.
func WithAmbientTemperature(temp int8) Option {
	return func(d *Device) {
		d.config.AmbientTemperature = temp
	}
}
