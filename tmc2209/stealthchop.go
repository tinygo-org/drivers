package tmc2209

func EnableStealthChop() {
	// Set StealthChop enabled in the global config register
}

func DisableStealthChop() {
	// Set StealthChop disabled in the global config register
}

func EnableCoolStep(lowerThreshold, upperThreshold uint8) {
	// Enable CoolStep with specified thresholds
}

func DisableCoolStep() {
	// Disable CoolStep feature
}

func EnableAutomaticCurrentScaling() {
	// Enable Automatic Current Scaling
}

func DisableAutomaticCurrentScaling() {
	// Disable Automatic Current Scaling
}
