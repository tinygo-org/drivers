package drivers

// PWMSetter is a typical PWM controller implementation which allows shared
// setting of frequency and individual setting of duty cycle for several PWM channels.
type PWMSetter interface {
	// Top usually represents maximum value the running PWM counter can take, either by hardware or set at configuration.
	Top() uint32
	// SetChannels sets the PWM counter limit for channels specified. channels are PWM peripheral based, not MCU host based.
	// Users need to check first what PWMSetter and channel combination corresponds to a pin before using the PWMSetter.
	//  - channels corresponds to values, so len(channels) must be equal to len(values).
	//  - The PWMSetter can potentially synchronize PWM signal output for a single SetChannels call depending on configuration.
	// Example:
	//
	//  quarter := 25*pwm.Top()/100 // 25% duty-cycle.
	//  full := pwm.Top()           // 100% duty cycle (always on)
	//  off := uint32(0)            // PWM low output (always off)
	//  pwm.SetChannels([]uint8{c1, c2, c3}, []uint32{quarter, full, off})
	SetChannels(channels []uint8, values []uint32) error
}
