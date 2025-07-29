package tmc2209

func SetMicrostepsPerStep(microsteps uint16) uint8 {
	exponent := uint8(0)
	microstepsShifted := microsteps >> 1

	for microstepsShifted > 0 {
		microstepsShifted = microstepsShifted >> 1
		exponent++
	}

	SetMicrostepsPerStepPowerOfTwo(exponent)
	return exponent
}

func SetMicrostepsPerStepPowerOfTwo(exponent uint8) {
	switch exponent {
	case 0:
		// Set MRES_001
	case 1:
		// Set MRES_002
	case 2:
		// Set MRES_004
	case 3:
		// Set MRES_008
	case 4:
		// Set MRES_016
	case 5:
		// Set MRES_032
	case 6:
		// Set MRES_064
	case 7:
		// Set MRES_128
	default:
		// Set MRES_256
	}
}
