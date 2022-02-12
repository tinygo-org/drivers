package drivers

// This file contains some common units that can be used in a sensor driver.

// Temperature is a temperature in Celsius milli degrees (°C/1000). For example,
// the value 25000 is 25°C.
type Temperature int32

// Celsius returns the temperature in degrees Celsius.
func (t Temperature) Celsius() float32 {
	return float32(t) / 1000
}

// Fahrenheit returns the temperature in degrees Fahrenheit.
func (t Temperature) Fahrenheit() float32 {
	return t.Celsius()*1.8 + 32
}
