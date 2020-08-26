// Package tester contains mock structs to make it easier to test I2C devices.
//
// TODO: info on how to use this.
//
package tester // import "tinygo.org/x/drivers/tester"

// Failer is used by the I2CDevice type to abort when it's used in
// unexpected ways, such as reading an out-of-range register.
type Failer interface {
	// Fatalf prints the Printf-formatted message and exits the current
	// goroutine.
	Fatalf(f string, a ...interface{})
}
